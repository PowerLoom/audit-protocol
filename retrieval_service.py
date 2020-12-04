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

    """ Get all the pending requests """
    requests_list_key = f"pendingRetrievalRequests"
    all_requests = await redis_conn.smembers(requests_list_key)
    if all_requests:
        for requestId in all_requests:
            requestId = requestId.decode('utf-8')
            retrieval_logger.debug("Processing request: "+requestId)

            """ Get the required information about the requestId """
            key = f"retrievalRequestInfo:{requestId}"
            out  = await redis_conn.hgetall(key=key)
            request_info = {k.decode('utf-8'):i.decode('utf-8') for k,i in out.items()}
            retrieval_logger.debug(f"Retrieved information for request: {requestId}")
            retrieval_logger.debug(request_info)

            # Get the token for that projectId
            token_key = f"filecoinToken:{request_info['projectId']}"
            token = await redis_conn.get(token_key)
            token = token.decode('utf-8')

            """ 
                - For the project_id, using the from_height and to_height from the request_info, 
                get all the DAG block cids. 
            """
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
                block_height = int(block_height)

                # Get the block's staged_cid
                project_id = request_info['projectId']
                staged_block_key = f"blockFilecoinStorage:{project_id}:{block_height}"
                block_staged_cid = await redis_conn.hget(
                    key=staged_block_key,
                    field="blockStageCid"
                )
                block_staged_cid = block_staged_cid.decode('utf-8')
                retrieval_logger.debug(f"Getting data for DAG Block: {block_staged_cid}")

                block_dag = powgate_client.data.get(block_staged_cid, token=token)
                block_dag = block_dag.decode('utf-8')
                block_dag = json.loads(block_dag)
                retrieval_data = {}


                if request_info['data'] == '2':
                    """ Save only the payload data """
                    payload_cid = block_dag['Data']['Cid']
                    # Get payload from filecoin
                    payload = powgate_client.data.get(payload_cid,token=token)
                    payload = payload.decode('utf-8')

                    retrieval_data['Cid'] = payload_cid
                    retrieval_data['Type'] = 'COLD_FILECOIN'
                    retrieval_data['payload']  = payload
                    
                elif int(request_info['data']) in range(0,2): # Either 0 or 1
                    """ Save the entire DAG Block """
                    retrieval_data = {k:v for k,v in block_dag.items()}

                    if request_info['data'] == '1':
                        """ Save the payload data in the DAG block """
                        payload_cid = retrieval_data['Data']['Cid']
                        # Get payload from filecoin
                        payload = powgate_client.data.get(payload_cid,token=token)
                        payload = payload.decode('utf-8')
                        retrieval_data['Data']['payload'] = payload

                else:
                    raise ValueError(f"Invalid value for data field in request_info: {request_info['data']}")

                # Save the required retrieval data 
                block_file_path = f'static/{block_cid}'
                with open(block_file_path, 'w') as f:
                    f.write(json.dumps(retrieval_data))

                retrieval_files_key = f"retrievalRequestFiles:{requestId}"
                _ = await redis_conn.zadd(
                    key=retrieval_files_key,
                    member=block_file_path,
                    score=int(block_height)
                )
                retrieval_logger.debug(f"Block {block_cid} is saved")
            
            # Once the request is complete, then delete the request id from pending retrieval requests set
            _ = await redis_conn.srem(requests_list_key, requestId)
            retrieval_logger.debug(f"Request: {requestId} has been removed from pending retrieveal requests")
    else:
        retrieval_logger.debug(f"No pending requests found....")
    redis_pool.release(redis_conn_raw)


if __name__ == "__main__":
    asyncio.run(redis_boilerplate())
    while True:
        retrieval_logger.debug("Looking for pending retrieval requests....")
        asyncio.run(retrieve_files())
        retrieval_logger.debug("Sleeping for 20 secs")
        time.sleep(20)
