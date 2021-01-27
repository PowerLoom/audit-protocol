import logging
import sys
import aioredis
from dynaconf import settings
import time
import asyncio
from bloom_filter import BloomFilter
from eth_utils import keccak
import json
from utils.ipfs_async import client as ipfs_client
from utils.redis_conn import provide_async_reader_conn_inst, provide_async_writer_conn_inst
from utils.diffmap_utils import preprocess_dag
from data_models import ContainerData, FilecoinJobData, SiaSkynetData, BackupMetaData, SiaRenterData
from pydantic import ValidationError
from typing import Union, List
import os
from deal_watcher_service import unpin_cids
import coloredlogs
from utils import helper_functions
from utils import redis_keys, backup_utils

""" Inititalize the logger """
pruning_logger = logging.getLogger(__name__)
pruning_logger.setLevel(logging.DEBUG)
formatter = logging.Formatter("%(levelname)-8s %(name)-4s %(asctime)s %(msecs)d %(module)s-%(funcName)s: %(message)s")
stream_handler = logging.StreamHandler(sys.stdout)
stream_handler.setFormatter(formatter)
pruning_logger.addHandler(stream_handler)
pruning_logger.debug("Initialized logger")
coloredlogs.install(level="DEBUG", logger=pruning_logger, stream=sys.stdout)

def startup_boilerplate():
    try:
        os.stat(os.getcwd() + '/containers')
    except:
        os.mkdir(os.getcwd() + '/containers')

    try:
        os.stat(os.getcwd() + '/bloom_filter_objects')
    except:
        os.mkdir(os.getcwd() + '/bloom_filter_objects')

    try:
        os.stat(os.getcwd() + '/bloom_filter_objects')
    except:
        os.mkdir(os.getcwd() + '/bloom_filter_objects')

    try:
        os.stat(os.getcwd() + '/temp_files')
    except:
        os.mkdir(os.getcwd() + '/temp_files')


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
    project_ids: set = await reader_redis_conn.smembers(key)
    for project_id in project_ids:
        project_id = project_id.decode('utf-8')
        pruning_logger.debug("Getting block height for project: ")
        pruning_logger.debug(project_id)

        # Get the block_height for project_id
        max_block_height = await helper_functions.get_block_height(
            project_id=project_id,
            reader_redis_conn=reader_redis_conn
        )

        pruning_logger.debug("Retrieved block height:")
        pruning_logger.debug(max_block_height)
        # Check if the project_id is still a small chaiN
        if max_block_height < settings.MAX_IPFS_BLOCKS:
            continue

        # Get the height of last pruned cid
        last_pruned_height = await helper_functions.get_last_pruned_height(
            project_id=project_id,
            reader_redis_conn=reader_redis_conn,
            writer_redis_conn=writer_redis_conn
        )
        pruning_logger.debug("Last Pruned Height: ")
        pruning_logger.debug(last_pruned_height)

        if settings.CONTAINER_HEIGHT + last_pruned_height > max_block_height - settings.MAX_IPFS_BLOCKS:
            """ Check if are settings.container_height amount of blocks left """
            pruning_logger.debug("Not enough blocks for: ")
            pruning_logger.debug(project_id)
            continue

        to_height, from_height = last_pruned_height + settings.CONTAINER_HEIGHT, last_pruned_height + 1

        # Get all the blockCids
        block_cids_key = redis_keys.get_dag_cids_key(project_id=project_id)
        block_cids_to_prune = await reader_redis_conn.zrangebyscore(
            key=block_cids_key,
            min=from_height,
            max=to_height,
            withscores=True
        )

        """ Add dag blocks to the to-unpin-list """
        dag_unpin_key = redis_keys.get_target_dags_key(project_id=project_id)
        for cid, height in block_cids_to_prune:
            _ = await writer_redis_conn.zadd(
                key=dag_unpin_key,
                member=cid,
                score=int(height)
            )

        """ Store the min and max heights of the dag blocks that need to be pruned """
        target_height_key = redis_keys.get_prune_from_height_key(project_id)
        _ = await writer_redis_conn.set(target_height_key, from_height)
        target_height_key = redis_keys.get_prune_to_height_key(project_id)
        _ = await writer_redis_conn.set(target_height_key, to_height)

        pruning_logger.debug("Saving from and to height:")
        pruning_logger.debug(from_height)
        pruning_logger.debug(to_height)

        # Add this project_id to the set toBeUnpinned ProjectIds
        to_unpin_projects_key = redis_keys.get_to_unpin_projects_key()
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

    target_height_key = redis_keys.get_prune_from_height_key(project_id)
    from_height = await reader_redis_conn.get(target_height_key)
    target_height_key = redis_keys.get_prune_to_height_key(project_id)
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
    bloom_filter_settings = settings.BLOOM_FILTER_SETTINGS
    bloom_filter_settings['filename'] = f"bloom_filter_objects/{container_id}.bloom"
    bloom_filter = BloomFilter(**bloom_filter_settings)

    """ Get all the targets """
    dag_unpin_key = redis_keys.get_target_dags_key(project_id)
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

    """ Create Time stamp"""
    timestamp = int(time.time())

    container_data = {
        'container': container,
        'bloomFilterSettings': bloom_filter_settings,
        'containerId': container_id,
        'fromHeight': int(from_height),
        'toHeight': int(to_height),
        'projectId': project_id,
        'timestamp': timestamp
    }
    return container_data


