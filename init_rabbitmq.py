from config import settings
import pika


def create_rabbitmq_conn():
    c = pika.BlockingConnection(pika.ConnectionParameters(
        host=settings.rabbitmq.host,
        port=settings.rabbitmq.port,
        virtual_host='/',
        credentials=pika.PlainCredentials(
            username=settings.rabbitmq.user,
            password=settings.rabbitmq.password
        ),
        heartbeat=30
    ))
    return c


def init_queue(ch: pika.adapters.blocking_connection.BlockingChannel, queue_name, routing_key, exchange_name):
    ch.queue_declare(queue=queue_name, durable=True)
    ch.queue_bind(exchange=exchange_name, queue=queue_name, routing_key=routing_key)
    print(
        'Initialized rabbitmq setup | Queue: %s | Exchange: %s | Routing Key: %s',
        queue_name,
        exchange_name,
        routing_key
    )


def init_exchanges_queues():
    c = create_rabbitmq_conn()
    ch: pika.adapters.blocking_connection.BlockingChannel = c.channel()
    exchange_name = settings.rabbitmq.setup['core']['exchange']
    ch.exchange_declare(exchange=exchange_name, exchange_type='direct', durable=True)
    print('Initialized rabbitmq Direct exchange: %s', exchange_name)

    # queue name, routing key pairs
    to_be_inited = [
        ('audit-protocol-commit-payloads', 'commit-payloads'),
        ('audit-protocol-diff-requests', 'diff-calculation')
    ]
    for queue_name, routing_key in to_be_inited:
        # add namespace?
        init_queue(ch, queue_name, routing_key, exchange_name)


if __name__ == '__main__':
    init_exchanges_queues()
