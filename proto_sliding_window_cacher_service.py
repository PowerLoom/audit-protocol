import asyncio
import json
import logging
import sys
from functools import wraps
from typing import Tuple
from config import settings
from async_ipfshttpclient.main import AsyncIPFSClient
from exceptions import ProjectFinalizedHeightError
from utils import redis_keys
from utils.redis_conn import RedisPool
from utils import retrieval_utils
from pair_data_aggregation_service import v2_pairs_data
from v2_pairs_daily_stats_snapshotter import v2_pairs_daily_stats_snapshotter
from httpx import AsyncClient, Timeout, Limits, AsyncHTTPTransport
from redis import asyncio as aioredis
import asyncio
import json
import logging
import sys
import os

sliding_cacher_logger = logging.getLogger(__name__)
sliding_cacher_logger.setLevel(logging.DEBUG)
formatter = logging.Formatter("%(levelname)-8s %(name)-4s %(asctime)s %(msecs)d %(module)s-%(funcName)s: %(message)s")
stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setFormatter(formatter)
stdout_handler.setLevel(logging.DEBUG)
sliding_cacher_logger.addHandler(stdout_handler)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setFormatter(formatter)
stderr_handler.setLevel(logging.ERROR)
sliding_cacher_logger.addHandler(stderr_handler)
sliding_cacher_logger.debug("Initialized logger")
# coloredlogs.install(level="DEBUG", logger=sliding_cacher_logger, stream=sys.stdout)

def acquire_bounded_semaphore(fn):
    @wraps(fn)
    async def wrapped(*args, **kwargs):
        sem: asyncio.BoundedSemaphore = kwargs['semaphore']
        await sem.acquire()
        result = None
        try:
            result = await fn(*args, **kwargs)
        except Exception as e:
            sliding_cacher_logger.error('Error in semaphore acquisition decorator: %s', e, exc_info=True)
            if isinstance(e, ProjectFinalizedHeightError):
                result = e  # so that the caller can filter out such bad results
        finally:
            sem.release()
            return result
    return wrapped


def convert_time_period_str_to_timestamp(time_period_str: str):
    ts_map = {'24h': 24 * 60 * 60, '7d': 7 * 24 * 60 * 60}
    return ts_map.get(time_period_str, 60 * 60)  # 1 hour timestamp returned by default


async def find_tail(
        head: int,
        tail: int,
        project_id: str,
        time_period_ts: int,
        redis_conn: aioredis.Redis,
        ipfs_read_client: AsyncIPFSClient
):
    sliding_cacher_logger.debug("Seeking tail starting for projectId: %s:%d",
                                project_id,
                                time_period_ts
    )

    current_height = tail
    head_block = await retrieval_utils.get_dag_block_by_height(
        project_id=project_id,
        block_height=head,
        reader_redis_conn=redis_conn,
        ipfs_read_client=ipfs_read_client
    )
    if not head_block['data']['payload']:
        sliding_cacher_logger.warning(
            "Seeking tail against head block DAG CID %s height %s for time period %s"
            "No payload found while fetching dag_block at height %s for projectId: %s. Skipping building of primary index",
            head_block['data']['dag_cid'], head, time_period_ts,
            current_height, project_id, exc_info=True
        )
        return None
    present_ts = head_block['data']['payload']['timestamp']
    while current_height < head:
        try:
            dag_block = await retrieval_utils.get_dag_block_by_height(
                project_id=project_id,
                block_height=current_height,
                reader_redis_conn=redis_conn,
                ipfs_read_client=ipfs_read_client
            )
            if dag_block and dag_block['data']['payload'] and present_ts - dag_block['data']['payload']['timestamp'] <= time_period_ts:
                sliding_cacher_logger.debug("Found tail after traversing %s blocks for projectId: %s:%d",
                                            abs(current_height - tail),
                                            project_id,
                                            time_period_ts
                )
                return current_height
        except Exception as err:
            sliding_cacher_logger.warning(
                "Seeking tail against head block DAG CID %s height %s for time period %s"
                "Exception while fetching dag_block at height %s for projectId: %s. Error:%s",
                head_block['data']['dag_cid'], head, time_period_ts,
                current_height, project_id, err, exc_info=True
            )
        current_height += 1

    sliding_cacher_logger.error("Could not find tail for projectId:%s", project_id)
    return None


