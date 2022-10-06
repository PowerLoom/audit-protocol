from config import settings
from typing import Union
import aio_pika
import pika
import pika.exceptions
import logging
import sys
import functools
import uuid

formatter = logging.Formatter('%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s')
logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setLevel(logging.DEBUG)
stdout_handler.setFormatter(formatter)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setLevel(logging.ERROR)
stderr_handler.setFormatter(formatter)

logger.addHandler(stdout_handler)
logger.addHandler(stderr_handler)

def get_rabbitmq_routing_key(queueType: str):
    return settings.rabbitmq.setup.queues[queueType].routing_key_prefix + settings.instance_id

def get_rabbitmq_queue_name(queueType: str):
    return settings.rabbitmq.setup.queues[queueType].queue_name_prefix + settings.instance_id

def get_rabbitmq_core_exchange():
    return settings.rabbitmq.setup.core.exchange + settings.instance_id



async def get_rabbitmq_connection():
    return await aio_pika.connect_robust(
        host=settings.rabbitmq.host,
        port=settings.rabbitmq.port,
        virtual_host='/',
        login=settings.rabbitmq.user,
        password=settings.rabbitmq.password
    )


async def get_rabbitmq_channel(connection_pool) -> aio_pika.Channel:
    async with connection_pool.acquire() as connection:
        return await connection.channel()


