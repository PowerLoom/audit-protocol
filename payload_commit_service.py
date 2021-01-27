import aioredis
from config import settings
import logging
import asyncio
import sys
import json
from eth_utils import keccak
from main import make_transaction
from maticvigil.EVCore import EVCore
from utils.redis_conn import provide_async_reader_conn_inst, provide_async_writer_conn_inst
import coloredlogs
from utils import redis_keys

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
        payload_commit_id: str,
        reader_redis_conn,
        writer_redis_conn,
        semaphore: asyncio.BoundedSemaphore,
):
    """
        - This function will take a pending payload and commit it to a smart contract
    """
    payload_commit_key = redis_keys.get_payload_commit_key(payload_commit_id=payload_commit_id)
    out = await reader_redis_conn.hgetall(payload_commit_key)
    if out:
        payload_data = {k.decode('utf-8'): v.decode('utf-8') for k, v in out.items()}
    else:
        payload_logger.warning("No payload data found in redis..")
        return -1

    payload_logger.debug(payload_data)
    snapshot = dict()
    snapshot['cid'] = payload_data['snapshotCid']
    snapshot['type'] = "HOT_IPFS"

    token_hash = '0x' + keccak(text=json.dumps(snapshot)).hex()
    _ = await semaphore.acquire()
    try:
        result = await make_transaction(
            snapshot_cid=payload_data['snapshotCid'],
            token_hash=token_hash,
            payload_commit_id=payload_commit_id,
            last_tentative_block_height=payload_data['tentativeBlockHeight'],
            project_id=payload_data['projectId'],
            writer_redis_conn=writer_redis_conn,
            contract=contract
        )
    except Exception as e:
        payload_logger.debug("There was an error while committing record: ")
        payload_logger.error(e, exc_info=True)
        result = -1
    finally:
        semaphore.release()

    if result == -1:
        payload_logger.warning("The payload commit was unsuccessful..")
        pending_payload_commits_key = redis_keys.get_pending_payload_commits_key()
        _ = await writer_redis_conn.lpush(pending_payload_commits_key, payload_commit_id)
        return -1
    else:
        payload_logger.debug("Successfully committed payload")

    return 1


@provide_async_reader_conn_inst
@provide_async_writer_conn_inst
async def commit_pending_payloads(
        reader_redis_conn=None,
        writer_redis_conn=None
):
    """
        - The goal of this function will be to check if there are any pending
        payloads left to commit, and take action on them
    """
    global contract

    payload_logger.debug("Checking for pending payloads to commit...")
    pending_payload_commits_key = redis_keys.get_pending_payload_commits_key()
    pending_payloads = await reader_redis_conn.lrange(pending_payload_commits_key, 0, -1)
    if len(pending_payloads) > 0:
        payload_logger.debug("Pending payloads found: ")
        payload_logger.debug(pending_payloads)
        tasks = []

        semaphore = asyncio.BoundedSemaphore(settings.max_payload_commits)

        """ Processing each of the pending payloads """
        while True:

            payload_commit_id = await writer_redis_conn.rpop(pending_payload_commits_key)
            if not payload_commit_id:
                break

            payload_commit_id = payload_commit_id.decode('utf-8')
            tasks.append(commit_single_payload(
                payload_commit_id=payload_commit_id,
                reader_redis_conn=reader_redis_conn,
                writer_redis_conn=writer_redis_conn,
                semaphore=semaphore
            ))

        await asyncio.gather(*tasks)
    else:
        payload_logger.debug("No pending payload's found...")


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
    while True:
        await asyncio.gather(
            commit_pending_payloads(),
            asyncio.sleep(settings.payload_commit_interval)
        )

if __name__ == "__main__":
    payload_logger.debug("Starting the loop")
    f = asyncio.ensure_future(periodic_commit_payload())
    f.add_done_callback(verifier_crash_cb)
    try:
        asyncio.get_event_loop().run_until_complete(asyncio.gather(f))
    except:
        asyncio.get_event_loop().stop()

