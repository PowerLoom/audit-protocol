from config import settings
from maticvigil.EVCore import EVCore
from async_ipfshttpclient.main import AsyncIPFSClient
from utils import redis_keys
from utils import helper_functions
from data_models import PendingTransaction, DAGBlock, DAGFinalizerCallback
from tenacity import retry, wait_random, retry_if_exception_type, stop_after_attempt
from typing import Tuple
import async_timeout
import asyncio
import json
import io
import logging
from redis import asyncio as aioredis
import hmac
import sys


class IPFSDAGCreationException(Exception):
    pass


logger = logging.getLogger(__name__)
logger.setLevel(level=logging.DEBUG)

formatter = logging.Formatter("%(levelname)-8s %(name)-4s %(asctime)s %(msecs)d %(module)s-%(funcName)s: %(message)s")

stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setLevel(logging.DEBUG)
stdout_handler.setFormatter(formatter)
stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setLevel(logging.ERROR)
stderr_handler.setFormatter(formatter)
logger.addHandler(stdout_handler)
logger.addHandler(stderr_handler)

if settings.webhook_listener.validate_header_sig:
    evc = EVCore(verbose=True)
else:
    evc = None


def check_signature(core_payload, signature):
    """
    Given the core_payload, check the signature generated by the webhook listener
    """
    if not evc:
        return
    api_key = evc._api_write_key
    _sign_rebuilt = hmac.new(
        key=api_key.encode('utf-8'),
        msg=json.dumps(core_payload).encode('utf-8'),
        digestmod='sha256'
    ).hexdigest()

    logger.debug("Signature from header: ")
    logger.debug(signature)
    logger.debug("Signature rebuilt from core payload: ")
    logger.debug(_sign_rebuilt)

    return _sign_rebuilt == signature


async def update_pending_tx_block_touch(
        pending_tx_set_entry: bytes,
        touched_at_block: int,
        project_id: str,
        tentative_block_height: int,
        writer_redis_conn: aioredis.Redis,
        event_data: dict = None
):
    # update last touched block in pending txs key to ensure it is known this guy is already home
    # first, remove
    r_ = await writer_redis_conn.zrem(
        redis_keys.get_pending_transactions_key(project_id), pending_tx_set_entry
    )
    # then, put in new entry
    new_pending_tx_set_entry_obj: PendingTransaction = PendingTransaction.parse_raw(pending_tx_set_entry)
    new_pending_tx_set_entry_obj.lastTouchedBlock = touched_at_block
    if event_data:
        new_pending_tx_set_entry_obj.event_data = event_data
    r__ = await writer_redis_conn.zadd(
        name=redis_keys.get_pending_transactions_key(project_id),
        mapping={new_pending_tx_set_entry_obj.json(): tentative_block_height}
    )
    return {'status': bool(r_) and bool(r__), 'results': {'zrem': r_, 'zadd': r__}}


async def save_event_data(event_data: DAGFinalizerCallback, pending_tx_set_entry: bytes, writer_redis_conn: aioredis.Redis):
    """
        - Given event_data, save the txHash, timestamp, projectId, snapshotCid, tentativeBlockHeight
          to update state in :pendingTransactions zset
    """

    fields = {
        'txHash': event_data.txHash,
        'projectId': event_data.event_data.projectId,
        'timestamp': event_data.event_data.timestamp,
        'snapshotCid': event_data.event_data.snapshotCid,
        'payloadCommitId': event_data.event_data.payloadCommitId,
        'apiKeyHash': event_data.event_data.apiKeyHash,
        'tentativeBlockHeight': event_data.event_data.tentativeBlockHeight
    }

    return await update_pending_tx_block_touch(
        pending_tx_set_entry=pending_tx_set_entry,
        touched_at_block=-1,
        project_id=fields['projectId'],
        tentative_block_height=int(event_data.event_data.tentativeBlockHeight),
        event_data=fields,
        writer_redis_conn=writer_redis_conn
    )


async def get_dag_block(dag_cid: str, ipfs_read_client: AsyncIPFSClient):
    e_obj = None
    try:
        async with async_timeout.timeout(settings.ipfs_timeout) as cm:
            try:
                dag = await ipfs_read_client.dag.get(dag_cid)
            except Exception as e:
                e_obj = e
    except (asyncio.exceptions.CancelledError, asyncio.exceptions.TimeoutError) as err:
        e_obj = err

    if e_obj or cm.expired:
        return {}

    return dag.as_json()


async def put_dag_block(dag_json: str, ipfs_write_client: AsyncIPFSClient):
    dag_json = dag_json.encode('utf-8')
    out = await ipfs_write_client.dag.put(io.BytesIO(dag_json), pin=True)
    dag_cid = out['Cid']['/']

    return dag_cid