async def store_container_data(
        container_data: dict,
        backup_targets: List[str],
        backup_metadata: dict,
        writer_redis_conn=None
) -> int:
    """
    - Store the metadata for backed up containers on redis
    - Returns -1 if there was a failure, else returns 0
    """
    container_meta_data = {k: v for k, v in container_data.items()}

    _ = container_meta_data.pop('container')
    container_id = container_meta_data.pop('containerId')

    try:
        backup_metadata_obj = BackupMetaData(**backup_metadata)
    except ValidationError as verr:
        pruning_logger.debug("There was an error while creating BackupMetaData model:")
        pruning_logger.error(verr, exc_info=True)
        return -1

    container_meta_data['backupTargets'] = backup_targets
    container_meta_data['backupMetaData'] = backup_metadata_obj

    try:
        container_meta_data = ContainerData(**container_meta_data)
    except ValidationError as verr:
        pruning_logger.debug("There was an error while creating ContainerData model:")
        pruning_logger.error(verr, exc_info=True)
        return -1

    # Convert some fields to json strings
    try:
        container_meta_data.convert_to_json()
        # I am explicitly encoding the backupTargets because, otherwise there will
        # be multiple escape strings and you would have to json.loads it twice
        container_meta_data.backupTargets = container_meta_data.backupTargets.encode('utf-8')
    except TypeError as terr:
        pruning_logger.debug("There was an error while converting some fields to json: ")
        pruning_logger.error(terr, exc_info=True)
        return -1

    # Store the container meta data on redis
    container_data_key = f"containerData:{container_id}"
    _ = await writer_redis_conn.hmset_dict(
        key=container_data_key,
        **container_meta_data.dict(),
    )

    # If there is a filecoin backup then add the container_id to list of executing containers
    if 'filecoin' in backup_targets:
        new_containers_key = f"executingContainers"
        _ = await writer_redis_conn.sadd(new_containers_key, container_id)

    return 0


