from eth_utils import keccak
from async_ipfshttpclient.main import ipfs_read_client
from utils import redis_keys
from utils import helper_functions
from utils import dag_utils
from redis import asyncio as aioredis
import json
import logging
import sys
from config import settings
from bloom_filter import BloomFilter
from tenacity import wait_random_exponential, stop_after_attempt, retry
from data_models import ProjectBlockHeightStatus, PendingTransaction

retrieval_utils_logger = logging.getLogger(__name__)
retrieval_utils_logger.setLevel(level=logging.DEBUG)
formatter = logging.Formatter("%(levelname)-8s %(name)-4s %(asctime)s %(msecs)d %(module)s-%(funcName)s: %(message)s")
stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setFormatter(formatter)
stdout_handler.setLevel(logging.DEBUG)
retrieval_utils_logger.addHandler(stdout_handler)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setFormatter(formatter)
stderr_handler.setLevel(logging.ERROR)
retrieval_utils_logger.addHandler(stderr_handler)
retrieval_utils_logger.debug("Initialized logger")

SNAPSHOT_STATUS_MAP = {
    "SNAPSHOT_COMMIT_PENDING": 1,
    "TX_ACK_PENDING": 2,
    "TX_CONFIRMATION_PENDING": 3,
    "TX_CONFIRMED": 4
}


async def check_ipfs_pinned(from_height: int, to_height: int):
    """
        - Given the span, check if these blocks exist of IPFS yet
        or if they have to be retrieved through the container
    """
    pass


def check_intersection(span_a, span_b):
    """
        - Given two spans, check the intersection between them
    """

    set_a = set(range(span_a[0], span_a[1] + 1))
    set_b = set(range(span_b[0], span_b[1] + 1))
    overlap = set_a & set_b
    result = float(len(overlap)) / len(set_a)
    result = result * 100
    return result


async def check_overlap(
        from_height: int,
        to_height: int,
        project_id: str,
        reader_redis_conn: aioredis.Redis
):
    """
        - Given a span, check its intersection with other spans and find
        the span which intersects the most.
        - If there is no intersection with any of the spans, the return -1
    """

    # Get the list of all spans
    live_span_key = redis_keys.get_live_spans_key(project_id=project_id, span_id="*")
    span_keys = await reader_redis_conn.keys(pattern=live_span_key)
    retrieval_utils_logger.debug(span_keys)
    # Iterate through each span and check the intersection for each span
    # with the given from_height and to_height
    max_overlap = 0.0
    max_span_id = ""
    each_height_spans = {}
    for span_key in span_keys:
        if isinstance(span_key, bytes):
            span_key = span_key.decode('utf-8')

        out = await reader_redis_conn.get(span_key)
        if out:
            try:
                span_data = json.loads(out.decode('utf-8'))
            except json.JSONDecodeError as jerr:
                retrieval_utils_logger.error(jerr, exc_info=True)
                continue
        else:
            continue

        target_span = (span_data.get('fromHeight'), span_data.get('toHeight'))
        try:
            overlap = check_intersection(span_a=(from_height, to_height), span_b=target_span)
        except Exception as e:
            retrieval_utils_logger.debug("Check intersection function failed.. ")
            retrieval_utils_logger.error(e, exc_info=True)
            overlap = 0.0

        if overlap > max_overlap:
            max_overlap = overlap
            max_span_id = span_key.split(':')[-1]

        # Check overlap for each height:
        current_height = from_height
        while current_height <= to_height:
            if each_height_spans.get(current_height) is None:
                try:
                    overlap = check_intersection(span_a=(current_height, current_height), span_b=target_span)
                except Exception as e:
                    retrieval_utils_logger.debug("Check intersection Failed")
                if overlap == 100.0:
                    each_height_spans[current_height] = span_key.split(':')[-1]
            current_height = current_height + 1

    return max_overlap, max_span_id, each_height_spans


async def get_container_id(
        dag_block_height: int,
        dag_cid: str,
        project_id: str,
        reader_redis_conn: aioredis.Redis
):
    """
        - Given the dag_block_height and dag_cid, get the container_id for the container
        which holds this dag_block
    """

    target_containers = await reader_redis_conn.zrangebyscore(
        name=redis_keys.get_containers_created_key(project_id),
        max=settings.container_height * 2 + dag_block_height + 1,
        min=dag_block_height - settings.container_height * 2 - 1
    )
    container_id = None
    container_data = dict()
    for container_id in target_containers:
        """ Get the data for the container """
        container_id = container_id.decode('utf-8')
        container_data_key = redis_keys.get_container_data_key(container_id)
        out = await reader_redis_conn.hgetall(container_data_key)
        container_data = {k.decode('utf-8'): v.decode('utf-8') for k, v in out.items()}
        bloom_filter_settings = json.loads(container_data['bloomFilterSettings'])
        bloom_object = BloomFilter(**bloom_filter_settings)
        if dag_cid in bloom_object:
            break

    return container_id, container_data