async def get_payload(payload_cid: str, ipfs_read_client: AsyncIPFSClient):
    """ Given the payload cid, retrieve the payload. Payloads are also DAG blocks """
    return await get_dag_block(payload_cid, ipfs_read_client)


@retry(
    reraise=True, wait=wait_random(min=1, max=2),
    retry=retry_if_exception_type(IPFSDAGCreationException),
    stop=stop_after_attempt(5)
)
async def create_dag_block(
        tx_hash: str,
        project_id: str,
        tentative_block_height: int,
        payload_cid: str,
        timestamp: int,
        reader_redis_conn: aioredis.Redis,
        writer_redis_conn: aioredis.Redis,
        ipfs_write_client: AsyncIPFSClient,
        prev_cid_fetch: bool = True
) -> Tuple[str, DAGBlock]:
    """ Get the last dag cid using the tentativeBlockHeight"""
    # a lock on a project does not exist more than settings.webhook_listener.redis_lock_lifetime seconds.
    prev_root = None
    last_dag_cid = await helper_functions.get_dag_cid(
        project_id=project_id,
        block_height=tentative_block_height - 1,
        reader_redis_conn=reader_redis_conn
    )

    if not prev_cid_fetch:
        prev_root = last_dag_cid
        last_dag_cid = None
    """ Fill up the dag """
    dag = DAGBlock(
        height=tentative_block_height,
        prevCid={'/': last_dag_cid} if last_dag_cid else None,
        data={'cid': {'/': payload_cid}},
        txHash=tx_hash,
        timestamp=timestamp,
        prevRoot=prev_root
    )

    """ Convert dag structure to json and put it on ipfs dag """
    try:
        dag_cid = await put_dag_block(dag.json(exclude_none=True), ipfs_write_client)
    except Exception as e:
        logger.error("Failed to put dag block on ipfs: %s | Exception: %s", dag, e, exc_info=True)
        raise IPFSDAGCreationException from e
    else:
        logger.debug("DAG created: %s", dag)
    """ Update redis keys """
    block_height_key = redis_keys.get_block_height_key(project_id=project_id)
    _ = await writer_redis_conn.set(block_height_key, tentative_block_height)

    last_dag_cid_key = redis_keys.get_last_dag_cid_key(project_id)
    _ = await writer_redis_conn.set(last_dag_cid_key, dag_cid)

    _ = await writer_redis_conn.zadd(
        name=redis_keys.get_dag_cids_key(project_id),
        mapping={dag_cid: tentative_block_height}
    )

    return dag_cid, dag


async def discard_event(
        project_id: str,
        request_id: str,
        tentative_block_height: int,
        writer_redis_conn: aioredis.Redis
):
    redis_output = []
    d_r = await clear_payload_commit_data(
        project_id=project_id,
        tentative_height_pending_tx_entry=tentative_block_height,
        writer_redis_conn=writer_redis_conn
    )
    redis_output.extend(d_r)

    # Delete the payload cid from the list of payloadCids
    # out = await writer_redis_conn.zrem(
    #     key=redis_keys.get_payload_cids_key(project_id),
    #     member=payload_cid
    # )
    # redis_output.append(out)

    # Add the transaction Hash to discarded Transactions
    #TODO Do we need to record txHash or just requestID is sufficient??
    out = await writer_redis_conn.zadd(
        name=redis_keys.get_discarded_transactions_key(project_id),
        mapping={request_id: tentative_block_height}
    )
    redis_output.append(out)

    return redis_output


async def clear_payload_commit_data(
        project_id: str,
        tentative_height_pending_tx_entry: int,
        writer_redis_conn: aioredis.Redis
):
    """
    This function will be called once a dag block creation is successful to
    clear up all the transient, temporary redis keys associated with that
    particular dag block, since these key will not be needed anymore
    once the dag block has been created successfully.
        - Remove the tagged transaction hash entry from pendingTransactions set, by its tentative height score
        - Remove the payload_commit_id from pendingBlocks
    """
    deletion_result = []

    # remove tx_hash from list of pending transactions
    out = await writer_redis_conn.zremrangebyscore(
        name=redis_keys.get_pending_transactions_key(project_id=project_id),
        min=tentative_height_pending_tx_entry,
        max=tentative_height_pending_tx_entry
    )
    deletion_result.append(out)

    return deletion_result


async def clear_payload_commit_processing_logs(
        project_id: str,
        payload_commit_id: str,
        writer_redis_conn: aioredis.Redis
):
    _ = await writer_redis_conn.delete(redis_keys.get_payload_commit_id_process_logs_zset_key(
        project_id=project_id, payload_commit_id=payload_commit_id
    )
    )
    return _