from config import settings
from data_models import PayloadCommit
from utils.ipfs_async import client as ipfs_client
from eth_utils import keccak
from main import make_transaction
from maticvigil.EVCore import EVCore
from aio_pika.pool import Pool
from aio_pika import IncomingMessage
from utils.redis_conn import RedisPool
from utils import redis_keys
from utils.rabbitmq_utils import get_rabbitmq_channel, get_rabbitmq_connection
from functools import partial
from tenacity import AsyncRetrying, stop_after_attempt, wait_random
import aio_pika
import aioredis
import time
import logging
import asyncio
import sys
import json

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

payload_logger.debug("Initialized Payload")

evc = EVCore(verbose=True)
contract = evc.generate_contract_sdk(
    contract_address=settings.audit_contract,
    app_name='auditrecords'
)


async def commit_single_payload(
    message: IncomingMessage,
    aioredis_pool: RedisPool
):
    async with message.process():
        # semaphore: asyncio.BoundedSemaphore
        """
            - This function will take a pending payload and commit it to a smart contract
        """
        writer_redis_conn: aioredis.Redis = aioredis_pool.writer_redis_pool
        payload_commit_obj: PayloadCommit = PayloadCommit.parse_raw(message.body)
        payload_commit_id = payload_commit_obj.commitId
        # payload_logger.debug(payload_data)
        snapshot = dict()
        core_payload = payload_commit_obj.payload
        if core_payload:
            project_id = payload_commit_obj.projectId
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
                # await writer_redis_conn.zadd(
                #     key=redis_keys.get_payload_commit_id_process_logs_zset_key(
                #         project_id=project_id, payload_commit_id=payload_commit_id
                #     ),
                #     member=json.dumps(
                #         {
                #             'worker': 'payload_commit_service',
                #             'update': {
                #                 'action': 'IPFS.Commit',
                #                 'info': {
                #                     'core_payload': core_payload,
                #                     'status': 'Failed',
                #                     'exception': e.__repr__()
                #                 }
                #             }
                #         }
                #     ),
                #     score=int(time.time())
                # )
                return
            # else:
                # await writer_redis_conn.zadd(
                #     key=redis_keys.get_payload_commit_id_process_logs_zset_key(
                #         project_id=project_id, payload_commit_id=payload_commit_id
                #     ),
                #     member=json.dumps(
                #         {
                #             'worker': 'payload_commit_service',
                #             'update': {
                #                 'action': 'IPFS.Commit',
                #                 'info': {
                #                     'core_payload': core_payload,
                #                     'status': 'Success',
                #                     'exception': snapshot_cid
                #                 }
                #             }
                #         }
                #     ),
                #     score=int(time.time())
                # )
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
        _ = await semaphore.acquire()
        result_tx_hash = await make_transaction(
            snapshot_cid=snapshot_cid,
            token_hash=token_hash,
            payload_commit_id=payload_commit_id,
            last_tentative_block_height=payload_commit_obj.tentativeBlockHeight,
            project_id=payload_commit_obj.projectId,
            writer_redis_conn=writer_redis_conn,
            contract=contract
        )
        if result_tx_hash:
            if not payload_commit_obj.resubmitted:
                last_snapshot_cid_key = redis_keys.get_last_snapshot_cid_key(project_id)
                payload_logger.debug("Setting the last snapshot_cid as %s for project ID %s", snapshot_cid, project_id)
                _ = await writer_redis_conn.set(last_snapshot_cid_key, snapshot_cid)
            payload_logger.debug('Setting tx hash %s against payload commit ID %s', result_tx_hash, payload_commit_id)
            await writer_redis_conn.set(redis_keys.get_payload_commit_key(payload_commit_id), result_tx_hash)
        semaphore.release()


async def periodic_commit_payload(loop: asyncio.AbstractEventLoop):
    aioredis_pool = RedisPool()
    await aioredis_pool.populate()
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
        commit_payload_queue = 'audit-protocol-commit-payloads'
        routing_key = 'commit-payloads'
        receiving_queue = await channel.declare_queue(name=commit_payload_queue, durable=True, auto_delete=False)
        await receiving_queue.bind(audit_protocol_backend_exchange, routing_key=routing_key)
        payload_logger.debug(f'Consuming payload commit queue %s with routing key %s...', commit_payload_queue, routing_key)
        consumer_callable = partial(commit_single_payload, **{'aioredis_pool': aioredis_pool})
        await receiving_queue.consume(consumer_callable)


def verifier_crash_cb(fut: asyncio.Future):
    try:
        exc = fut.exception()
    except (asyncio.CancelledError, aioredis.ConnectionClosedError):
        payload_logger.error('Respawning commit payload task...')
        t = asyncio.ensure_future(periodic_commit_payload())
        t.add_done_callback(verifier_crash_cb)
    except Exception as e:
        payload_logger.error('Commit payload task crashed')
        payload_logger.error(e, exc_info=True)


if __name__ == "__main__":
    ev_loop = asyncio.get_event_loop()
    semaphore = asyncio.BoundedSemaphore(settings.max_payload_commits)
    payload_logger.debug("Starting the loop")
    f = asyncio.ensure_future(periodic_commit_payload(ev_loop))
    f.add_done_callback(verifier_crash_cb)
    try:
        ev_loop.run_forever()
    except:
        asyncio.get_event_loop().stop()