# Adapted from:
# https://github.com/pika/pika/blob/12dcdf15d0932c388790e0fa990810bfd21b1a32/examples/asynchronous_publisher_example.py
# https://github.com/pika/pika/blob/12dcdf15d0932c388790e0fa990810bfd21b1a32/examples/asynchronous_consumer_example.py
class RabbitmqSelectLoopInteractor(object):
    """This is an example publisher/consumer that will handle unexpected interactions
    with RabbitMQ such as channel and connection closures.

    If RabbitMQ closes the connection, it will reopen it. You should
    look at the output, as there are limited reasons why the connection may
    be closed, which usually are tied to permission related issues or
    socket timeouts.

    It uses delivery confirmations and illustrates one way to keep track of
    messages that have been sent and if they've been confirmed by RabbitMQ.

    """
    # interval at which the select loop polls to push out new notifications
    PUBLISH_INTERVAL = 0.1

    def __init__(self, consume_queue_name=None, consume_callback=None, consumer_worker_name=''):
        """Setup the example publisher object, passing in the URL we will use
        to connect to RabbitMQ.
        """
        self._connection = None
        self._channel: Union[None, pika.channel.Channel] = None
        self.should_reconnect = False
        self._consumer_worker_name = consumer_worker_name
        self.was_consuming = False
        self._consumer_tag = None
        self.queued_messages = dict()
        self._deliveries = None
        self._acked = None
        self._nacked = None
        self._message_number = None
        self._closing = False
        self._stopping = False
        self._exchange = None
        self._routing_key = None
        self._consuming = False
        self._consume_queue = consume_queue_name
        self._consume_callback = consume_callback
        # In production, experiment with higher prefetch values
        # for higher consumer throughput
        self._prefetch_count = 1

    def connect(self):
        """This method connects to RabbitMQ, returning the connection handle.
        When the connection is established, the on_connection_open method
        will be invoked by pika.

        :rtype: pika.SelectConnection

        """
        logger.info(
            '%s: RabbitMQ select loop interactor: Creating RabbitMQ select ioloop connection to %s',
            self._consumer_worker_name, (settings.rabbitmq.host, settings.rabbitmq.port)
        )
        return pika.SelectConnection(
            parameters=pika.ConnectionParameters(
                host=settings.rabbitmq.host,
                port=settings.rabbitmq.port,
                virtual_host='/',
                credentials=pika.PlainCredentials(
                    username=settings.rabbitmq.user,
                    password=settings.rabbitmq.password
                ),
                heartbeat=30
            ),
            on_open_callback=self.on_connection_open,
            on_open_error_callback=self.on_connection_open_error,
            on_close_callback=self.on_connection_closed)

    def on_connection_open(self, _unused_connection):
        """This method is called by pika once the connection to RabbitMQ has
        been established. It passes the handle to the connection object in
        case we need it, but in this case, we'll just mark it unused.

        :param pika.SelectConnection _unused_connection: The connection

        """
        logger.info('%s: RabbitMQ select loop interactor: Connection opened', self._consumer_worker_name)
        self.open_channel()

    def on_connection_open_error(self, _unused_connection, err):
        """This method is called by pika if the connection to RabbitMQ
        can't be established.

        :param pika.SelectConnection _unused_connection: The connection
        :param Exception err: The error

        """
        logger.error(
            '%s: RabbitMQ select loop interactor: Connection open failed, reopening in 5 seconds: %s',
            self._consumer_worker_name, err
        )
        self._connection.ioloop.call_later(5, self._connection.ioloop.stop)

    def on_connection_closed(self, _unused_connection, reason):
        """This method is invoked by pika when the connection to RabbitMQ is
        closed unexpectedly. Since it is unexpected, we will reconnect to
        RabbitMQ if it disconnects.

        :param pika.connection.Connection connection: The closed connection obj
        :param Exception reason: exception representing reason for loss of
            connection.

        """
        self._channel = None
        if self._stopping:
            self._connection.ioloop.stop()
        else:
            if ('200' and 'Normal shutdown' in reason.__repr__()) or ('SelfExitException' in reason.__repr__()):
                logger.warning(
                    '%s: RabbitMQ select loop interactor: Connection closed: %s',
                    self._consumer_worker_name, reason
                )
            else:
                logger.warning(
                    '%s: RabbitMQ select loop interactor: Connection closed, reopening in 5 seconds: %s',
                    self._consumer_worker_name, reason
                )
                self._connection.ioloop.call_later(5, self._connection.ioloop.stop)

    def open_channel(self):
        """This method will open a new channel with RabbitMQ by issuing the
        Channel.Open RPC command. When RabbitMQ confirms the channel is open
        by sending the Channel.OpenOK RPC reply, the on_channel_open method
        will be invoked.

        """
        logger.info(
            '%s: RabbitMQ select loop interactor: Creating a new channel', self._consumer_worker_name
        )
        self._connection.channel(on_open_callback=self.on_channel_open)

    def on_channel_open(self, channel):
        """This method is invoked by pika when the channel has been opened.
        The channel object is passed in so we can make use of it.

        Since the channel is now open, we'll declare the exchange to use.

        :param pika.channel.Channel channel: The channel object

        """
        logger.info('%s: RabbitMQ select loop interactor: Channel opened', self._consumer_worker_name)
        self._channel = channel
        self.add_on_channel_close_callback()
        # self.setup_exchange(self.EXCHANGE)
        try:
            self.start_publishing()
            if self._consume_queue and self._consume_callback:
                self.start_consuming(
                    queue_name=self._consume_queue,
                    consume_cb=self._consume_callback,
                    auto_ack=False
                )
        except Exception as err:
            logger.error(
                "%s: RabbitMQ select loop interactor: Failed on_channel_open hook with error_msg: %s",
                self._consumer_worker_name, str(err), exc_info=True
            )
            # must be raised back to caller to be handled there
            raise err

    def add_on_channel_close_callback(self):
        """This method tells pika to call the on_channel_closed method if
        RabbitMQ unexpectedly closes the channel.

        """
        logger.info('%s: RabbitMQ select loop interactor: Adding channel close callback', self._consumer_worker_name)
        self._channel.add_on_close_callback(self.on_channel_closed)

    def on_channel_closed(self, channel, reason):
        """Invoked by pika when RabbitMQ unexpectedly closes the channel.
        Channels are usually closed if you attempt to do something that
        violates the protocol, such as re-declare an exchange or queue with
        different parameters. In this case, we'll close the connection
        to shutdown the object.

        :param pika.channel.Channel channel: The closed channel
        :param Exception reason: why the channel was closed

        """
        logger.warning(
            '%s: RabbitMQ select loop interactor: Channel %i was closed: %s',
            self._consumer_worker_name, channel, reason
        )
        self._channel = None
        try:
            self._connection.close()
        except Exception as e:
            if isinstance(e, pika.exceptions.ConnectionWrongStateError) and\
                    'Illegal close' in e.__repr__() and 'connection state=CLOSED' in e.__repr__():
                logger.error(
                    '%s: RabbitMQ select loop interactor: Tried closing connection '
                    'that was already closed on channel close callback. Exception: %s. Will close ioloop now.',
                    self._consumer_worker_name,
                    e
                )
            else:
                logger.error(
                    '%s: RabbitMQ select loop interactor: Exception closing connection '
                    'on channel close callback: %s. Will close ioloop now.',
                    self._consumer_worker_name, e, exc_info=True
                )
            self._connection.ioloop.stop()

    def start_publishing(self):
        """This method will enable delivery confirmations and schedule the
        first message to be sent to RabbitMQ

        """
        logger.info(
            '%s: RabbitMQ select loop interactor: Issuing consumer related RPC commands', self._consumer_worker_name
        )
        self.enable_delivery_confirmations()
        self.schedule_next_message()

    def enable_delivery_confirmations(self):
        """Send the Confirm.Select RPC method to RabbitMQ to enable delivery
        confirmations on the channel. The only way to turn this off is to close
        the channel and create a new one.

        When the message is confirmed from RabbitMQ, the
        on_delivery_confirmation method will be invoked passing in a Basic.Ack
        or Basic.Nack method from RabbitMQ that will indicate which messages it
        is confirming or rejecting.

        """
        logger.info(
            '%s: RabbitMQ select loop interactor: Issuing Confirm.Select RPC command', self._consumer_worker_name
        )
        self._channel.confirm_delivery(self.on_delivery_confirmation)

    def on_delivery_confirmation(self, method_frame):
        """Invoked by pika when RabbitMQ responds to a Basic.Publish RPC
        command, passing in either a Basic.Ack or Basic.Nack frame with
        the delivery tag of the message that was published. The delivery tag
        is an integer counter indicating the message number that was sent
        on the channel via Basic.Publish. Here we're just doing house keeping
        to keep track of stats and remove message numbers that we expect
        a delivery confirmation of from the list used to keep track of messages
        that are pending confirmation.

        :param pika.frame.Method method_frame: Basic.Ack or Basic.Nack frame

        """
        confirmation_type = method_frame.method.NAME.split('.')[1].lower()
        logger.info(
            '%s: RabbitMQ select loop interactor: Received %s for delivery tag: %i',
            self._consumer_worker_name, confirmation_type, method_frame.method.delivery_tag
        )
        if confirmation_type == 'ack':
            self._acked += 1
        elif confirmation_type == 'nack':
            self._nacked += 1
        self._deliveries.remove(method_frame.method.delivery_tag)
        logger.info(
            '%s: RabbitMQ select loop interactor: Published %i messages, %i have yet to be confirmed, '
            '%i were acked and %i were nacked', self._consumer_worker_name, self._message_number,
            len(self._deliveries), self._acked, self._nacked)

    def enqueue_msg_delivery(self, exchange, routing_key, msg_body):
        # NOTE: if used in a multi threaded/multi processing context, this will introduce a race condition given that
        #       publish_message() may try to read the queued messages map while it is being concurrently updated by
        #       another process/thread
        self.queued_messages[str(uuid.uuid4())] = [msg_body, exchange, routing_key]

    def schedule_next_message(self):
        """If we are not closing our connection to RabbitMQ, schedule another
        message to be delivered in PUBLISH_INTERVAL seconds.

        """
        # logger.info('Scheduling next message for %0.1f seconds',
        #             self.PUBLISH_INTERVAL)
        self._connection.ioloop.call_later(self.PUBLISH_INTERVAL,
                                           self.publish_message)

    def publish_message(self):
        """If the class is not stopping, publish a message to RabbitMQ,
        appending a list of deliveries with the message number that was sent.
        This list will be used to check for delivery confirmations in the
        on_delivery_confirmations method.

        Once the message has been sent, schedule another message to be sent.
        The main reason I put scheduling in was just so you can get a good idea
        of how the process is flowing by slowing down and speeding up the
        delivery intervals by changing the PUBLISH_INTERVAL constant in the
        class.

        """
        if self._channel is None or not self._channel.is_open:
            return
        # check for queued messages
        pushed_outputs = list()
        # logger.debug('queued msgs: %s', self.queued_messages)
        try:
            for unique_id in self.queued_messages:
                msg_info = self.queued_messages[unique_id]
                msg, exchange, routing_key = msg_info
                logger.debug(
                    '%s: RabbitMQ select loop interactor: Got queued message body to send to exchange %s '
                    'via routing key %s: %s',
                    self._consumer_worker_name, exchange, routing_key, msg
                )
                properties = pika.BasicProperties(
                    delivery_mode=2,
                    content_type='text/plain',
                    content_encoding='utf-8'
                )
                self._channel.basic_publish(
                    exchange=exchange,
                    routing_key=routing_key,
                    body=msg.encode('utf-8'),
                    properties=properties
                )
                self._message_number += 1
                self._deliveries.append(self._message_number)
                logger.info(
                    '%s: RabbitMQ select loop interactor: Published message # %i to exchange %s via routing key %s: %s',
                    self._consumer_worker_name, self._message_number, exchange, routing_key, msg
                )
                pushed_outputs.append(unique_id)
            self.queued_messages = {k: self.queued_messages[k] for k in self.queued_messages if k not in pushed_outputs}
        except Exception as err:
            logger.error(
                "%s: RabbitMQ select loop interactor: Error while publishing message to rabbitmq error_msg: "
                "%s, exchange: %s, routing_key: %s, msg: %s",
                self._consumer_worker_name, str(err), exchange, routing_key, msg, exc_info=True
            )
        finally:
            self.schedule_next_message()

    def run(self):
        """Run the example code by connecting and then starting the IOLoop.

        """
        attempt = 1
        while not self._stopping:
            logger.info(
                '%s: RabbitMQ interactor select ioloop runner starting | Attempt # %s',
                self._consumer_worker_name, attempt
            )
            self._connection = None
            self._deliveries = []
            self._acked = 0
            self._nacked = 0
            self._message_number = 0
            self._connection = self.connect()
            self._connection.ioloop.start()
            logger.info(
                '%s: RabbitMQ interactor select ioloop exited in runner loop | Attempt # %s',
                self._consumer_worker_name, attempt
            )
            attempt += 1
        logger.info('%s: RabbitMQ interactor select ioloop stopped', self._consumer_worker_name)

    def close_channel(self):
        """Invoke this command to close the channel with RabbitMQ by sending
        the Channel.Close RPC command.

        """
        if self._channel is not None:
            logger.info('%s: RabbitMQ select loop interactor: Closing the channel', self._consumer_worker_name)
            self._channel.close()

    def close_connection(self):
        """This method closes the connection to RabbitMQ."""
        if self._connection is not None:
            logger.info('%s: RabbitMQ select loop interactor: Closing connection', self._consumer_worker_name)
            self._connection.close()

    def start_consuming(self, queue_name, consume_cb, auto_ack=False):
        """This method sets up the consumer by first calling
        add_on_cancel_callback so that the object is notified if RabbitMQ
        cancels the consumer. It then issues the Basic.Consume RPC command
        which returns the consumer tag that is used to uniquely identify the
        consumer with RabbitMQ. We keep the value to use it when we want to
        cancel consuming. The on_message method is passed in as a callback pika
        will invoke when a message is fully received.
        """
        logger.info(
            '%s: RabbitMQ select loop interactor: Issuing consumer related RPC commands',
            self._consumer_worker_name
        )
        self.add_on_cancel_callback()
        self._consumer_tag = self._channel.basic_consume(
            queue=queue_name,
            on_message_callback=consume_cb,
            auto_ack=auto_ack
        )
        self.was_consuming = True
        self._consuming = True
        # self._channel.start_consuming()

    def add_on_cancel_callback(self):
        """Add a callback that will be invoked if RabbitMQ cancels the consumer
        for some reason. If RabbitMQ does cancel the consumer,
        on_consumer_cancelled will be invoked by pika.
        """
        logger.info(
            '%s: RabbitMQ select loop interactor: Adding consumer cancellation callback',
            self._consumer_worker_name
        )
        self._channel.add_on_cancel_callback(self.on_consumer_cancelled)

    def on_consumer_cancelled(self, method_frame):
        """Invoked by pika when RabbitMQ sends a Basic.Cancel for a consumer
        receiving messages.
        :param pika.frame.Method method_frame: The Basic.Cancel frame
        """
        logger.info(
            '%s: Consumer was cancelled remotely, shutting down: %r',
            self._consumer_worker_name, method_frame
        )
        if self._channel:
            self._channel.close()

    def stop_consuming(self):
        """Tell RabbitMQ that you would like to stop consuming by sending the
        Basic.Cancel RPC command.
        """
        if self._channel:
            logger.info(
                '%s: RabbitMQ select loop interactor: Sending a Basic.Cancel RPC command to RabbitMQ',
                self._consumer_worker_name
            )
            cb = functools.partial(
                self.on_cancelok, userdata=self._consumer_tag)
            self._channel.basic_cancel(self._consumer_tag, cb)

    def on_cancelok(self, _unused_frame, userdata):
        """This method is invoked by pika when RabbitMQ acknowledges the
        cancellation of a consumer. At this point we will close the channel.
        This will invoke the on_channel_closed method once the channel has been
        closed, which will in-turn close the connection.
        :param pika.frame.Method _unused_frame: The Basic.CancelOk frame
        :param str|unicode userdata: Extra user data (consumer tag)
        """
        self._consuming = False
        logger.info(
            '%s: RabbitMQ select loop interactor: RabbitMQ acknowledged the cancellation of the consumer: %s',
            self._consumer_worker_name, userdata
        )
        self.close_channel()

    def stop(self):
        """
        Cleanly shutdown the connection to RabbitMQ by stopping the consumer with RabbitMQ.
        """
        if not self._closing:
            self._closing = True
            self._stopping = True
            logger.info('%s: RabbitMQ select loop interactor: Stopping', self._consumer_worker_name)
            if self._consuming:
                self.stop_consuming()
                # self._connection.ioloop.start()
            else:
                self._connection.ioloop.stop()
            logger.info('%s: RabbitMQ select loop interactor: Stopped', self._consumer_worker_name)

