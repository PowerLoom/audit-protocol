from config import settings
import pika


def create_rabbitmq_conn():
    c = pika.BlockingConnection(pika.ConnectionParameters(
        host=settings.RABBITMQ.HOST,
        port=settings.RABBITMQ.PORT,
        virtual_host='/',
        credentials=pika.PlainCredentials(
            username=settings.RABBITMQ.USER,
            password=settings.RABBITMQ.PASSWORD
        ),
        heartbeat=30
    ))
    return c


def init_queue(ch: pika.adapters.blocking_connection.BlockingChannel, queue_name, routing_key, exchange_name):
    ch.queue_declare(queue_name)
    ch.queue_bind(exchange=exchange_name, queue=queue_name, routing_key=routing_key)
    print(
        'Initialized RabbitMQ setup | Queue: %s | Exchange: %s | Routing Key: %s',
        queue_name,
        exchange_name,
        routing_key
    )


def init_exchanges_queues():
    c = create_rabbitmq_conn()
    ch: pika.adapters.blocking_connection.BlockingChannel = c.channel()
    exchange_name = settings.rabbitmq.setup['core']['exchange']
    ch.exchange_declare(exchange=exchange_name, exchange_type='direct', durable=True)
    print('Initialized RabbitMQ Direct exchange: %s', exchange_name)

    # queue name, routing key pairs
    to_be_inited = [
        ('audit-protocol-diff-requests', 'diff-calculation')
    ]
    for queue_name, routing_key in to_be_inited:
        # add namespace?
        init_queue(ch, queue_name, routing_key, exchange_name)


if __name__ == '__main__':
    init_exchanges_queues()
