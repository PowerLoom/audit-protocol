from config import settings
from async_ipfshttpclient.main import AsyncIPFSClient
from utils import redis_keys
from utils import helper_functions
from data_models import PendingTransaction, DAGBlock, DAGFinalizerCallback, DAGBlockPayloadLinkedPath
from tenacity import retry, wait_random_exponential, retry_if_exception_type, stop_after_attempt
from typing import Tuple
from httpx import AsyncClient
import aiohttp
import async_timeout
import asyncio
import json
import io
import logging
from redis import asyncio as aioredis
import hmac
import sys
from utils.file_utils import write_bytes_to_file, read_text_file

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


async def send_commit_callback(httpx_session: AsyncClient, url, payload):
    if type(url) is bytes:
        url = url.decode('utf-8')
    resp = await httpx_session.post(url=url, json=payload)
    json_response = resp.json()
    return json_response


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

async def get_dag_block(dag_cid: str, project_id:str, ipfs_read_client: AsyncIPFSClient):
    e_obj = None
    dag = read_text_file(settings.local_cache_path + "/" + project_id + "/"+ dag_cid + ".json", logger )
    if dag is None:
        logger.info("Failed to read dag-block with CID %s for project %s from local cache ",
        dag_cid,project_id)
        try:
            async with async_timeout.timeout(settings.ipfs_timeout) as cm:
                try:
                    dag = await ipfs_read_client.dag.get(dag_cid)
                    dag = dag.as_json()
                except Exception as ex:
                    e_obj = ex
        except (asyncio.exceptions.CancelledError, asyncio.exceptions.TimeoutError) as err:
            e_obj = err


        if e_obj or cm.expired:
            return {}
    else:
        dag = json.loads(dag)
    return dag

async def put_dag_block(dag_json: str, project_id:str, ipfs_write_client: AsyncIPFSClient):
    dag_json = dag_json.encode('utf-8')
    out = await ipfs_write_client.dag.put(io.BytesIO(dag_json), pin=True)
    dag_cid = out['Cid']['/']
    try:
        write_bytes_to_file(settings.local_cache_path + "/" + project_id , "/" + dag_cid + ".json", dag_json, logger )
    except Exception as exc:
        logger.error("Failed to write dag-block %s for project %s to local cache due to exception %s",
        dag_json,project_id, exc, exc_info=True)

    return dag_cid


async def get_payload(payload_cid: str, project_id:str, ipfs_read_client: AsyncIPFSClient):
    """ Given the payload cid, retrieve the payload. Payloads are also DAG blocks """
    return await get_dag_block(payload_cid, project_id, ipfs_read_client)