async def check_container_cached(
        container_id: str,
        reader_redis_conn: aioredis.Redis
):
    """
        - Given the container_id check if the data for that container is
        cached on redis
    """

    cached_container_key = redis_keys.get_cached_containers_key(container_id)
    out = await reader_redis_conn.exists(cached_container_key)
    if out is 1:
        return True
    else:
        return False


async def check_containers(
        from_height,
        to_height,
        project_id: str,
        reader_redis_conn: aioredis.Redis,
        each_height_spans: dict

):
    """
        - Given the from_height and to_height, check for each dag_cid, what container is required
        and if that container is cached
    """
    # Get the dag cid's for the span
    out = await reader_redis_conn.zrangebyscore(
        name=redis_keys.get_dag_cids_key(project_id=project_id),
        max=to_height,
        min=from_height,
        withscores=True
    )

    last_pruned_height = await helper_functions.get_last_pruned_height(project_id=project_id,
                                                                       reader_redis_conn=reader_redis_conn)

    containers_required = []
    cached = {}
    for dag_cid, dag_block_height in out:
        dag_cid = dag_cid.decode('utf-8')

        # Check if the dag_cid is beyond the range of max_ipfs_blocks
        if (dag_block_height > last_pruned_height) or (each_height_spans.get(dag_block_height) is not None):
            # The dag block is safe
            continue
        else:
            container_id, container_data = await get_container_id(
                dag_block_height=dag_block_height,
                dag_cid=dag_cid,
                project_id=project_id,
                reader_redis_conn=reader_redis_conn
            )

            containers_required.append({container_id: container_data})
            is_cached = await check_container_cached(container_id, reader_redis_conn=reader_redis_conn)
            cached[container_id] = is_cached

    return containers_required, cached


async def fetch_from_span(
        from_height: int,
        to_height: int,
        span_id: str,
        project_id: str,
        reader_redis_conn: aioredis.Redis
):
    """
        - Given the span_id and the span, fetch blocks in that range, if
        they exist
    """
    live_span_key = redis_keys.get_live_spans_key(span_id=span_id, project_id=project_id)
    span_data = await reader_redis_conn.get(live_span_key)

    if span_data:
        try:
            span_data = json.loads(span_data)
        except json.JSONDecodeError as jerr:
            retrieval_utils_logger.error(jerr, exc_info=True)
            return -1

    current_height = from_height
    blocks = []
    while current_height <= to_height:
        blocks.append(span_data['dag_blocks'][current_height])
        current_height = current_height + 1

    return blocks


async def save_span(
        from_height: int,
        to_height: int,
        project_id: str,
        dag_blocks: dict,
        writer_redis_conn: aioredis.Redis
):
    """
        - Given the span, save it.
        - Important to assign it a timeout to make sure that key disappears after a
        certain time.
    """
    span_data = {
        'fromHeight': from_height,
        'toHeight': to_height,
        'projectId': project_id
    }
    span_id = keccak(text=json.dumps(span_data)).hex()

    span_data.pop('projectId')
    span_data['dag_blocks'] = dag_blocks
    live_span_key = redis_keys.get_live_spans_key(project_id=project_id, span_id=span_id)
    _ = await writer_redis_conn.set(live_span_key, json.dumps(span_data))
    _ = await writer_redis_conn.expire(live_span_key, settings.span_expire_timeout)


async def fetch_blocks(
        from_height: int,
        to_height: int,
        project_id: str,
        data_flag: bool,
        reader_redis_conn: aioredis.Redis
):
    """
        - Given the from_height and to_height fetch the blocks based on whether there are any spans
        that exists or not
    """

    max_overlap, max_span_id, each_height_spans = await check_overlap(
        project_id=project_id,
        from_height=from_height,
        to_height=to_height,
        reader_redis_conn=reader_redis_conn
    )

    current_height = to_height
    dag_blocks = {}
    while current_height >= from_height:
        dag_cid = await helper_functions.get_dag_cid(project_id=project_id, block_height=current_height,
                                                     reader_redis_conn=reader_redis_conn)
        if each_height_spans.get(current_height) is None:
            # not in span (supposed to be a LRU cache of sorts with a moving window as DAG blocks keep piling up)
            dag_cid = await helper_functions.get_dag_cid(project_id=project_id, block_height=current_height,
                                                         reader_redis_conn=reader_redis_conn)
            dag_block = await dag_utils.get_dag_block(dag_cid)
            if data_flag:
                dag_block = await retrieve_block_data(block_dag_cid=dag_cid, data_flag=1)
        else:
            dag_block: list = await fetch_from_span(
                from_height=current_height,
                to_height=current_height,
                span_id=each_height_spans[current_height],
                project_id=project_id,
                reader_redis_conn=reader_redis_conn
            )

        dag_blocks[dag_cid] = dag_block
        current_height = current_height - 1

        # retrieval_utils_logger.debug(dag_blocks)

    return dag_blocks


