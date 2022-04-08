from config import settings
from data_models import PayloadCommit, PendingTransaction
from utils.ipfs_async import client as ipfs_client
from eth_utils import keccak
from maticvigil.EVCore import EVCore
from aio_pika.pool import Pool
from aio_pika import IncomingMessage
from utils.redis_conn import RedisPool, get_redis_conn_from_pool, REDIS_CONN_CONF
from utils import redis_keys
from utils.rabbitmq_utils import get_rabbitmq_channel, get_rabbitmq_connection
from functools import partial
from maticvigil.exceptions import EVBaseException
from tenacity import AsyncRetrying, Retrying, stop_after_attempt, wait_random
from greenletio import async_
import uvloop
import signal
import aiojobs
import aio_pika
import aioredis
import time
import logging
import asyncio
import sys
import json
import requests

formatter = logging.Formatter(u"%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s")

stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setLevel(logging.DEBUG)
stdout_handler.setFormatter(formatter)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setLevel(logging.ERROR)
stderr_handler.setFormatter(formatter)
payload_logger = logging.getLogger(__name__)
payload_logger.setLevel(logging.DEBUG)
payload_logger.addHandler(stdout_handler)
payload_logger.addHandler(stderr_handler)
# coloredlogs.install(level="DEBUG", logger=payload_logger, stream=sys.stdout)

payload_logger.debug("Starting Payload Commit Service...")

asyncio.set_event_loop_policy(uvloop.EventLoopPolicy())


async def make_transaction(
        snapshot_cid,
        payload_commit_id,
        token_hash,
        last_tentative_block_height,
        project_id,
        writer_redis_conn: aioredis.Redis,
        resubmission_block=0
):
    """
        - Create a unique transaction_id associated with this transaction,
        and add it to the set of pending transactions
    """

    e_obj = None
    try:
        loop = asyncio.get_running_loop()
    except Exception as e:
        payload_logger.warning("There was an error while trying to get event loop")
        payload_logger.error(e, exc_info=True)
        return None
    kwargs = dict(
        payloadCommitId=payload_commit_id,
        snapshotCid=snapshot_cid,
        apiKeyHash=token_hash,
        projectId=project_id,
        tentativeBlockHeight=last_tentative_block_height,
    )
    # partial_func = partial(audit_record_store_contract.commitRecord, **kwargs)
    await TX_SENDER_LIMITING_SEMAPHORE.acquire()
    try:
        # try 3 times, wait 1-2 seconds between retries
        async for attempt in AsyncRetrying(reraise=True, stop=stop_after_attempt(3), wait=wait_random(1, 2)):
            with attempt:
                tx_hash_obj = await audit_record_store_contract.commitRecord(**kwargs)
                if tx_hash_obj:
                    break
    except EVBaseException as evbase:
        e_obj = evbase
    except requests.exceptions.HTTPError as errh:
        e_obj = errh
    except requests.exceptions.ConnectionError as errc:
        e_obj = errc
    except requests.exceptions.Timeout as errt:
        e_obj = errt
    except requests.exceptions.RequestException as errr:
        e_obj = errr
    except Exception as e:
        e_obj = e
    else:
        # right away update pending transactions
        pending_transaction_key = redis_keys.get_pending_transactions_key(project_id)
        tx_hash = tx_hash_obj[0]['txHash']
        event_data = {
            'txHash': tx_hash,
            'projectId': project_id,
            'payloadCommitId': payload_commit_id,
            'timestamp': int(time.time()),  # will be updated to actual blockchain timestamp once callback arrives
            'snapshotCid': snapshot_cid,
            'apiKeyHash': token_hash,
            'tentativeBlockHeight': last_tentative_block_height
        }
        tx_hash_pending_entry = PendingTransaction(
            txHash=tx_hash,
            lastTouchedBlock=resubmission_block,
            event_data=event_data
        )
        _ = await writer_redis_conn.zadd(
            key=pending_transaction_key,
            member=tx_hash_pending_entry.json(),
            score=last_tentative_block_height
        )
        payload_logger.info(
            "Successful transaction %s committed to AuditRecord contract against payload commit ID %s",
            tx_hash_obj, payload_commit_id
        )
        await writer_redis_conn.zadd(
            key=redis_keys.get_payload_commit_id_process_logs_zset_key(
                project_id=project_id, payload_commit_id=payload_commit_id
            ),
            member=json.dumps(
                {
                    'worker': 'payload_commit_service',
                    'update': {
                        'action': 'AuditRecord.Commit',
                        'info': {
                            'msg': kwargs,
                            'status': 'Success',
                            'txHash': tx_hash_obj
                        }
                    }
                }
            ),
            score=int(time.time())
        )
    finally:
        TX_SENDER_LIMITING_SEMAPHORE.release()

    if e_obj:
        payload_logger.info(
            "=" * 80 +
            "Commit Payload to AuditRecord contract failed. Tx was not successful against commit ID %s\n%s" +
            "=" * 80,
            payload_commit_id, e_obj
        )
        await writer_redis_conn.zadd(
            key=redis_keys.get_payload_commit_id_process_logs_zset_key(
                project_id=project_id, payload_commit_id=payload_commit_id
            ),
            member=json.dumps(
                {
                    'worker': 'payload_commit_service',
                    'update': {
                        'action': 'AuditRecord.Commit',
                        'info': {
                            'msg': kwargs,
                            'status': 'Failed',
                            'exception': e_obj.__repr__()
                        }
                    }
                }
            ),
            score=int(time.time())
        )
        return None


    return tx_hash


