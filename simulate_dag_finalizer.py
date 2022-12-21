from config import settings
from data_models import DAGFinalizerCallback, DAGFinalizerCBEventData, AuditRecordTxEventData, PendingTransaction
if settings.use_consensus:
    from snapshot_consensus.helpers.redis_keys import (
        get_project_registered_peers_set_key, get_epoch_submissions_htable_key
    )
    from snapshot_consensus.data_models import SubmissionDataStoreEntry, SnapshotSubmission
from utils.redis_conn import provide_redis_conn
from eth_utils import keccak
from uuid import uuid4
from utils import redis_keys
import time
import httpx
import random
import redis
import string
import logging
import sys
import coloredlogs


logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
formatter = logging.Formatter("%(levelname)-8s %(name)-4s %(asctime)s %(msecs)d %(module)s-%(funcName)s: %(message)s")
stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setFormatter(formatter)
stdout_handler.setLevel(logging.DEBUG)
logger.addHandler(stdout_handler)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setFormatter(formatter)
stderr_handler.setLevel(logging.ERROR)
logger.addHandler(stderr_handler)
logger.debug("Initialized logger")

coloredlogs.install(level="DEBUG", logger=logger, stream=sys.stdout)


@provide_redis_conn
def standalone_self_healing(redis_conn: redis.Redis):
    snapshot_cid = 'bafkreig3c2m4geyf3sf5nsfvbbgyy6p7c7ufatpfg4s3zpc7koqi5phsvq'  # to be used so we have a valid CID
    project_id = 'simulationRun'
    # initial clear
    for k in redis_conn.scan_iter(match='*simulationRun*', count=10):
        redis_conn.delete(k)
        logger.debug('Cleaned last run project state key %s', k)
    beginning_height = 1
    # put in pending tx entries to simulate payload commit to tx manager from 1 to num_blocks
    # last_sent_block = midway through num blocks
    # send finalization callbacks from 1 to (last_sent_block - 1)
    # set last touched block against pending tx entry at `last_sent_block` to -1
    # but finalized height is still (last_sent_block - 1). Simulates that DAG put at IPFS failed at `last_sent_block`
    #
    num_blocks = 20
    details = dict()
    for i in range(beginning_height, beginning_height+num_blocks):
        tx_hash = '0x'+keccak(text=''.join(random.choices(string.ascii_lowercase, k=5))).hex()
        pending_tx_entry = PendingTransaction(
            txHash=tx_hash,
            requestID=str(uuid4()),
            lastTouchedBlock=0,
            event_data=AuditRecordTxEventData(
                apiKeyHash='0x'+keccak(text=''.join(random.choices(string.ascii_lowercase, k=5))).hex(),
                tentativeBlockHeight=i,
                projectId=project_id,
                snapshotCid=snapshot_cid,
                payloadCommitId='0x'+keccak(text=''.join(random.choices(string.ascii_lowercase, k=5))).hex(),
                timestamp=int(time.time()),
                skipAnchorProof=True,
                txHash=tx_hash
            )
        )
        _ = redis_conn.zadd(
            name=redis_keys.get_pending_transactions_key(project_id),
            mapping={pending_tx_entry.json(): i}
        )
        if _:
            details[i] = pending_tx_entry
            logger.debug(
                'Added pending tx entry against height %s : %s', i, pending_tx_entry
            )
    last_sent_block = int(beginning_height+num_blocks/2)
    # send finalization call backs
    for i in range(beginning_height, last_sent_block):
        finalization_cb = DAGFinalizerCallback(
            txHash=details[i].txHash,
            requestID=details[i].requestID,
            event_data=DAGFinalizerCBEventData(
                apiKeyHash='0x'+keccak(text=''.join(random.choices(string.ascii_lowercase, k=5))).hex(),
                tentativeBlockHeight=i,
                projectId=project_id,
                snapshotCid=snapshot_cid,
                payloadCommitId='0x'+keccak(text=''.join(random.choices(string.ascii_lowercase, k=5))).hex(),
                timestamp=int(time.time())
            )
        )
        req_json = finalization_cb.dict()
        req_json.update({'event_name': 'RecordAppended'})
        r = httpx.post(
            url=f'http://{settings.webhook_listener.host}:{settings.webhook_listener.port}/',
            json=req_json
        )
        details[i].event_data = DAGFinalizerCBEventData.parse_obj(finalization_cb.event_data)
        if r.status_code == 200:
            logger.debug('Published callback to DAG finalizer at height %s : %s', i, finalization_cb)
        else:
            logger.error(
                'Failure publishing callback to DAG finalizer at height %s : %s | Response status: %s',
                i, finalization_cb, r.status_code
            )
            return
        logger.debug('Sleeping...')
        time.sleep(0.5)
    # set last touched block against pending tx entry at `last_sent_block` to -1
    # adapted from dag_utils.update_pending_tx_block_touch since it is an async function
    # first, remove
    redis_conn.zremrangebyscore(
        redis_keys.get_pending_transactions_key(project_id),
        min=last_sent_block,
        max=last_sent_block
    )
    # then, put in new entry
    new_pending_tx_set_entry_obj: PendingTransaction = details[last_sent_block]
    new_pending_tx_set_entry_obj.lastTouchedBlock = -1
    new_pending_tx_set_entry_obj.event_data = AuditRecordTxEventData(
        txHash=new_pending_tx_set_entry_obj.txHash,
        projectId=project_id,
        apiKeyHash='0x' + keccak(text=''.join(random.choices(string.ascii_lowercase, k=5))).hex(),
        timestamp=int(time.time()),
        payloadCommitId='0x' + keccak(text=''.join(random.choices(string.ascii_lowercase, k=5))).hex(),
        snapshotCid=snapshot_cid,
        tentativeBlockHeight=last_sent_block
    )
    _ = redis_conn.zadd(
        name=redis_keys.get_pending_transactions_key(project_id),
        mapping={new_pending_tx_set_entry_obj.json(): last_sent_block}
    )
    if _:
        logger.info(
            'Updated pending tx entry at height %s so that last touched block = -1 : %s',
            last_sent_block, new_pending_tx_set_entry_obj
        )
    else:
        logger.info(
            'Could not update pending tx entry at height %s so that last touched block = -1 : %s',
            last_sent_block, new_pending_tx_set_entry_obj
        )
    # resume sending callbacks from last sent block+1 to end
    for i in range(last_sent_block+1, beginning_height+num_blocks):
        finalization_cb = DAGFinalizerCallback(
            txHash=details[i].txHash,
            requestID=details[i].requestID,
            event_data=DAGFinalizerCBEventData(
                apiKeyHash='0x' + keccak(text=''.join(random.choices(string.ascii_lowercase, k=5))).hex(),
                tentativeBlockHeight=i,
                projectId=project_id,
                snapshotCid=snapshot_cid,
                payloadCommitId='0x' + keccak(text=''.join(random.choices(string.ascii_lowercase, k=5))).hex(),
                timestamp=int(time.time())
            )
        )
        req_json = finalization_cb.dict()
        req_json.update({'event_name': 'RecordAppended'})
        r = httpx.post(
            url=f'http://{settings.webhook_listener.host}:{settings.webhook_listener.port}/',
            json=req_json
        )
        details[i].event_data = DAGFinalizerCBEventData.parse_obj(finalization_cb.event_data)
        if r.status_code == 200:
            logger.debug('Published callback to DAG finalizer at height %s : %s', i, finalization_cb)
        else:
            logger.error(
                'Failure publishing callback to DAG finalizer at height %s : %s | Response status: %s',
                i, finalization_cb, r.status_code
            )
            return
        logger.debug('Sleeping...')
        time.sleep(1)