# BLOCK_STATUS_SNAPSHOT_COMMIT_PENDING = 1
# TX_ACK_PENDING=2
# TX_CONFIRMATION_PENDING = 3,
# TX_CONFIRMED=4,

async def retrieve_block_status(
        project_id: str,
        project_block_height: int,  # supplying 0 indicates finalized height of project should be fetched separately
        block_height: int,
        reader_redis_conn: aioredis.Redis,
        writer_redis_conn: aioredis.Redis
) -> ProjectBlockHeightStatus:
    block_status = ProjectBlockHeightStatus(
        project_id=project_id,
        block_height=block_height,
    )
    if project_block_height == 0:
        project_block_height = await helper_functions.get_block_height(
            project_id=project_id,
            reader_redis_conn=reader_redis_conn
        )
    if block_height > project_block_height:
        """This means the queried blockHeight is not yet finalized. """
        """ Access the payloadCId at block_height """
        project_payload_cids_key_zset = redis_keys.get_payload_cids_key(project_id)
        r = await reader_redis_conn.zrangebyscore(
            name=project_payload_cids_key_zset,
            min=block_height,
            max=block_height,
            withscores=False
        )
        if len(r) == 0:
            return block_status
        payload_cid = r[0].decode('utf-8')

        project_pending_txns_key_zset = redis_keys.get_pending_transactions_key(project_id)
        pending_txs = await reader_redis_conn.zrangebyscore(
            name=project_pending_txns_key_zset,
            min=block_height,
            max=block_height,
            withscores=False
        )
        if len(pending_txs) == 0:
            block_status.status = SNAPSHOT_STATUS_MAP['TX_ACK_PENDING']
            return block_status
        
        all_empty_txhash = True
        for tx in pending_txs:
            pending_txn = PendingTransaction.parse_raw(tx)
            # itrate until we find a entry with txHash
            if pending_txn.txHash == None or pending_txn.txHash == "":
                continue
            
            # set this false, when atleast one txHash exist
            all_empty_txhash = False
            
            # check if tx is confirmed
            if pending_txn.lastTouchedBlock == -1:
                block_status.tx_hash = pending_txn.txHash
                block_status.status = SNAPSHOT_STATUS_MAP['TX_CONFIRMED']
                block_status.payload_cid = payload_cid
                return

        # if all the entry had empty txHash
        if all_empty_txhash:
            block_status.status = SNAPSHOT_STATUS_MAP['TX_ACK_PENDING']
            return block_status
        
        # if txHash was there but none with lastTouchedBlock == -1 then take latest pending tx
        block_status.payload_cid = payload_cid
        pending_txn = PendingTransaction.parse_raw(pending_txs[0])
        block_status.tx_hash = pending_txn.txHash
        block_status.status = SNAPSHOT_STATUS_MAP['TX_CONFIRMATION_PENDING']

    else:
        """ Access the DAG CID at block_height """
        project_payload_cids_key_zset = redis_keys.get_dag_cids_key(project_id)
        r = await reader_redis_conn.zrangebyscore(
            name=project_payload_cids_key_zset,
            min=block_height,
            max=block_height,
            withscores=False
        )
        if len(r) == 0:
            #This scenario can happen when a project's blockHeight is pushed ahead
            # and current height is not present in the project DAG.
            return None
        dag_cid = r[0].decode('utf-8')

        block = await retrieve_block_data(dag_cid, writer_redis_conn=writer_redis_conn, data_flag=0)
        block_status.payload_cid = block['data']['cid']['/']
        block_status.tx_hash = block['txHash']
        block_status.status = SNAPSHOT_STATUS_MAP['TX_CONFIRMED']
    return block_status