async def single_payload_commit():
    aioredis_pool = RedisPool(5)
    await aioredis_pool.populate()
    while True:
        msg_body = await PAYLOAD_COMMIT_REQUESTS_QUEUE_INTERNAL.get()
        # semaphore: asyncio.BoundedSemaphore
        """
            - This function will take a pending payload and commit it to a smart contract
        """
        writer_redis_conn: aioredis.Redis = aioredis_pool.writer_redis_pool
        payload_commit_obj: PayloadCommit = PayloadCommit.parse_raw(msg_body)
        payload_commit_id = payload_commit_obj.commitId
        # payload_logger.debug(payload_data)
        snapshot = dict()
        core_payload = payload_commit_obj.payload
        project_id = payload_commit_obj.projectId
        snapshot_cid = None
        if core_payload:
            try:
                # try 3 times, wait 1-2 seconds between retries
                async for attempt in AsyncRetrying(reraise=True, stop=stop_after_attempt(3), wait=wait_random(1, 2)):
                    with attempt:
                        if type(core_payload) is dict:
                            snapshot_cid = await ipfs_client.add_json(core_payload)
                        else:
                            try:
                                core_payload = json.dumps(core_payload)
                            except:
                                pass
                            snapshot_cid = await ipfs_client.add_str(str(core_payload))
                        if snapshot_cid:
                            break
            except Exception as e:
                await writer_redis_conn.zadd(
                    key=redis_keys.get_payload_commit_id_process_logs_zset_key(
                        project_id=project_id, payload_commit_id=payload_commit_id
                    ),
                    member=json.dumps(
                        {
                            'worker': 'payload_commit_service',
                            'update': {
                                'action': 'IPFS.Commit',
                                'info': {
                                    'core_payload': core_payload,
                                    'status': 'Failed',
                                    'exception': e.__repr__()
                                }
                            }
                        }
                    ),
                    score=int(time.time())
                )
                return
            else:
                if snapshot_cid:
                    await writer_redis_conn.zadd(
                        key=redis_keys.get_payload_commit_id_process_logs_zset_key(
                            project_id=project_id, payload_commit_id=payload_commit_id
                        ),
                        member=json.dumps(
                            {
                                'worker': 'payload_commit_service',
                                'update': {
                                    'action': 'IPFS.Commit',
                                    'info': {
                                        'core_payload': core_payload,
                                        'status': 'Success',
                                        'CID': snapshot_cid
                                    }
                                }
                            }
                        ),
                        score=int(time.time())
                    )
        else:
            snapshot_cid = payload_commit_obj.snapshotCID
        payload_cid_key = redis_keys.get_payload_cids_key(project_id)
        _ = await writer_redis_conn.zadd(
            key=payload_cid_key,
            score=int(payload_commit_obj.tentativeBlockHeight),
            member=snapshot_cid
        )
        snapshot['cid'] = snapshot_cid
        snapshot['type'] = "HOT_IPFS"
        if not payload_commit_obj.apiKeyHash:
            token_hash = '0x' + keccak(text=json.dumps(snapshot)).hex()
        else:
            token_hash = payload_commit_obj.apiKeyHash
        result_tx_hash = await make_transaction(
            snapshot_cid=snapshot_cid,
            token_hash=token_hash,
            payload_commit_id=payload_commit_id,
            last_tentative_block_height=payload_commit_obj.tentativeBlockHeight,
            project_id=payload_commit_obj.projectId,
            resubmission_block=payload_commit_obj.resubmissionBlock,
            writer_redis_conn=writer_redis_conn
        )
        if result_tx_hash:
            if not payload_commit_obj.resubmitted:
                last_snapshot_cid_key = redis_keys.get_last_snapshot_cid_key(project_id)
                payload_logger.debug("Setting the last snapshot_cid as %s for project ID %s", snapshot_cid, project_id)
                _ = await writer_redis_conn.set(last_snapshot_cid_key, snapshot_cid)
            payload_logger.debug('Setting tx hash %s against payload commit ID %s', result_tx_hash, payload_commit_id)
        PAYLOAD_COMMIT_REQUESTS_QUEUE_INTERNAL.task_done()