def register_submission(project_id, epoch_end, peer_id, snapshot_cid, redis_conn: redis.Redis):
    _ = redis_conn.hset(
        name=get_epoch_submissions_htable_key(
            project_id=project_id,
            epoch_end=epoch_end
        ),
        key=peer_id,
        value=SubmissionDataStoreEntry(snapshotCID=snapshot_cid, submittedTS=int(time.time())).json()
    )
    if redis_conn.ttl(name=get_epoch_submissions_htable_key(
            project_id=project_id,
            epoch_end=epoch_end,
    )) == -1:
        redis_conn.expire(
            name=get_epoch_submissions_htable_key(
                project_id=project_id,
                epoch_end=epoch_end,
            ),
            time=3600
        )
    return _


@provide_redis_conn
def consensus_self_healing(redis_conn: redis.Redis):
    snapshot_cid = 'bafkreig3c2m4geyf3sf5nsfvbbgyy6p7c7ufatpfg4s3zpc7koqi5phsvq'  # to be used so we have a valid CID
    project_id = 'consensusSimulationRun'
    # initial clear
    for k in redis_conn.scan_iter(match='*consensusSimulationRun*', count=10):
        redis_conn.delete(k)
        logger.debug('Cleaned last run project state key %s', k)
    # add accepted peers
    peers = ['peer1', 'peer2', 'peer3']
    _ = redis_conn.sadd(
        get_project_registered_peers_set_key(project_id),
        *peers
    )
    if _ == 3:
        logger.debug('Set project %s accepted peers to %s', project_id, peers)
    else:
        logger.warning('Could not set project %s accepted peers to %s', project_id, peers)
        return
    beginning_height = 1
    # set epoch size and first end height for project
    project_epoch_size = 10
    _ = redis_conn.set(redis_keys.get_project_epoch_size(project_id), str(project_epoch_size))
    if _:
        logger.debug('Set project %s epoch size to %s', project_id, project_epoch_size)
    else:
        logger.warning('Could not set project %s epoch size to %s', project_id, project_epoch_size)
        return
    # first epoch is begin =1 to end=10
    first_epoch_end_height = 10
    _ = redis_conn.set(redis_keys.get_project_first_epoch_end_height(project_id), str(first_epoch_end_height))
    if _:
        logger.debug('Set project %s first epoch end height to %s', project_id, project_epoch_size)
    else:
        logger.warning('Could not set project %s first epoch end height to %s', project_id, project_epoch_size)
        return
    # put in pending tx entries to simulate payload commit to tx manager from 1 to num_blocks
    # last_sent_block = midway through num blocks
    # send finalization callbacks from 1 to (last_sent_block - 1)
    # set last touched block against pending tx entry at `last_sent_block` to -1
    # but finalized height is still (last_sent_block - 1). Simulates that DAG put at IPFS failed at `last_sent_block`
    #
    num_blocks = 20
    details = dict()
    for i in range(beginning_height, beginning_height + num_blocks):
        tx_hash = '0x' + keccak(text=''.join(random.choices(string.ascii_lowercase, k=5))).hex()
        pending_tx_entry = PendingTransaction(
            txHash=tx_hash,
            requestID=str(uuid4()),
            lastTouchedBlock=0,
            event_data=AuditRecordTxEventData(
                apiKeyHash='0x' + keccak(text=''.join(random.choices(string.ascii_lowercase, k=5))).hex(),
                tentativeBlockHeight=i,
                projectId=project_id,
                snapshotCid=snapshot_cid,
                payloadCommitId='0x' + keccak(text=''.join(random.choices(string.ascii_lowercase, k=5))).hex(),
                timestamp=int(time.time()),
                skipAnchorProof=True,
                txHash=tx_hash
            )
        )
        _ = redis_conn.zadd(
            name=redis_keys.get_pending_transactions_key(project_id),
            mapping={pending_tx_entry.json(): i}
        )
        if _:
            details[i] = pending_tx_entry
            logger.debug(
                'Added pending tx entry against height %s : %s', i, pending_tx_entry
            )
    last_sent_block = int(beginning_height + num_blocks / 2)
    # send finalization call backs
    for i in range(beginning_height, last_sent_block):
        finalization_cb = DAGFinalizerCallback(
            txHash=details[i].txHash,
            requestID=details[i].requestID,
            event_data=DAGFinalizerCBEventData(
                apiKeyHash='0x' + keccak(text=''.join(random.choices(string.ascii_lowercase, k=5))).hex(),
                tentativeBlockHeight=i,
                projectId=project_id,
                snapshotCid=snapshot_cid,
                payloadCommitId='0x' + keccak(text=''.join(random.choices(string.ascii_lowercase, k=5))).hex(),
                timestamp=int(time.time())
            )
        )
        req_json = finalization_cb.dict()
        req_json.update({'event_name': 'RecordAppended'})
        r = httpx.post(
            url=f'http://{settings.webhook_listener.host}:{settings.webhook_listener.port}/',
            json=req_json
        )
        details[i].event_data = DAGFinalizerCBEventData.parse_obj(finalization_cb.event_data)
        if r.status_code == 200:
            logger.debug('Published callback to DAG finalizer at height %s : %s', i, finalization_cb)
        else:
            logger.error(
                'Failure publishing callback to DAG finalizer at height %s : %s | Response status: %s',
                i, finalization_cb, r.status_code
            )
            return
        logger.debug('Sleeping...')
        time.sleep(0.5)
    # remove pending tx entry at `last_sent_block`
    redis_conn.zremrangebyscore(
        redis_keys.get_pending_transactions_key(project_id),
        min=last_sent_block,
        max=last_sent_block
    )
    expected_epoch_end_at_last_sent_block = (last_sent_block-1) * project_epoch_size + first_epoch_end_height
    # then, add consensus epoch at `last_sent_block`
    for peer in peers:
        _ = register_submission(
            project_id=project_id,
            epoch_end=expected_epoch_end_at_last_sent_block,
            peer_id=peer,
            snapshot_cid=snapshot_cid,
            redis_conn=redis_conn
        )
        if _:
            logger.info(
                'Registered submission at height %s, epoch end for peer %s',
                last_sent_block, expected_epoch_end_at_last_sent_block, peer
            )
        else:
            logger.info(
                'Could not register submission at height %s, epoch end for peer %s',
                last_sent_block, expected_epoch_end_at_last_sent_block, peer
            )
            return
    # resume sending callbacks from last sent block+1 to end
    for i in range(last_sent_block + 1, beginning_height + num_blocks):
        finalization_cb = DAGFinalizerCallback(
            txHash=details[i].txHash,
            requestID=details[i].requestID,
            event_data=DAGFinalizerCBEventData(
                apiKeyHash='0x' + keccak(text=''.join(random.choices(string.ascii_lowercase, k=5))).hex(),
                tentativeBlockHeight=i,
                projectId=project_id,
                snapshotCid=snapshot_cid,
                payloadCommitId='0x' + keccak(text=''.join(random.choices(string.ascii_lowercase, k=5))).hex(),
                timestamp=int(time.time())
            )
        )
        req_json = finalization_cb.dict()
        req_json.update({'event_name': 'RecordAppended'})
        r = httpx.post(
            url=f'http://{settings.webhook_listener.host}:{settings.webhook_listener.port}/',
            json=req_json
        )
        details[i].event_data = DAGFinalizerCBEventData.parse_obj(finalization_cb.event_data)
        if r.status_code == 200:
            logger.debug('Published callback to DAG finalizer at height %s : %s', i, finalization_cb)
        else:
            logger.error(
                'Failure publishing callback to DAG finalizer at height %s : %s | Response status: %s',
                i, finalization_cb, r.status_code
            )
            return
        logger.debug('Sleeping...')
        time.sleep(1)


if __name__ == '__main__':
    standalone_self_healing()
