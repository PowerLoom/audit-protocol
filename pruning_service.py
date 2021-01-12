import ipfshttpclient
import logging
import sys
import aioredis
from config import settings
import time
import asyncio
from bloom_filter import BloomFilter
from eth_utils import keccak
import json
from pygate_grpc.client import PowerGateClient
from ipfs_async import client as ipfs_client
from redis_conn import provide_async_reader_conn_inst, provide_async_writer_conn_inst
from utils import preprocess_dag

""" Inititalize the logger """
pruning_logger = logging.getLogger(__name__)
pruning_logger.setLevel(logging.DEBUG)
formatter = logging.Formatter("%(levelname)-8s %(name)-4s %(asctime)s %(msecs)d %(module)s-%(funcName)s: %(message)s")
stream_handler = logging.StreamHandler(sys.stdout)
stream_handler.setFormatter(formatter)
pruning_logger.addHandler(stream_handler)
pruning_logger.debug("Initialized logger")


@provide_async_reader_conn_inst
@provide_async_writer_conn_inst
async def choose_targets(
        reader_redis_conn=None,
        writer_redis_conn=None
):
    """
        - This function is responsible for choosing target cids that need to be un-pinned and put them into
        redis list
    """

    # Get the list of all projectIds that are saved in the redis set
    key = f"storedProjectIds"
    project_ids: set = await reader_redis_conn.smembers(key)  # No await since this is not aioredis connection
    for project_id in project_ids:
        project_id = project_id.decode('utf-8')
        pruning_logger.debug("Getting block height for project: " + project_id)

        # Get the block_height for project_id
        last_block_key = f"projectID:{project_id}:blockHeight"
        out: bytes = await reader_redis_conn.get(last_block_key)
        if out:
            max_block_height: int = int(out.decode('utf-8'))
        else:
            """ No blocks exists for that projectId"""
            max_block_height: int = 0

        pruning_logger.debug("Retrieved block height:")
        pruning_logger.debug(max_block_height)
        # Check if the project_id is still a small chaiN
        if max_block_height < settings.max_ipfs_blocks:
            continue

        # Get the height of last pruned cid
        last_pruned_key = f"lastPruned:{project_id}"
        out: bytes = await reader_redis_conn.get(last_pruned_key)
        if out:
            last_pruned_height = int(out.decode('utf-8'))
        else:
            _ = await writer_redis_conn.set(last_pruned_key, 0)
            last_pruned_height = 0
        pruning_logger.debug("Last Pruned Height: ")
        pruning_logger.debug(last_pruned_height)

        if settings.container_height + last_pruned_height > max_block_height - settings.max_ipfs_blocks:
            """ Check if are settings.container_height amount of blocks left """
            pruning_logger.debug("Not enough blocks for: ")
            pruning_logger.debug(project_id)
            continue

        to_height, from_height = last_pruned_height + settings.container_height, last_pruned_height + 1

        # Get all the blockCids and payloadCids
        block_cids_key = f"projectID:{project_id}:Cids"
        block_cids_to_prune = await reader_redis_conn.zrangebyscore(
            key=block_cids_key,
            min=from_height,
            max=to_height,
            withscores=True
        )

        """ Add dag blocks to the to-unpin-list """
        dag_unpin_key = f"projectID:{project_id}:targetDags"
        for cid, height in block_cids_to_prune:
            _ = await writer_redis_conn.zadd(
                key=dag_unpin_key,
                member=cid,
                score=int(height)
            )

        """ Store the min and max heights of the dag blocks that need to be pruned """
        target_height_key = f"projectID:{project_id}:pruneFromHeight"
        _ = await writer_redis_conn.set(target_height_key, from_height)
        target_height_key = f"projectID:{project_id}:pruneToHeight"
        _ = await writer_redis_conn.set(target_height_key, to_height)
        pruning_logger.debug("Saving from and to height:")
        pruning_logger.debug(from_height)
        pruning_logger.debug(to_height)

        # Add this project_id to the set toBeUnpinned ProjectIds
        to_unpin_projects_key = f"toUnpinProjects"
        _ = await writer_redis_conn.sadd(to_unpin_projects_key, project_id)


