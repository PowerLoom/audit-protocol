from typing import Union
from eth_utils import keccak
from httpx import _exceptions as httpx_exceptions
from async_ipfshttpclient.main import AsyncIPFSClient
from utils import redis_keys
from utils import helper_functions
from redis import asyncio as aioredis
import json
import logging
import sys
from config import settings
from bloom_filter import BloomFilter
from tenacity import wait_random_exponential, stop_after_attempt, retry
from data_models import ProjectBlockHeightStatus, PendingTransaction
from utils.file_utils import read_text_file
from utils.dag_utils import get_dag_block

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

async def fetch_blocks(
        from_height: int,
        to_height: int,
        project_id: str,
        data_flag: bool,
        reader_redis_conn: aioredis.Redis,
        ipfs_read_client: AsyncIPFSClient
):
    """
        - Given the from_height and to_height fetch the blocks based on whether there are any spans
        that exists or not
    """
    #TODO: Add support to fetch from archived data using dagSegments and traversal logic.
    current_height = to_height
    dag_blocks = list()
    while current_height >= from_height:
        dag_cid = await helper_functions.get_dag_cid(project_id=project_id, block_height=current_height,
                                                     reader_redis_conn=reader_redis_conn)
        fetch_data_flag = 0 # Get only the DAG Block
        if data_flag:
            fetch_data_flag = 1 # Get DAG Block with data
        dag_block = await retrieve_block_data(
            block_dag_cid=dag_cid, data_flag=fetch_data_flag,
            project_id=project_id, ipfs_read_client=ipfs_read_client)
        dag_block['dagCid'] = dag_cid
        dag_blocks.append(dag_block)
        current_height = current_height - 1

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
        writer_redis_conn: aioredis.Redis,
        ipfs_read_client: AsyncIPFSClient
) -> Union[ProjectBlockHeightStatus, None]:
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
            pending_txn: PendingTransaction = PendingTransaction.parse_raw(tx)
            # itrate until we find a entry with txHash
            if pending_txn.event_data.txHash is None or pending_txn.event_data.txHash == "":
                continue

            # set this false, when atleast one txHash exist
            all_empty_txhash = False

            # check if tx is confirmed
            if pending_txn.lastTouchedBlock == -1:
                block_status.tx_hash = pending_txn.event_data.txHash
                block_status.status = SNAPSHOT_STATUS_MAP['TX_CONFIRMED']
                block_status.payload_cid = payload_cid
                return block_status

        # if all the entry had empty txHash
        if all_empty_txhash:
            block_status.status = SNAPSHOT_STATUS_MAP['TX_ACK_PENDING']
            return block_status

        # if txHash was there but none with lastTouchedBlock == -1 then take latest pending tx
        block_status.payload_cid = payload_cid
        pending_txn = PendingTransaction.parse_raw(pending_txs[0])
        block_status.tx_hash = pending_txn.event_data.txHash
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
            retrieval_utils_logger.error('No dagCid found for project_id: %s, block_height: %s', project_id, block_height)
            return None
        dag_cid = r[0].decode('utf-8')

        block = await retrieve_block_data(
            block_dag_cid=dag_cid,
            project_id=project_id,
            ipfs_read_client=ipfs_read_client,
            data_flag=0
        )
        retrieval_utils_logger.info('In block status check: Got DAG block with cid %s at block height %s: %s', dag_cid, block_height, block)
        if not block or not block.get('data'):
            return None
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
async def retrieve_block_data(
        block_dag_cid:str,
        project_id:str,
        ipfs_read_client: AsyncIPFSClient,
        data_flag=0
):
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

    block = await get_dag_block(block_dag_cid, project_id, ipfs_read_client=ipfs_read_client)
    # handle case of no dag_block or null payload in dag_block
    if data_flag == 0 and block is None:
        return dict()
    retrieval_utils_logger.debug('Retrieved dag block with CID %s: %s', block_dag_cid, block)
    # the data field may not be present in the dag block because of the DAG finalizer omitting null fields in DAG block model while converting to JSON
    if 'data' not in block.keys() or block['data'] is None:  
        if data_flag == 1:
            block['data'] = {'payload': None, 'cid': 'null'}
            return block
        elif data_flag == 0:
            return block
        else:
            return None
    if data_flag == 0:
        return block
    payload = dict()
    """ Get the payload Data """
    payload_data = await retrieve_payload_data(
        payload_cid=block['data']['cid']['/'],
        project_id=project_id,
        ipfs_read_client=ipfs_read_client
    )
    # handle case of null payload in dag_block
    if payload_data:
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


async def retrieve_payload_data(
        payload_cid,
        ipfs_read_client: AsyncIPFSClient,
        project_id,
):
    """
        - Given a payload_cid, get its data from ipfs, at the same time increase its hit
    """
    #payload_key = redis_keys.get_hits_payload_data_key()
    #if writer_redis_conn:
        #r = await writer_redis_conn.zincrby(payload_key, 1.0, payload_cid)
        #retrieval_utils_logger.debug("Payload Data hit for: ")
        #retrieval_utils_logger.debug(payload_cid)
    payload_data = None
    if project_id is not None:
        payload_data = read_text_file(settings.local_cache_path + "/" + project_id + "/"+ payload_cid + ".json", None )
    if payload_data is None:
        # TODO: set such logs to TRACE level
        retrieval_utils_logger.info("Failed to read payload with CID %s for project %s from local cache ",
        payload_cid,project_id)
        # Get the payload Data from ipfs
        try:
            _payload_data = await ipfs_read_client.cat(payload_cid)
        except (httpx_exceptions.TransportError, httpx_exceptions.StreamError) as e:
            retrieval_utils_logger.error("Failed to read payload with CID %s for project %s from IPFS : %s",
            payload_cid,project_id, e)
            return None
        else:
            if not isinstance(_payload_data,str):
                return _payload_data.decode('utf-8')
            else:
                return _payload_data
    else:
        return payload_data

async def get_dag_block_by_height(
        project_id, block_height,
        reader_redis_conn: aioredis.Redis,
        ipfs_read_client: AsyncIPFSClient
) -> dict:
    dag_block = dict()
    dag_cid = await helper_functions.get_dag_cid(
        project_id=project_id,
        block_height=block_height,
        reader_redis_conn=reader_redis_conn
    )
    if not dag_cid:
        return dict()

    try:
        dag_block = await retrieve_block_data(
            block_dag_cid=dag_cid, project_id=project_id,data_flag=1,
            ipfs_read_client=ipfs_read_client
        )
    except Exception as e:
        dag_block = dict()
        dag_block['data'] = {'payload': None, 'cid': 'null'}
        if isinstance(e, httpx_exceptions.HTTPError) or isinstance(e, httpx_exceptions.StreamError):
            retrieval_utils_logger.error(
            "Failed to read dag block with CID %s for project %s from IPFS because of IPFS read error %s | Assigned null payload to block structure: %s",
            dag_cid, project_id, e, dag_block
        )
        else:
            retrieval_utils_logger.error(
                "Failed to read dag block with CID %s for project %s from IPFS because of exception %s | Assigned null payload to block structure: %s",
                dag_cid, project_id, e, dag_block
            )
    dag_block["dagCid"] = dag_cid

    return dag_block
