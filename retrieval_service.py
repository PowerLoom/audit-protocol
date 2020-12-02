import logging
import sys
import aioredis
from config import settings
import ipfshttpclient
import asyncio

""" Initialize ipfs client """
ipfs_client = ipfshttpclient.connect()

""" Inititalize the logger """
retrieval_logger = logging.getLogger(__name__)
retrieval_logger.setLevel(logging.DEBUG)
formatter = logging.Formatter("%(levelname)-8s %(name)-4s %(asctime)s %(msecs)d %(module)s-%(funcName)s: %(message)s")
stream_handler = logging.StreamHandler(sys.stdout)
stream_handler.setFormatter(formatter)
retrieval_logger.addHandler(stream_handler)
retrieval_logger.debug("Initialized logger")

""" aioredis boilerplate """
redis_pool = None
async def redis_boilerplate():
    global redis_pool
    redis_pool = await aioredis.create_pool(
        address=(settings.REDIS.HOST, settings.REDIS.PORT),
        maxsize=5
    )

async def retrieve_files():
    powgate_client = PowerGateClient("127.0.0.1:5002")