@provide_async_reader_conn_inst
async def build_container(
        project_id: str,
        reader_redis_conn=None,
):
    """
        - Generate a unique ID for this container
        - Iterate through each DAG block.
        - Retrieve the payload for each DAG block
        - Add the payload into the DAG block itself 
        - Add the block_dag_cid to the bloom_filter of this container
    """

    target_height_key = f"projectID:{project_id}:pruneFromHeight"
    from_height = await reader_redis_conn.get(target_height_key)
    target_height_key = f"projectID:{project_id}:pruneToHeight"
    to_height = await reader_redis_conn.get(target_height_key)
    pruning_logger.debug("Retrieved from and to height")
    pruning_logger.debug(from_height)
    pruning_logger.debug(to_height)

    """ Create the unique container ID for this project_id """
    container_id_data = {
        'fromHeight': int(from_height),
        'toHeight': int(to_height),
        'projectId': project_id
    }
    container_id = keccak(text=json.dumps(container_id_data)).hex()

    """ Create a Bloom filter """
    bloom_filter_settings = settings.bloom_filter_settings
    bloom_filter_settings['filename'] = f"bloom_filter_objects/{container_id}.bloom"
    bloom_filter = BloomFilter(**bloom_filter_settings)

    """ Get all the targets """
    dag_unpin_key = f"projectID:{project_id}:targetDags"
    dag_cids = await reader_redis_conn.zrevrange(
        key=dag_unpin_key,
        start=0,
        stop=-1,
        withscores=False
    )

    container = {'dagChain': [], 'payloads': {}}
    for dag_cid in dag_cids:
        dag_cid = dag_cid.decode('utf-8')

        pruning_logger.debug("Creating dag block for: ")
        pruning_logger.debug(dag_cid)

        _dag_block = await ipfs_client.dag.get(dag_cid)
        dag_block = _dag_block.as_json()
        dag_block = preprocess_dag(dag_block)

        """ Get the snapshot cid and retrieve the data of the snapshot from ipfs """
        snapshot_cid = dag_block['data']['cid']
        _snapshot_payload = await ipfs_client.cat(snapshot_cid)
        snapshot_payload = _snapshot_payload.decode('utf-8')

        """ Add the payload to the container """
        container['payloads'][snapshot_cid] = snapshot_payload

        """ Add the payload cid to the bloom filter """
        bloom_filter.add(dag_cid)
        container['dagChain'].append({dag_cid: dag_block})

    container_data = {
        'container': container,
        'bloomFilterSettings': bloom_filter_settings,
        'containerId': container_id,
        'fromHeight': int(from_height),
        'toHeight': int(to_height),
        'projectId': project_id
    }
    return container_data


@provide_async_reader_conn_inst
@provide_async_writer_conn_inst
async def backup_to_filecoin(
        container_data: dict,
        reader_redis_conn=None,
        writer_redis_conn=None
):
    """
        - Convert container data to json string. 
        - Push it to filecoin
        - Save the job Id
    """
    powgate_client = PowerGateClient(settings.POWERGATE_CLIENT_ADDR, False)

    """ Get the token """
    KEY = f"filecoinToken:{container_data['projectId']}"
    token = await reader_redis_conn.get(KEY)
    if not token:
        user = powgate_client.admin.users.create()
        token = user.token
        _ = await writer_redis_conn.set(KEY, token)

    else:
        token = token.decode('utf-8')

    """ Convert the data to json string """
    json_data = json.dumps(container_data).encode('utf-8')

    # Stage and push the data
    stage_res = powgate_client.data.stage_bytes(json_data, token=token)
    job = powgate_client.config.apply(stage_res.cid, override=True, token=token)

    # Save the jobId and other data
    fields = {k: v for k, v in container_data.items()}
    _ = fields.pop('container')
    fields['bloomFilterSettings'] = json.dumps(fields.pop('bloomFilterSettings'))
    fields['containerCid'] = stage_res.cid
    fields['jobId'] = job.jobId

    job_status = {
        'jobId': job.jobId,
        'jobStatus': 'JOB_STATUS_UNCHECKED',
        'retry': 0,
    }

    container_id = fields.pop('containerId')
    project_id = fields.pop('projectId')

    container_data_key = f"projectID:{project_id}:containerData:{container_id}"
    _ = await writer_redis_conn.hmset_dict(
        key=container_data_key,
        **fields
    )
    result = {'containerCid': stage_res.cid, 'jobId': job.jobId}
    return result