@acquire_bounded_semaphore
async def build_primary_index(
        project_id: str,
        time_period: str,
        head_dag_block_height: int,
        semaphore: asyncio.BoundedSemaphore,
        writer_redis_conn: aioredis.Redis,
        ipfs_read_client: AsyncIPFSClient
):
    """
        : param time_period: supported time_period strings as of now:  ['24h', '7d', '0']
    """
    # find markers
    # NOTE: every periodic run, the head although is always chosen to be the max height
    #  1. maybe don't store it? 2. or, might be useful state information?
    idx_head_key = redis_keys.get_sliding_window_cache_head_marker(project_id, time_period)
    idx_tail_key = redis_keys.get_sliding_window_cache_tail_marker(project_id, time_period)
    head_marker = head_dag_block_height
    tail_marker = None

    # if time_period is 0 then just set head and exit
    if time_period == '0':
        await writer_redis_conn.set(redis_keys.get_sliding_window_cache_head_marker(project_id, time_period), head_marker)
        sliding_cacher_logger.info('Set head at %s index for %s time_period data Project ID: %s', head_marker, time_period, project_id)
        return

    time_period_ts = convert_time_period_str_to_timestamp(time_period)
    markers = [await writer_redis_conn.get(k) for k in [idx_head_key, idx_tail_key]]
    if not all(markers):
        sliding_cacher_logger.info('Finding %s tail marker for the first time for project %s', time_period, project_id)
        #passing prev tail as 1 since we are looking for tail for the first time.
        tail_marker = await find_tail(
            head_marker,1, project_id, time_period_ts, writer_redis_conn, ipfs_read_client
        )
        if not tail_marker:
            sliding_cacher_logger.error(
                'not enough blocks against project ID: %s for %s calculation', project_id, time_period
            )
            return
        await writer_redis_conn.set(redis_keys.get_sliding_window_cache_head_marker(project_id, time_period), head_marker)
        await writer_redis_conn.set(redis_keys.get_sliding_window_cache_tail_marker(project_id, time_period), tail_marker)
        sliding_cacher_logger.info(
            'Set %s - %s index for %s data | First run | Project ID: %s',
            head_marker, tail_marker, time_period, project_id
        )
    else:
        tail_marker = int(markers[1])
        tail_ahead = await find_tail(
            head_marker, tail_marker, project_id, time_period_ts, writer_redis_conn, ipfs_read_client
        )
        if not tail_ahead:
            sliding_cacher_logger.error(
                'not enough blocks against project ID: %s to seek tail ahead for %s calculation | present head: %s',
                project_id, time_period, head_marker
            )
            # do not update markers in cache
            return
        else:
            await writer_redis_conn.set(redis_keys.get_sliding_window_cache_head_marker(project_id, time_period),
                                        head_marker)
            await writer_redis_conn.set(redis_keys.get_sliding_window_cache_tail_marker(project_id, time_period),
                                        tail_ahead)
            sliding_cacher_logger.info(
                'Set %s - %s index for %s data | Project ID: %s',
                head_marker, tail_ahead, time_period, project_id
            )


@acquire_bounded_semaphore
async def get_epoch_end_per_project(
    project_id: str,
    semaphore: asyncio.BoundedSemaphore,
    writer_redis_conn: aioredis.Redis,
    ipfs_read_client: AsyncIPFSClient
) -> Tuple[int, int]:
    """
    returns tuple of (head_dag_block_height_of_project, epoch_end_contained_in_head_dag_block)
    """
    project_height_key = redis_keys.get_block_height_key(project_id)
    max_height = await writer_redis_conn.get(project_height_key)
    if not max_height:
        raise ProjectFinalizedHeightError("Can\'t fetch max block height against project ID: %s", project_id)
    dag_block = await retrieval_utils.get_dag_block_by_height(
        project_id=project_id,
        block_height=int(max_height.decode('utf-8')),
        reader_redis_conn=writer_redis_conn,
        ipfs_read_client=ipfs_read_client
    )
    if not dag_block or not dag_block['data']['payload']:
        return (int(max_height), 0)  # anomalous result to be filtered out by caller
    return (int(max_height), dag_block['data']['payload']['chainHeightRange']['end'])


@acquire_bounded_semaphore
async def get_max_height_pair_project(
    project_id: str,
    height_map: dict,
    registered_projects: list,
    semaphore: asyncio.BoundedSemaphore,
    writer_redis_conn: aioredis.Redis,
    ipfs_read_client: AsyncIPFSClient
):
    project_height_key = redis_keys.get_block_height_key(project_id)
    max_height = await writer_redis_conn.get(project_height_key)
    if not max_height:
        return Exception("Can\'t fetch max block height against project ID: %s", project_id)
    try:
        max_height = int(max_height.decode('utf-8'))
        dag_block = await retrieval_utils.get_dag_block_by_height(
            project_id=project_id,
            block_height=max_height,
            reader_redis_conn=writer_redis_conn,
            ipfs_read_client=ipfs_read_client
        )
        if dag_block:
            height_map[project_id] = {"source_height": dag_block["data"]["payload"]["chainHeightRange"]["end"], "dag_block_height": max_height}
        else:
            sliding_cacher_logger.error("Could not fetch dag block at height %s for project %s",max_height, project_id)
            return Exception("Could not fetch dag block at height %s for project %s",max_height, project_id)
    except Exception as err:
        return err
    finally:
        return max_height