@provide_async_writer_conn_inst
@provide_async_reader_conn_inst
async def prune_targets(
    reader_redis_conn=None,
    writer_redis_conn=None
):
    """
        - Get the list of all project_id's that need to be unpinned
        - For each project_id in all project_id's do:
            - Build container for the data that needs to be backed up
            - Backup the container
            - Store the container metadata on redis
            - Add the container_id to the list of container_id's which need to be
            monitored by a deal watcher service and update the storage deals for that
            container.
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
        backup_targets = list()
        backup_metadata = dict()
        filecoin_fail, sia_skynet_fail, sia_renter_fail = True, True, True
        if 'filecoin' in settings.BACKUP_TARGETS:
            filecoin_fail = False
            filecoin_job_data: Union[int, FilecoinJobData] = await backup_utils.backup_to_filecoin(container_data=container_data)
            if filecoin_job_data == -1:
                pruning_logger.warning("Failed to backup the container to filecoin")
                filecoin_fail = True
            else:
                pruning_logger.debug("Container backed up to filecoin successfully")
                backup_targets.append('filecoin')
                backup_metadata['filecoin'] = filecoin_job_data

        if 'sia:skynet' in settings.BACKUP_TARGETS:
            sia_skynet_fail = False
            sia_skynet_data: Union[int, SiaSkynetData] = await backup_utils.backup_to_sia_skynet(container_data=container_data)
            if sia_skynet_data == -1:
                pruning_logger.warning("Failed to backup the container to Sia Skynet")
                sia_skynet_fail = True
            else:
                pruning_logger.debug("Container backed up to Sia Skynet successfully")
                backup_targets.append('sia:skynet')
                backup_metadata['sia_skynet'] = sia_skynet_data
                pruning_logger.debug("Unpinning Cid's")
                _ = await unpin_cids(
                    from_height=container_data['fromHeight'],
                    to_height=container_data['toHeight'],
                    reader_redis_conn=reader_redis_conn,
                    project_id=container_data['projectId']
                )

        if 'sia:renter' in settings.BACKUP_TARGETS:
            sia_renter_fail = False
            sia_renter_data: Union[int, SiaRenterData] = await backup_utils.backup_to_sia_renter(container_data=container_data)
            if sia_renter_data == -1:
                pruning_logger.warning("Failed to backup the container to Sia Renter")
                sia_renter_fail = True
            else:
                pruning_logger.debug("Container backed up to Sia Renter successfully")
                backup_targets.append('sia:renter')
                backup_metadata['sia_renter'] = sia_renter_data
                pruning_logger.debug("Unpinning Cid's")
                _ = await unpin_cids(
                    from_height=container_data['fromHeight'],
                    to_height=container_data['toHeight'],
                    reader_redis_conn=reader_redis_conn,
                    project_id=container_data['projectId']
                )

        if filecoin_fail and sia_skynet_fail and sia_renter_fail:
            pruning_logger.debug("Failed to backup data to any platform")
            continue

        """ Store the container Data on redis"""
        result = await store_container_data(
            container_data=container_data,
            backup_targets=backup_targets,
            backup_metadata=backup_metadata,
            writer_redis_conn=writer_redis_conn
        )

        if result == -1:
            pruning_logger.warning("Failed to store container Meta Data on redis...")
            continue

        """ Once the container has been backed up, then add it to the list of containers available """
        containers_created_key = f"projectID:{project_id}:containers"
        _ = await writer_redis_conn.zadd(
            key=containers_created_key,
            member=container_data['containerId'],
            score=container_data['toHeight']
        )

        # Set the lastPruned for this projectId
        last_pruned_key = f"lastPruned:{project_id}"
        _ = await writer_redis_conn.set(last_pruned_key, container_data['toHeight'])

        """ Remove the project Id from the list of target Project IDs """
        _ = await writer_redis_conn.srem(to_unpin_projects_key, project_id)
        pruning_logger.debug('Successfully Pruned....')
        
        """ Empty up the targetDags redis ZSET"""
        target_dags_key = f"projectID:{project_id}:targetDags"
        _ = await writer_redis_conn.zremrangebyrank(
            key=target_dags_key,
            start=0,
            stop=-1
        )


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
            asyncio.sleep(settings.PRUNING_SERVICE_INTERVAL),
        )

if __name__ == "__main__":
    startup_boilerplate()
    pruning_logger.debug("Starting the loop")
    f = asyncio.ensure_future(periodic_pruning())
    f.add_done_callback(verifier_crash_cb)
    try:
        asyncio.get_event_loop().run_until_complete(asyncio.gather(f))
    except:
        asyncio.get_event_loop().stop()

