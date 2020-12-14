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
from bloom_filter import BloomFilter

""" Initialize ipfs client """
ipfs_client = ipfshttpclient.connect(settings.IPFS_URL)

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

            block_height_key = f"projectID:{request_info['projectId']}:blockHeight"
            max_block_height = await redis_conn.get(block_height_key)
            max_block_height = int(max_block_height)

            # Iterate through each dag block
            for block_cid, block_height in all_cids:
                block_cid = block_cid.decode('utf-8')
                block_height = int(block_height)
                block_dag = None
                payload_data = None
                
                """ Check if the DAG block is pinned """
                if block_height < max_block_height - settings.max_ipfs_blocks:
                    """ Get the data directly through the IPFS client """
                    block_dag = ipfs_client.dag.get(block_cid).as_json()
                    payload_data = ipfs_client.cat(block_dag['data']['cid'])
                    if isinstance(payload_data, bytes):
                        payload_data = payload_data.decode('utf-8')

                else:
                    """ Get the container for that block height """
                    containers_created_key = f"projectID:{request_info['projectId']}:containers"
                    target_containers = await redis_conn.zrangebyscore(
                        key=containers_created_key,
                        max=settings.container_height*2 + block_height + 1,
                        min=block_height - settings.container_height*2 - 1
                    )

                    """ Iterate through each containerId and then check if the block exists in that container """
                    bloom_object = None
                    container_data = None
                    for container_id in target_containers:
                        """ Get the data for the container """
                        container_id = container_id.decode('utf-8')
                        container_data_key = f"projectID:{request_info['projectId']}:containerData:{container_id}"
                        out = await redis_conn.hgetall(container_data_key)
                        container_data = {k.decode('utf-8'):v.decode('utf-8') for k,v in out.items()}
                        bloom_filter_settings = json.loads(container_data['bloomFilterSettings'])
                        retrieval_logger.debug(bloom_filter_settings)
                        bloom_object = BloomFilter(**bloom_filter_settings)
                        if block_cid in bloom_object:
                            break

                    retrieval_logger.debug("Found the matching container")
                    retrieval_logger.debug(container_data)

                    out = powgate_client.data.get(container_data['containerCid'], token=token).decode('utf-8')
                    container = json.loads(out)['container']
                    _block_index = block_height % settings.container_height
                    block_cid, block_dag = next(iter(container['dagChain'][_block_index].items()))
                    payload_data = container['payloads'][block_dag['data']['cid']]

                retrieval_logger.debug("Retrieved the DAG Block...")

                retrieval_data = {}


                if request_info['data'] == '2':
                    """ Save only the payload data """
                    payload_cid = block_dag['data']['cid']
                    # Get payload from filecoin
                    payload = payload_data

                    retrieval_data['cid'] = payload_cid
                    retrieval_data['type'] = 'COLD_FILECOIN'
                    retrieval_data['payload']  = payload
                    
                elif int(request_info['data']) in range(0,2): # Either 0 or 1
                    """ Save the entire DAG Block """
                    retrieval_data = {k:v for k,v in block_dag.items()}

                    if request_info['data'] == '1':
                        """ Save the payload data in the DAG block """
                        retrieval_data['data']['payload'] = payload_data

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
