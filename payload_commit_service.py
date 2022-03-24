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
import aio_pika
import aioredis
import coloredlogs
import logging
import asyncio
import sys
import json

formatter = logging.Formatter(u"%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s")

stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setLevel(logging.DEBUG)
# stdout_handler.setFormatter(formatter)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setLevel(logging.ERROR)
# stderr_handler.setFormatter(formatter)
payload_logger = logging.getLogger(__name__)
payload_logger.setLevel(logging.DEBUG)
payload_logger.addHandler(stdout_handler)
payload_logger.addHandler(stderr_handler)
coloredlogs.install(level="DEBUG", logger=payload_logger, stream=sys.stdout)

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
    # semaphore: asyncio.BoundedSemaphore
    """
        - This function will take a pending payload and commit it to a smart contract
    """
    writer_redis_conn = aioredis_pool.writer_redis_pool
    payload_commit_obj: PayloadCommit = PayloadCommit.parse_raw(message.body)
    payload_commit_id = payload_commit_obj.commitId
    # payload_logger.debug(payload_data)
    snapshot = dict()
    core_payload = payload_commit_obj.payload
    project_id = payload_commit_obj.projectId
    if type(core_payload) is dict:
        snapshot_cid = await ipfs_client.add_json(core_payload)
    else:
        try:
            core_payload = json.dumps(core_payload)
        except:
            pass
        snapshot_cid = await ipfs_client.add_str(str(core_payload))

    payload_cid_key = redis_keys.get_payload_cids_key(project_id)
    _ = await writer_redis_conn.zadd(
        key=payload_cid_key,
        score=int(payload_commit_obj.tentativeBlockHeight),
        member=snapshot_cid
    )
    snapshot['cid'] = snapshot_cid
    snapshot['type'] = "HOT_IPFS"

    token_hash = '0x' + keccak(text=json.dumps(snapshot)).hex()
    _ = await semaphore.acquire()
    try:
        result = await make_transaction(
            snapshot_cid=snapshot_cid,
            token_hash=token_hash,
            payload_commit_id=payload_commit_id,
            last_tentative_block_height=payload_commit_obj.tentativeBlockHeight,
            project_id=payload_commit_obj.projectId,
            writer_redis_conn=writer_redis_conn,
            contract=contract
        )
    except Exception as e:
        payload_logger.debug("There was an error while committing record: ")
        payload_logger.error(e, exc_info=True)
        result = -1
    else:
        last_snapshot_cid_key = redis_keys.get_last_snapshot_cid_key(project_id)
        payload_logger.debug("Setting the last snapshot_cid as: ")
        payload_logger.debug(snapshot_cid)
        _ = await writer_redis_conn.set(last_snapshot_cid_key, snapshot_cid)
    finally:
        semaphore.release()

    if result == -1:
        payload_logger.warning("The payload commit was unsuccessful..")
        return -1
    else:
        payload_logger.debug("Successfully committed payload")
    # TODO: remove all abominations like this that return arbitrary integers that have to be made sense of later on
    return 1


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


async def periodic_commit_payload():
    aioredis_pool = RedisPool()
    await aioredis_pool.populate()
    rmq_connection_pool = Pool(get_rabbitmq_connection, max_size=5, loop=asyncio.get_running_loop())
    rmq_channel_pool = Pool(
        partial(get_rabbitmq_channel, rmq_connection_pool), max_size=20, loop=asyncio.get_running_loop()
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
        payload_logger.debug(f'Consuming payload commit queue %s with routing key %s...')
        await receiving_queue.consume(commit_single_payload, arguments={'aioredis_pool': aioredis_pool})


if __name__ == "__main__":
    semaphore = asyncio.BoundedSemaphore(settings.max_payload_commits)
    payload_logger.debug("Starting the loop")
    f = asyncio.ensure_future(periodic_commit_payload())
    f.add_done_callback(verifier_crash_cb)
    try:
        asyncio.get_event_loop().run_until_complete(asyncio.gather(f))
    except:
        asyncio.get_event_loop().stop()

