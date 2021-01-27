from pygate_grpc.client import PowerGateClient
import logging
import aioredis
import sys
from config import settings

from random import choice
from string import ascii_letters
import requests
import time
import asyncio


""" Initialize logging """
test_logger = logging.getLogger(__name__)
test_logger.setLevel(logging.DEBUG)
stream_handler = logging.StreamHandler(sys.stdout)
formatter = logging.Formatter("%(asctime)-8s-%(msecs)s %(name)s-%(funcName)s: %(message)s")
stream_handler.setFormatter(formatter)
test_logger.addHandler(stream_handler)
test_logger.debug("Initialized logging...")

""" Redis Boilerplate """
redis_pool = None
async def redis_boilerplate():
    global redis_pool
    redis_pool = await aioredis.create_pool(
        address=(settings.REDIS.HOST, settings.REDIS.PORT)
    )


async def commit_to_audit_protocol():
    """ Push data to /commit_payload """
    projectId = 24
    random_string = ''.join([choice(ascii_letters) for i in range(1024)])
    data = {'payload':{'random_string':random_string}, 'projectId': projectId}
    response = requests.post('http://localhost:9000/commit_payload',json=data)
    test_logger.debug(response.text)


async def push_to_filecoin():
    """
        This function will push payloads directly to filecoin
    :return:
    """
    powgate_client = PowerGateClient("127.0.0.1:5002")
    user = powgate_client.admin.users.create()
    redis_conn_raw = await redis_pool.acquire()
    redis_conn = aioredis.Redis(redis_conn_raw)

    projectId = 24
    _ = await redis_conn.sadd("storedProjectIds", projectId)
    for i in range(500):
        test_logger.debug(f"Adding data {i} to filecoin..")
        random_string = ''.join([choice(ascii_letters) for i in range(1024)]).encode('utf-8')
        stage_res = powgate_client.data.stage_bytes(random_string, token=user.token)
        job = powgate_client.config.apply(stage_res.cid, override=True, token=user.token)

        key = f"testDataAt:{projectId}"
        _ = await redis_conn.zadd(
            key=key,
            score=i,
            member=stage_res.cid
        )

        _ = await redis_conn.set(f"testBlockHeight", i)
        time.sleep(1)

if __name__ == "__main__":
    asyncio.run(redis_boilerplate())
    for i in range(200):
        asyncio.run(commit_to_audit_protocol())
        test_logger.debug(f"{i}: Sleeping for 20 secs")
        time.sleep(20)
