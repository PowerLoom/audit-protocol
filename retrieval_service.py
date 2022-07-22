import logging
import sys
from config import settings
import asyncio
import json
from bloom_filter import BloomFilter
import coloredlogs
from utils import helper_functions
from utils import redis_keys
from utils.ipfs_async import client as ipfs_client
from utils.diffmap_utils import preprocess_dag
from utils.backup_utils import get_backup_data
from utils import retrieval_utils


""" Inititalize the logger """
retrieval_logger = logging.getLogger(__name__)
retrieval_logger.setLevel(logging.DEBUG)
formatter = logging.Formatter("%(levelname)-8s %(name)-4s %(asctime)s %(msecs)d %(module)s-%(funcName)s: %(message)s")
stream_handler = logging.StreamHandler(sys.stdout)
stream_handler.setFormatter(formatter)
retrieval_logger.addHandler(stream_handler)
retrieval_logger.debug("Initialized logger")
coloredlogs.install(level="DEBUG", logger=retrieval_logger, stream=sys.stdout)


async def retrieve_files(reader_redis_conn=None, writer_redis_conn=None):

    """ Get all the pending requests """
    requests_list_key = f"pendingRetrievalRequests"
    all_requests = await reader_redis_conn.smembers(requests_list_key)
    if all_requests:
        for requestId in all_requests:
            requestId = requestId.decode('utf-8')
            retrieval_logger.debug("Processing request: ")
            retrieval_logger.debug(requestId)

            """ Get the required information about the requestId """
            key = redis_keys.get_retrieval_request_info_key(requestId)
            out = await reader_redis_conn.hgetall(key=key)
            request_info = {k.decode('utf-8'): i.decode('utf-8') for k, i in out.items()}
            retrieval_logger.debug(f"Retrieved information for request: {requestId}")
            retrieval_logger.debug(request_info)

            """ Check if any of the files in this request are not pushed to filecoin yet """
            # Get the height of last pruned cid
            last_pruned_height = await helper_functions.get_last_pruned_height(
                project_id=request_info['projectId'],
                reader_redis_conn=reader_redis_conn
            )

            retrieval_logger.debug("Last Pruned Height:")
            retrieval_logger.debug(last_pruned_height)

            """ 
                - For the project_id, using the from_height and to_height from the request_info, 
                get all the DAG block cids. 
            """
            block_cids_key = redis_keys.get_dag_cids_key(project_id=request_info['projectId'])
            all_cids = await reader_redis_conn.zrangebyscore(
                key=block_cids_key,
                max=int(request_info['to_height']),
                min=int(request_info['from_height']),
                withscores=True
            )

            max_block_height = await helper_functions.get_block_height(
                project_id=request_info['projectId'],
                reader_redis_conn=reader_redis_conn
            )

            # Create a dictionary to save data into a
            span_dags = {}

            # Iterate through each dag block
            for block_cid, block_height in all_cids:
                block_cid = block_cid.decode('utf-8')
                block_height = int(block_height)
                # retrieval_logger.debug("Fetching block at height:")
                # retrieval_logger.debug(block_cid)
                # retrieval_logger.debug(block_height)

                """ Check if the DAG block is pinned """
                if (block_height > (max_block_height - settings.max_ipfs_blocks)) or (
                        last_pruned_height < int(request_info['to_height'])):
                    """ Get the data directly through the IPFS client """
                    _block_dag = await ipfs_client.dag.get(block_cid)
                    block_dag = _block_dag.as_json()
                    block_dag = preprocess_dag(block_dag)

                    payload_data = await ipfs_client.cat(block_dag['data']['cid'])
                    if isinstance(payload_data, bytes):
                        payload_data = payload_data.decode('utf-8')

                else:
                    """ Get the container for that block height """
                    containers_created_key = redis_keys.get_containers_created_key(request_info['projectId'])
                    target_containers = await reader_redis_conn.zrangebyscore(
                        key=containers_created_key,
                        max=settings.container_height * 2 + block_height + 1,
                        min=block_height - settings.container_height * 2 - 1
                    )

                    """ Iterate through each containerId and then check if the block exists in that container """
                    bloom_object = None
                    container_data = {}
                    container_id = ""
                    for container_id in target_containers:
                        """ Get the data for the container """
                        container_id = container_id.decode('utf-8')
                        container_data_key = redis_keys.get_container_data_key(container_id)
                        out = await reader_redis_conn.hgetall(container_data_key)
                        container_data = {k.decode('utf-8'): v.decode('utf-8') for k, v in out.items()}
                        bloom_filter_settings = json.loads(container_data['bloomFilterSettings'])
                        retrieval_logger.debug(bloom_filter_settings)
                        bloom_object = BloomFilter(**bloom_filter_settings)
                        if block_cid in bloom_object:
                            break

                    retrieval_logger.debug("Found the matching container")
                    retrieval_logger.debug(container_data)

                    container = await get_backup_data(container_data=container_data, container_id=container_id)
                    _target_blocks = container['dagChain']
                    for _tt_dag_block in _target_blocks:
                        _block_dag = list(_tt_dag_block.values()).pop()
                        retrieval_logger.debug(_block_dag)
                        if _block_dag['height'] == block_height:
                            block_dag = _block_dag
                            break
                    # _block_index = settings.container_height - (block_height-1) % settings.container_height
                    # current_block_cid, block_dag = next(iter(container['dagChain'][_block_index].items()))
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
                    retrieval_data['payload'] = payload

                elif int(request_info['data']) in range(0, 2):  # Either 0 or 1
                    """ Save the entire DAG Block """
                    retrieval_data = {k: v for k, v in block_dag.items()}

                    if request_info['data'] == '1':
                        """ Save the payload data in the DAG block """
                        retrieval_data['data']['payload'] = payload_data

                else:
                    raise ValueError(f"Invalid value for data field in request_info: {request_info['data']}")

                # Save the required retrieval data 
                block_file_path = f'static/{block_cid}'
                with open(block_file_path, 'w') as f:
                    f.write(json.dumps(retrieval_data))

                span_dags[block_height] = retrieval_data

                retrieval_files_key = redis_keys.get_retrieval_request_files_key(requestId)
                _ = await writer_redis_conn.zadd(
                    name=retrieval_files_key,
                    mapping={block_file_path: int(block_height)}
                )
                retrieval_logger.debug(f"Block {block_cid} is saved")

            # Once the request is complete, then delete the request id from pending retrieval requests set
            _ = await writer_redis_conn.srem(requests_list_key, requestId)
            retrieval_logger.debug(f"Request: {requestId} has been removed from pending retrieveal requests")
            _ = await writer_redis_conn.delete(key)
            retrieval_logger.debug("Request Data has been deleted from redis")

            # Create a span and cache it
            retrieval_logger.debug("Saving span data")
            out = await retrieval_utils.save_span(
                from_height=int(request_info['from_height']),
                to_height=int(request_info['to_height']),
                project_id=request_info['projectId'],
                dag_blocks=span_dags,
                writer_redis_conn=writer_redis_conn
            )

    else:
        retrieval_logger.debug(f"No pending requests found....")


def verifier_crash_cb(fut: asyncio.Future):
    try:
        exc = fut.exception()
    except asyncio.CancelledError:
        retrieval_logger.error('Respawning retrieval task...')
        t = asyncio.ensure_future(periodic_retrieval())
        t.add_done_callback(verifier_crash_cb)
    except Exception as e:
        retrieval_logger.error('retrieval task crashed')
        retrieval_logger.error(e, exc_info=True)

        
async def periodic_retrieval():
    while True:
        await asyncio.gather(
            retrieve_files(),
            asyncio.sleep(settings.retrieval_service_interval)
        )
        

if __name__ == "__main__":
    f = asyncio.ensure_future(periodic_retrieval())
    f.add_done_callback(verifier_crash_cb)
    try:
        asyncio.get_event_loop().run_until_complete(asyncio.gather(f))
    except:
        asyncio.get_event_loop().stop()