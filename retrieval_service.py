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
    powgate_client = PowerGateClient("127.0.0.1:5002", False)

    redis_conn_raw = await redis_pool.acquire()
    redis_conn = aioredis.Redis(redis_conn_raw)

    # First get all the requests that are put in the set
    requests_list_key = f"pendingRetrievalRequests"
    all_requests = await redis_conn.smembers(requests_list_key)
    if all_requests:
        for request in all_requests:
            request = request.decode('utf-8')
            retrieval_logger.debug("Processing request: "+request)

            # Get the request data
            key = f"retrievalRequestInfo:{request}"
            out  = await redis_conn.hgetall(key=key)
            request_info = {k.decode('utf-8'):i.decode('utf-8') for k,i in out.items()}

            # Get the token for that projectId
            token_key = f"filecoinToken:{request['projectId']}"
            token = await redis_conn.get(token_key)
            token = token.deocde('utf-8')

            # For that project_id, using the from_ and to_, get all the cids
            block_cids_key = f"projectID:{request_info['projectId']}:Cids"
            all_cids = await redis_conn.zrangebyscore(
                key=block_cids_key,
                max=int(request_info['to_height']),
                min=int(request_info['from_height']),
                withscores=True
            )

            # Iterate through each dag block
            for block_cid, block_height in all_cids:
                block_cid = block_cid.decode('utf-8')
                retrieval_logger.debug(f"Getting data for: {block_cid}")

                # Get the dag block from filecoin
                data = powgate_client.data.get(block_cid, token=token)
                data = data.decode('utf-8')
                data = json.loads(data)

                if request_info['data'] == '1':
                    payload_cid = data['Data']['Cid']
                    payload = powgate_client.data.get(payload_cid,token=token)
                    payload = payload.decode('utf-8')
                    data['Data']['payload'] = payload

                block_file_path = f'static/{block_cid}'
                with open(block_file_path, 'w') as f:
                    f.write(json.dumps(data))

                retrieval_files_key = f"retrievalRequestFiles:{request}"
                _ = await redis_conn.zadd(
                    key=retrieval_files_key,
                    member=block_file_path,
                    score=int(block_height)
                )
                retrieval_logger.debug(f"Block {block_cid} is saved")
            
            # Once the request is complete, the delete the request id from redis
            _ = await redis_conn.srem(requests_list_key, request)
            retrieval_logger.debug(f"Request: {request} has been removed from pending retrieveal requests")


if __name__ == "__main__":
    asyncio.run(redis_boilerplate())
    while True:
        retrieval_logger.debug("Retrieveing files")
        asyncio.run(retrieve_files())
        retrieval_logger.debug("Sleeping for 20 secs")
        time.sleep(20)