@provide_async_writer_conn_inst
@provide_async_reader_conn_inst
async def prune_targets(
    reader_redis_conn=None,
    writer_redis_conn=None
):
    """
        - Get the list of all projectID's that are in our protocol
        - iterate through each of the projectID:
            - get the list of Dag blocks that are chosen as targets
    :return:
        None
    """
    # Get all project_ids
    to_unpin_projects_key = f"toUnpinProjects"
    project_ids: set = await reader_redis_conn.smembers(to_unpin_projects_key)

    # Get all the cids in the redis set toBeUnpinnedCids
    for project_id in project_ids:

        project_id = project_id.decode('utf-8')
        pruning_logger.debug("Processing projectID: ")
        pruning_logger.debug(project_id)

        """ Get the container for each project_id """
        container_data = await build_container(project_id)

        """ Save the container into a json file """
        container_file_path = f"containers/{container_data['containerId']}.json"
        with open(container_file_path, 'w') as container_file:
            json.dump(container_data, container_file)

        """ Backup the container data """
        if settings.backup_target == 'FILECOIN':
            result = await backup_to_filecoin(container_data)
            pruning_logger.debug("Container data has been successfully backed up to filecoin: ")
            pruning_logger.debug(result)

        """ Once the container has been backed up, then add it to the list of containers available """
        containers_created_key = f"projectID:{project_id}:containers"
        _ = await writer_redis_conn.zadd(
            key=containers_created_key,
            member=container_data['containerId'],
            score=container_data['toHeight']
        )

        """ For each payload in the dag structure, unpin it """
        for dag_data in container_data['container']['dagChain']:
            dag_cid, dag_block = next(iter(dag_data.items()))
            snapshot_cid = dag_block['data']['cid']
            try:
                _ = await ipfs_client.pin.rm(snapshot_cid)
            except ipfshttpclient.exceptions.ErrorResponse as e:
                pruning_logger.debug("This cid is not pinned....")

            """ Remove the dag_block from the list of targetPrunes """
            dag_unpin_key = f"projectID:{project_id}:targetDags"
            _ = await writer_redis_conn.zrem(
                key=dag_unpin_key,
                member=dag_cid
            )

        # Set the lastPruned for this projectId
        last_pruned_key = f"lastPruned:{project_id}"
        _ = await writer_redis_conn.set(last_pruned_key, container_data['toHeight'])

        """ Remove the project Id from the list of target Project IDs """
        _ = await writer_redis_conn.srem(to_unpin_projects_key, project_id)
        pruning_logger.debug('Successfully Pruned....')


def verifier_crash_cb(fut: asyncio.Future):
    try:
        exc = fut.exception()
    except (asyncio.CancelledError, aioredis.ConnectionClosedError):
        pruning_logger.error('Respawning pruning task...')
        t = asyncio.ensure_future(periodic_pruning())
        t.add_done_callback(verifier_crash_cb)
    except Exception as e:
        pruning_logger.error('Pruning task crashed')
        pruning_logger.error(e, exc_info=True)


async def periodic_pruning():
    while True:
        await asyncio.gather(
            choose_targets(),
        )
        await asyncio.gather(
            prune_targets(),
            asyncio.sleep(settings.pruning_service_interval),
        )

if __name__ == "__main__":
    pruning_logger.debug("Starting the loop")
    f = asyncio.ensure_future(periodic_pruning())
    f.add_done_callback(verifier_crash_cb)
    try:
        asyncio.get_event_loop().run_until_complete(asyncio.gather(f))
    except:
        asyncio.get_running_loop().stop()