async def commit_message_cb(
    message: IncomingMessage
):
    async with message.process():
        await PAYLOAD_COMMIT_REQUESTS_QUEUE_INTERNAL.put(message.body)
        msg_json = json.loads(message.body)
        payload_logger.debug('Sent incoming payload commit message at tentative DAG height %s'
                             'for project %s in chain height range %s to internal processing queue',
                             msg_json['tentativeBlockHeight'], msg_json['projectId'],
                             msg_json['payload']['chainHeightRange']
                             )


async def payload_commit_queue_listener(loop: asyncio.AbstractEventLoop):
    rmq_connection_pool = Pool(get_rabbitmq_connection, max_size=5, loop=loop)
    rmq_channel_pool = Pool(
        partial(get_rabbitmq_channel, rmq_connection_pool), max_size=20, loop=loop
    )
    async with rmq_channel_pool.acquire() as channel:
        await channel.set_qos(20)
        audit_protocol_backend_exchange = await channel.declare_exchange(
            settings.rabbitmq.setup['core']['exchange'], aio_pika.ExchangeType.DIRECT
        )
        # Declaring log request receiver queue and bind to exchange
        payload_commit_queue = 'audit-protocol-commit-payloads'
        routing_key = 'commit-payloads'
        receiving_queue = await channel.declare_queue(name=payload_commit_queue, durable=True, auto_delete=False)
        await receiving_queue.bind(audit_protocol_backend_exchange, routing_key=routing_key)
        payload_logger.debug(f'Consuming payload commit queue %s with routing key %s...', payload_commit_queue, routing_key)
        await receiving_queue.consume(commit_message_cb)


def payload_commit_q_listener_crash_cb(fut: asyncio.Future):
    try:
        exc = fut.exception()
    except (asyncio.CancelledError, aioredis.ConnectionClosedError):
        payload_logger.error('Respawning RabbitMQ payload commit queue listener...')
        t = asyncio.ensure_future(payload_commit_queue_listener(asyncio.get_running_loop()))
        t.add_done_callback(payload_commit_q_listener_crash_cb)
    except Exception as e:
        payload_logger.error('RabbitMQ payload commit queue listener crashed')
        payload_logger.error(e, exc_info=True)


def queue_dispatcher_crash_cb(fut: asyncio.Future):
    try:
        exc = fut.exception()
    except (asyncio.CancelledError, aioredis.ConnectionClosedError):
        payload_logger.error('Respawning internal payload commit queue job dispatcher...')
        t = asyncio.ensure_future(single_payload_commit())
        t.add_done_callback(payload_commit_q_listener_crash_cb)
    except Exception as e:
        payload_logger.error('internal payload commit queue listener crashed')
        payload_logger.error(e, exc_info=True)


async def shutdown(signal, loop):
    logging.info(f'Received exit signal {signal.name}...')
    tasks = [t for t in asyncio.tasks.all_tasks(loop) if t is not
             asyncio.tasks.current_task(loop)]

    [task.cancel() for task in tasks]

    logging.info(f'Cancelling {len(tasks)} outstanding tasks')
    await asyncio.gather(*tasks)
    logging.info('Shutdown complete.')


if __name__ == "__main__":
    ev_loop = asyncio.get_event_loop()
    signals = (signal.SIGTERM, signal.SIGINT)
    for s in signals:
        ev_loop.add_signal_handler(
            s, lambda x=s: ev_loop.create_task(shutdown(x, ev_loop)))
    PAYLOAD_COMMIT_REQUESTS_QUEUE_INTERNAL = asyncio.Queue(maxsize=10000)
    TX_SENDER_LIMITING_SEMAPHORE = asyncio.BoundedSemaphore(settings.max_payload_commits)
    evc = EVCore(verbose=True)
    audit_record_store_contract = evc.generate_contract_sdk(
        contract_address=settings.audit_contract,
        app_name='auditrecords'
    )

    class AsyncContractSDK:
        def __init__(self, commitRecord):
            self.commitRecord = commitRecord

        async def async_commit_record(self, *args, **kwargs):
            return await async_(self.commitRecord)(*args, **kwargs)

    async_contract = AsyncContractSDK(audit_record_store_contract.commitRecord)
    # monkey patch
    audit_record_store_contract.commitRecord = async_contract.async_commit_record

    payload_logger.debug("Starting RabbitMQ payload commit queue listener...")
    f = asyncio.ensure_future(payload_commit_queue_listener(ev_loop))
    f.add_done_callback(payload_commit_q_listener_crash_cb)
    payload_logger.debug("Started RabbitMQ payload commit queue listener...")
    for k in range(100):
        f2 = asyncio.ensure_future(single_payload_commit())
        f2.add_done_callback(queue_dispatcher_crash_cb)
    try:
        ev_loop.run_forever()
    except:
        asyncio.get_event_loop().stop()