# TODO: refactor function or introduce an enum/pydantic model against data_flag param for readability/maintainability
# passing ints against a flag is terrible coding practice
@retry(
    reraise=True,
    wait=wait_random_exponential(multiplier=1, max=30),
    stop=stop_after_attempt(3),
)
async def retrieve_block_data(block_dag_cid, writer_redis_conn=None, data_flag=0):
    """
        A function which will get dag block from ipfs and also increment its hits
        Args:
            block_dag_cid:str - The cid of the dag block that needs to be retrieved
            writer_redis_conn: redis.Redis - To increase hitcount to dag cid, do not pass args if caching is not desired
            data_flag:int - This is a flag which can take three values:
                0 - Return only the dag block and not its payload data
                1 - Return the dag block along with its payload data
                2 - Return only the payload data
    """

    assert data_flag in range(0, 3), \
        f"The value of data: {data_flag} is invalid. It can take values: 0, 1 or 2"

    """ TODO: Increment hits on block, to be used for some caching purpose """
    # block_dag_hits_key = redis_keys.get_hits_dag_block_key()
    # if writer_redis_conn:
    #     r = await writer_redis_conn.zincrby(block_dag_hits_key, 1.0, block_dag_cid)
    #     retrieval_utils_logger.debug("Block hit for: ")
    #     retrieval_utils_logger.debug(block_dag_cid)
    #     retrieval_utils_logger.debug(r)

    """ Retrieve the DAG block from ipfs """
    _block = await ipfs_read_client.dag.get(block_dag_cid)
    block = _block.as_json()
    # block = preprocess_dag(block)
    if data_flag == 0:
        return block

    payload = dict()

    """ Get the payload Data """
    payload_data = await ipfs_read_client.cat(block['data']['cid']['/'])
    payload_data = json.loads(payload_data)
    payload['payload'] = payload_data
    payload['cid'] = block['data']['cid']['/']

    if data_flag == 1:
        block['data'] = payload
        return block

    if data_flag == 2:
        return payload

async def retrieve_payload_cid(project_id: str, block_height: int,reader_redis_conn=None):
    """
        - Given a projectId and block_height, get its payloadCID from redis
    """
    project_payload_cids_key_zset = redis_keys.get_payload_cids_key(project_id)
    payload_cid = ""
    if reader_redis_conn:
        r = await reader_redis_conn.zrangebyscore(
            name=project_payload_cids_key_zset,
            min=block_height,
            max=block_height,
            withscores=False
        )
        if len(r) == 0:
            return payload_cid
        payload_cid = r[0].decode('utf-8')
    return payload_cid


async def retrieve_payload_data(payload_cid, writer_redis_conn=None):
    """
        - Given a payload_cid, get its data from ipfs, at the same time increase its hit
    """
    #payload_key = redis_keys.get_hits_payload_data_key()
    #if writer_redis_conn:
        #r = await writer_redis_conn.zincrby(payload_key, 1.0, payload_cid)
        #retrieval_utils_logger.debug("Payload Data hit for: ")
        #retrieval_utils_logger.debug(payload_cid)

    """ Get the payload Data from ipfs """
    _payload_data = await ipfs_read_client.cat(payload_cid)
    if not _payload_data:
        return None

    if isinstance(_payload_data, str):
        payload_data = _payload_data
    else:
        payload_data = payload_data.decode('utf-8')

    return payload_data


async def get_blocks_from_container(container_id, dag_cids: list):
    """
        Given the dag_cids, get those dag block from the given container
    """
    pass


### SHARED IN-MEMORY CACHE FOR DAG BLOCK DATA ####
### INCLUDES PAYLOAD DATA ########################
SHARED_DAG_BLOCKS_CACHE = {}

def prune_dag_block_cache(cache_size_unit, shared_cache):

    # only prune when size of cache greater than cache_size_unit * 5
    if len(shared_cache) <= cache_size_unit * 5:
        return

    # prune to make size of dict cache_size * 4
    pruning_length = cache_size_unit * 4
    # sort dict
    ordered_items = sorted(shared_cache.items(), key=lambda item: item[1]['block_height'])
    # prune result list and make it a dict again
    shared_cache = dict(ordered_items[:pruning_length])

async def get_dag_block_by_height(project_id, block_height, reader_redis_conn: aioredis.Redis, cache_size_unit):
    dag_block = {}
    # init global shared cache if doesn't exit
    if 'SHARED_DAG_BLOCKS_CACHE' not in globals():
        global SHARED_DAG_BLOCKS_CACHE
        SHARED_DAG_BLOCKS_CACHE = {}

    dag_cid = await helper_functions.get_dag_cid(
        project_id=project_id,
        block_height=block_height,
        reader_redis_conn=reader_redis_conn
    )
    if not dag_cid:
        return {}

    # use cache if available
    if dag_cid and SHARED_DAG_BLOCKS_CACHE.get(dag_cid, False):
        return SHARED_DAG_BLOCKS_CACHE.get(dag_cid)['data']

    dag_block = await retrieve_block_data(block_dag_cid=dag_cid, data_flag=1)
    dag_block = dag_block if dag_block else {}
    
    dag_block["dagCid"] = dag_cid

    # cache result
    SHARED_DAG_BLOCKS_CACHE[dag_cid] = {'data': dag_block, 'block_height': block_height}
    prune_dag_block_cache(cache_size_unit, SHARED_DAG_BLOCKS_CACHE)

    return dag_block