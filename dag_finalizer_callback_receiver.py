from fastapi import FastAPI, Request, Response, Header
from utils.rabbitmq_utils import get_rabbitmq_connection, get_rabbitmq_channel, get_rabbitmq_core_exchange, get_rabbitmq_routing_key
from config import settings
from utils import dag_utils
from aio_pika import ExchangeType, DeliveryMode, Message
from functools import partial
from aio_pika.pool import Pool
from data_models import DAGFinalizerCallback
from multiprocessing import Manager
import asyncio
import aiohttp
import logging
import sys


app = FastAPI()

formatter = logging.Formatter(u"%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s")

stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setLevel(logging.DEBUG)
# not setting formatter here since logs are intercepted by loguru in the gunicorn launcher script and further adapted
# stdout_handler.setFormatter(formatter)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setLevel(logging.ERROR)
# stderr_handler.setFormatter(formatter)
rest_logger = logging.getLogger(__name__)
rest_logger.setLevel(logging.DEBUG)
rest_logger.addHandler(stdout_handler)
rest_logger.addHandler(stderr_handler)

REDIS_WRITER_CONN_CONF = {
    "host": settings.redis.host,
    "port": settings.redis.port,
    "password": settings.redis.password,
    "db": settings.redis.db
}

REDIS_READER_CONN_CONF = {
    "host": settings.redis_reader.host,
    "port": settings.redis_reader.port,
    "password": settings.redis_reader.password,
    "db": settings.redis_reader.db
}
multiprocess_manager = Manager()
project_specific_locks_multiprocessing_map = multiprocess_manager.dict()


@app.on_event('startup')
async def startup_boilerplate():
    app.rmq_connection_pool = Pool(get_rabbitmq_connection, max_size=5, loop=asyncio.get_running_loop())
    app.rmq_channel_pool = Pool(
        partial(get_rabbitmq_channel, app.rmq_connection_pool), max_size=20, loop=asyncio.get_running_loop()
    )


@app.post('/')
async def handle_dag_cb(
        request: Request,
        response: Response,
):
    # global project_specific_locks_multiprocessing_map
    event_data = await request.json()
    response_body = dict()
    response_status_code = 200
    if 'event_name' in event_data.keys():
        if event_data['event_name'] == 'RecordAppended':
            # rest_logger.debug(event_data)
            async with request.app.rmq_channel_pool.acquire() as channel:
                # to save a call to rabbitmq. we already initialize exchanges and queues beforehand
                # always ensure exchanges and queues are initialized as part of launch sequence,
                # not to be checked here
                exchange = await channel.get_exchange(
                    name=get_rabbitmq_core_exchange(),
                    ensure=False
                )
                event_data_obj = DAGFinalizerCallback.parse_obj(event_data)
                message = Message(
                    event_data_obj.json().encode('utf-8'),
                    delivery_mode=DeliveryMode.PERSISTENT,
                )
                await exchange.publish(
                    message=message,
                    routing_key=get_rabbitmq_routing_key('dag-processing')
                )
                rest_logger.debug(
                    'Published finalizer callback to rabbitmq %s',
                    event_data_obj
                )
    response.status_code = response_status_code
    return response_body
