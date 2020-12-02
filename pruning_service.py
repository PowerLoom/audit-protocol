import ipfshttpclient
import logging
import sys
import aioredis
from config import settings
import time
import asyncio

""" Initialize ipfs client """
ipfs_client = ipfshttpclient.connect()

""" Inititalize the logger """
pruning_logger = logging.getLogger(__name__)
pruning_logger.setLevel(logging.DEBUG)
formatter = logging.Formatter("%(levelname)-8s %(name)-4s %(asctime)s %(msecs)d %(module)s-%(funcName)s: %(message)s")
stream_handler = logging.StreamHandler(sys.stdout)
stream_handler.setFormatter(formatter)
pruning_logger.addHandler(stream_handler)
pruning_logger.debug("Initialized logger")

""" aioredis boilerplate """
redis_pool = None

async def redis_boilerplate():
    global redis_pool
    redis_pool = await aioredis.create_pool(
        address=(settings.REDIS.HOST, settings.REDIS.PORT),
        maxsize=5
    )


async def choose_targets():

    """
        - This function is responsible for choosing target cids that need to be un-pinned and put them into
        redis list
    """

    """ Create a redis connections """
    redis_conn_raw = await redis_pool.acquire()
    redis_conn = aioredis.Redis(redis_conn_raw)

    # Get the list of all projectIds that are saved in the redis set
    key = f"storedProjectIds"
    project_ids: set = await redis_conn.smembers(key)  # No await since this is not aioredis connection
    for project_id in project_ids:
        project_id = project_id.decode('utf-8')
        pruning_logger.debug("Getting block height for project: " + project_id)

        # Get the block_height for project_id
        last_block_key = f"projectID:{project_id}:blockHeight"
        out: bytes = await redis_conn.get(last_block_key)
        if out:
            max_block_height: int = int(out.decode('utf-8'))
        else:
            """ No blocks exists for that projectId"""
            max_block_height: int = 0

        if max_block_height < settings.max_ipfs_blocks:
            continue

        # Get the height of last pruned cid
        last_pruned_key = f"lastPruned:{project_id}"
        out: bytes = await redis_conn.get(last_pruned_key)
        if out:
            last_pruned_height = out.decode('utf-8')
        else:
            _ = await redis_conn.set(last_pruned_key, 0)
            last_pruned_height = 0
        to_height, from_height = max_block_height - settings.max_ipfs_blocks, last_pruned_height

        # Get all the blockCids and payloadCids
        block_cids_key = f"projectID:{project_id}:Cids"
        block_cids_to_prune = await redis_conn.zrangebyscore(
            key=block_cids_key,
            min=from_height,
            max=to_height,
            withscores=False
        )
        payload_cids_key = f"projectID:{project_id}:payloadCids"
        payload_cids_to_prune = await redis_conn.zrangebyscore(
            key=payload_cids_key,
            min=from_height,
            max=to_height,
            withscores=False
        )
        all_cids = block_cids_to_prune + payload_cids_to_prune
        target_list_key = f"toBeUnpinnedCids:{project_id}"
        for cid in all_cids:
            _ = await redis_conn.sadd(target_list_key, cid)

        # Set the lastPruned for this projectId
        _ = await redis_conn.set(last_pruned_key, to_height)
    redis_pool.release(redis_conn_raw)


async def test_choose_targets():
    """
        - For testing purposes, the cids will be take from the set: testCid
        and block height will be taken from : blockHeight:{cid}
    """
    """ Create a redis connections """
    redis_conn_raw = await redis_pool.acquire()
    redis_conn = aioredis.Redis(redis_conn_raw)

    key = f"storedProjectIds"
    project_ids: set = await redis_conn.smembers(key)  # No await since this is not aioredis connection
    for project_id in project_ids:
        project_id = project_id.decode('utf-8')
        pruning_logger.debug("Getting block height for project: " + project_id)

        # Get the block_height for project_id
        last_block_key = f"testBlockHeight"
        out: bytes = await redis_conn.get(last_block_key)
        if out:
            max_block_height: int = int(out.decode('utf-8'))
        else:
            """ No blocks exists for that projectId"""
            max_block_height: int = 0

        if max_block_height < settings.max_ipfs_blocks:
            pruning_logger.debug("Not enough blocks to unpin. Skipping....")
            continue

        # Get the height of last pruned cid
        last_pruned_key = f"lastPruned:{project_id}"
        out: bytes = await redis_conn.get(last_pruned_key)
        if out:
            last_pruned_height = int(out.decode('utf-8'))
        else:
            _ = await redis_conn.set(last_pruned_key, 0)
            last_pruned_height = 0

        to_height, from_height = max_block_height - settings.max_ipfs_blocks, last_pruned_height

        pruning_logger.debug(f"Selected range to prune: {to_height}, {from_height}")

        # Get all the blockCids and payloadCids
        block_cids_key = f"testDataAt:{project_id}"
        block_cids_to_prune = await redis_conn.zrangebyscore(
            key=block_cids_key,
            min=from_height,
            max=to_height,
            withscores=False
        )

        all_cids = block_cids_to_prune
        target_list_key = f"toBeUnpinnedCids:{project_id}"
        for cid in all_cids:
            _ = await redis_conn.sadd(target_list_key, cid)

        _ = await redis_conn.set(last_pruned_key, to_height+1)
    redis_pool.release(redis_conn_raw)

    pass    

async def prune_targets():
    """
        - This function is responsible for taking target cids from the list and then un-pinning them.
    :return:
        None
    """
    global ipfs_client
    redis_conn_raw = await redis_pool.acquire()
    redis_conn = aioredis.Redis(redis_conn_raw)

    # Get all project_ids
    key = f"storedProjectIds"
    project_ids: set = await redis_conn.smembers(key)

    # Get all the cids in the redis set toBeUnpinnedCids
    for project_id in project_ids:
        project_id = project_id.decode('utf-8')
        all_cids_key = f"toBeUnpinnedCids:{project_id}"
        unpinned_cids_key = f"unpinnedCids:{project_id}"
        all_cids: set = await redis_conn.smembers(key=all_cids_key)
        pruning_logger.debug(f"Found {len(all_cids)} cids to unpin..")
        for cid in all_cids:
            ipfs_client.pin.rm(cid.decode('utf-8'))
            # Remove the cid from toBeUnpinnedCids and Add the cid to unpinnedCids Set
            _ = await redis_conn.srem(all_cids_key,cid)
            _ = await redis_conn.sadd(unpinned_cids_key, cid)
        pruning_logger.debug(f"Unpinned {len(all_cids)} for {project_id}")
    redis_pool.release(redis_conn_raw)


if __name__ == "__main__":
    asyncio.run(redis_boilerplate())
    while True:
        pruning_logger.debug("Running the Prunning")
        asyncio.run(test_choose_targets())
        asyncio.run(prune_targets())
        pruning_logger.debug("Sleeping for 10 secs")
        time.sleep(10)