# TODO: exception handling around dag block creation failures
async def create_dag_block_update_project_state(tx_hash, request_id, project_id, payload_commit_id,
                                                tentative_block_height_event_data, snapshot_cid, timestamp,
                                                reader_redis_conn, writer_redis_conn,
                                                fetch_prev_cid_for_dag_block_creation, parent_cid_height_diff,
                                                ipfs_write_client, httpx_client: AsyncClient,
                                                custom_logger_obj):
    _dag_cid, dag_block = await create_dag_block(
        tx_hash=tx_hash,
        project_id=project_id,
        tentative_block_height=tentative_block_height_event_data,
        payload_cid=snapshot_cid,
        timestamp=timestamp,
        reader_redis_conn=reader_redis_conn,
        writer_redis_conn=writer_redis_conn,
        prev_cid_fetch=fetch_prev_cid_for_dag_block_creation,
        parent_cid_height_diff=parent_cid_height_diff,
        ipfs_write_client=ipfs_write_client
    )
    custom_logger_obj.info('Created DAG block with CID %s at height %s', _dag_cid,
                           tentative_block_height_event_data)
    # clear from pending set
    _ = await writer_redis_conn.zremrangebyscore(
        name=redis_keys.get_pending_transactions_key(project_id),
        min=tentative_block_height_event_data,
        max=tentative_block_height_event_data
    )
    if _:
        custom_logger_obj.debug(
            'Cleared entry for requestID %s from pending set at block height %s | project ID %s',
            request_id, tentative_block_height_event_data, project_id
        )
    else:
        custom_logger_obj.debug(
            'Not sure if entry cleared for requestID %s from pending set at block height %s | '
            'Redis return: %s | Project ID : %s',
            request_id, tentative_block_height_event_data, _, project_id
        )
    # TODO: refactor DiffCalculationRequest as a timeseries indexing request for better on-the-fly higher order datasets
    # if dag_block.prevCid:
    #     diff_calculation_request = DiffCalculationRequest(
    #         project_id=project_id,
    #         dagCid=_dag_cid,
    #         lastDagCid=dag_block.prevCid['/'],
    #         payloadCid=pending_tx_obj.event_data.snapshotCid,
    #         txHash=pending_tx_obj.event_data.txHash,
    #         timestamp=pending_tx_obj.event_data.timestamp,
    #         tentative_block_height=_tt_block_height
    #     )
    #     async with app.rmq_channel_pool.acquire() as channel:
    #         # to save a call to rabbitmq. we already initialize exchanges and queues beforehand
    #         # always ensure exchanges and queues are initialized as part of launch sequence,
    #         # not to be checked here
    #         exchange = await channel.get_exchange(
    #             name=settings.rabbitmq.setup['core']['exchange'],
    #             ensure=False
    #         )
    #         message = Message(
    #             diff_calculation_request.json().encode('utf-8'),
    #             delivery_mode=DeliveryMode.PERSISTENT,
    #         )
    #         await exchange.publish(
    #             message=message,
    #             routing_key='diff-calculation'
    #         )
    #         custom_logger.debug(
    #             'Published diff calculation request | At height %s | Project %s',
    #             _tt_block_height, project_id
    #         )
    # else:
    #     custom_logger.debug(
    #         'No diff calculation request to publish for first block | At height %s | Project %s',
    #         _tt_block_height, project_id
    #     )
    # send commit ID confirmation callback
    # retrieve callback URL for project ID
    cb_url = await reader_redis_conn.get(f'powerloom:project:{project_id}:callbackURL')
    if cb_url:
        await send_commit_callback(httpx_session=httpx_client, url=cb_url, payload={
            'commitID': payload_commit_id,
            'projectID': project_id,
            'status': True
        })
    return _dag_cid, dag_block


@retry(
    reraise=True, wait=wait_random_exponential(multiplier=1, max=60),
    # Randomly wait up to 2^x * 1 seconds between each retry until the range reaches 60 seconds
    # then randomly up to 60 seconds afterwards
    retry=retry_if_exception_type(IPFSDAGCreationException),
    # stop=stop_after_attempt(5)  # redundant if we use wait_random_exponential()
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
        parent_cid_height_diff=1,
        prev_cid_fetch: bool = True
) -> Tuple[str, DAGBlock]:
    """ Get the last dag cid using the tentativeBlockHeight"""
    prev_root = None
    last_dag_cid = await helper_functions.get_dag_cid(
        project_id=project_id,
        block_height=tentative_block_height - parent_cid_height_diff,
        reader_redis_conn=reader_redis_conn
    )

    if not prev_cid_fetch:
        prev_root = last_dag_cid
        last_dag_cid = None
    """ Fill up the dag """
    if payload_cid:
        dag = DAGBlock(
            height=tentative_block_height,
            prevCid={'/': last_dag_cid} if last_dag_cid else None,
            # data={'cid': {'/': payload_cid}},
            data=DAGBlockPayloadLinkedPath(cid={'/': payload_cid}),
            txHash=tx_hash,
            timestamp=timestamp,
            prevRoot=prev_root
        )
    else:
        dag = DAGBlock(
            height=tentative_block_height,
            prevCid={'/': last_dag_cid} if last_dag_cid else None,
            data=None,
            txHash=tx_hash,
            timestamp=timestamp,
            prevRoot=prev_root
        )

    """ Convert dag structure to json and put it on ipfs dag """
    try:
        dag_cid = await put_dag_block(dag.json(exclude_none=True), project_id, ipfs_write_client)
    except Exception as e:
        logger.error("Failed to put dag block on ipfs: %s | Exception: %s", dag, e, exc_info=True)
        raise IPFSDAGCreationException from e
    else:
        logger.debug("DAG created: %s", dag)
    """ Update redis keys """
    block_height_key = redis_keys.get_block_height_key(project_id=project_id)
    _ = await writer_redis_conn.set(block_height_key, tentative_block_height)

    # FIXME: is last dag cid key necessary
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