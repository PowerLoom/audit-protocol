from pygate_grpc.client import PowerGateClient
import logging
import sys
import aioredis
from config import settings
import ipfshttpclient
import asyncio
import os
import json
import time

""" Initialize ipfs client """
ipfs_client = ipfshttpclient.connect(settings.IPFS_URL)

""" Inititalize the logger """
deal_watcher_logger = logging.getLogger(__name__)
deal_watcher_logger.setLevel(logging.DEBUG)
formatter = logging.Formatter("%(levelname)-8s %(name)-4s %(asctime)s %(msecs)d %(module)s-%(funcName)s: %(message)s")
stream_handler = logging.StreamHandler(sys.stdout)
stream_handler.setFormatter(formatter)
deal_watcher_logger.addHandler(stream_handler)
deal_watcher_logger.debug("Initialized logger")

""" aioredis boilerplate """
redis_pool = None
async def redis_boilerplate():
    global redis_pool
    redis_pool = await aioredis.create_pool(
        address=(settings.REDIS.HOST, settings.REDIS.PORT),
        maxsize=5
    )

""" Create the deal watcher function """
async def deal_watcher():
    """ This function will monitor all the jobs and update thier status """