async def build_primary_indexes(ipfs_read_client):
    aioredis_pool = RedisPool()
    await aioredis_pool.populate()
    writer_redis_conn: aioredis.Redis = aioredis_pool.writer_redis_pool
    # project ID -> {"series": ['24h', '7d']}
    registered_projects = await writer_redis_conn.hgetall(
        redis_keys.get_projects_registered_for_cache_indexing_key_with_namespace(settings.pooler_namespace)
        )
    sliding_cacher_logger.debug('Got %d registered projects for indexing', len(registered_projects))
    registered_project_ids = [x.decode('utf-8') for x in registered_projects.keys()]
    registered_projects_ts = [json.loads(v)['series'] for v in registered_projects.values()]
    project_id_to_register_series = dict(zip(registered_project_ids, registered_projects_ts))

    sliding_cacher_logger.debug("Fetching maximum height for all projectIds")
    tasks = list()
    semaphore = asyncio.BoundedSemaphore(20)
    for project_id in registered_project_ids:
        tasks.append(
            get_epoch_end_per_project(
            project_id=project_id, 
            semaphore=semaphore, 
            writer_redis_conn=writer_redis_conn, 
            ipfs_read_client=ipfs_read_client)
        )

    epoch_ends_map = await asyncio.gather(*tasks, return_exceptions=True)
    sliding_cacher_logger.debug("Fetched epoch end information for all projectIds: %s", epoch_ends_map)
    try:
        max_epoch_end = max(filter(lambda x: not(isinstance(x, Exception) or x[1] == 0), epoch_ends_map), key=lambda x: x[1])
    except ValueError:
        sliding_cacher_logger.error("Can\'t find max epoch end for all projects, sleeping till next cycle")
        return
    indexable_projects_to_dag_heights = {k: epoch_ends_map[idx][0] for idx, k in enumerate(registered_project_ids) if not(isinstance(epoch_ends_map[idx], Exception) or epoch_ends_map[idx][1] == 0)}
    
    sliding_cacher_logger.debug(f"Start building indexes for all projects | project_count:{len(project_id_to_register_series)}")
    tasks = list()
    for project_id, head_dag_height in indexable_projects_to_dag_heights.items():
        for time_period in project_id_to_register_series[project_id]:
            fn = build_primary_index(
                project_id=project_id,
                time_period=time_period,
                head_dag_block_height=head_dag_height,
                semaphore=semaphore,
                writer_redis_conn=writer_redis_conn,
                ipfs_read_client=ipfs_read_client
            )
            tasks.append(fn)
    await asyncio.gather(*tasks)
    return max_epoch_end[1]


async def periodic_retrieval():
    # TODO: make these configurable
    async_transport = AsyncHTTPTransport(
        limits=Limits(max_connections=100, max_keepalive_connections=20, keepalive_expiry=30)
    )
    async_httpx_client = AsyncClient(
        timeout=Timeout(timeout=5.0),
        follow_redirects=False,
        transport=async_transport
    )
    ipfs_write_client = AsyncIPFSClient(addr=settings.ipfs.url)
    ipfs_read_client = AsyncIPFSClient(addr=settings.ipfs.reader_url)
    await ipfs_write_client.init_session()
    await ipfs_read_client.init_session()
    aioredis_pool = RedisPool()
    await aioredis_pool.populate()
    redis_conn: aioredis.Redis = aioredis_pool.writer_redis_pool
    while True:
        try:
            common_epoch_end = await build_primary_indexes(ipfs_read_client=ipfs_read_client)
            if common_epoch_end:
                try:
                    await asyncio.gather(
                        v2_pairs_data(async_httpx_client, common_epoch_end, ipfs_write_client, ipfs_read_client),
                        v2_pairs_daily_stats_snapshotter(async_httpx_client, ipfs_write_client, redis_conn),
                        asyncio.sleep(90),
                    )
                except Exception as e:
                    sliding_cacher_logger.error('Error in uniswap summary aggregation: ')
                    sliding_cacher_logger.error(e, exc_info=True)
                    await asyncio.sleep(90)
            else:
                await asyncio.sleep(90)
            sliding_cacher_logger.debug('Finished a cycle of indexing...')
        except Exception as err:
            sliding_cacher_logger.error("Exception occured in indexing and aggregation cycle %s",
            err,
            exc_info=True)
            continue


def verifier_crash_cb(fut: asyncio.Future):
    try:
        exc = fut.exception()
    except asyncio.CancelledError:
        # sliding_cacher_logger.error('Respawning task for populating pair contracts, involved tokens and their metadata...')
        t = asyncio.ensure_future(periodic_retrieval())
        t.add_done_callback(verifier_crash_cb)
    except Exception as e:
        sliding_cacher_logger.error('Indexing task crashed')
        sliding_cacher_logger.error(e, exc_info=True)


if __name__ == "__main__":
    loop = asyncio.get_event_loop()
    f = asyncio.ensure_future(periodic_retrieval())
    f.add_done_callback(verifier_crash_cb)
    try:
        asyncio.get_event_loop().run_until_complete(f)
    except:
        asyncio.get_event_loop().stop()

