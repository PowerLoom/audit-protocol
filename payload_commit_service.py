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
import aio_pika
import aioredis
import time
import logging
import asyncio
import sys
import json
import requests
import redis
import ray

ray.init(include_dashboard=True, dashboard_port=8085, ignore_reinit_error=True, logging_level=logging.INFO)

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


@ray.remote
class TransactionActorAsync:
    def __init__(self):
        self.aioredis_pool = RedisPool(pool_size=10)
        formatter = logging.Formatter(
            u"%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s")
        stdout_handler = logging.StreamHandler(sys.stdout)
        stdout_handler.setLevel(logging.DEBUG)
        stdout_handler.setFormatter(formatter)

        stderr_handler = logging.StreamHandler(sys.stderr)
        stderr_handler.setLevel(logging.ERROR)
        stderr_handler.setFormatter(formatter)
        logging.basicConfig(level=logging.INFO)
        self.contract = None

    def _init_sdk(self):
        if not self.contract:
            self.evc = EVCore(verbose=True)
            self.contract = self.evc.generate_contract_sdk(
                contract_address=settings.audit_contract,
                app_name='auditrecords'
            )

    async def make_transaction(
            self, snapshot_cid, payload_commit_id, token_hash, last_tentative_block_height, project_id, resubmission_block=0
    ):
        """
            - Create a unique transaction_id associated with this transaction,
            and add it to the set of pending transactions
        """
        self._init_sdk()
        await self.aioredis_pool.populate()
        writer_redis_conn: aioredis.Redis = self.aioredis_pool.writer_redis_pool
        e_obj = None
        try:
            loop = asyncio.get_running_loop()
        except Exception as e:
            logging.warning("There was an error while trying to get event loop")
            logging.error(e, exc_info=True)
            return None
        kwargs = dict(
            payloadCommitId=payload_commit_id,
            snapshotCid=snapshot_cid,
            apiKeyHash=token_hash,
            projectId=project_id,
            tentativeBlockHeight=last_tentative_block_height,
        )
        partial_func = partial(self.contract.commitRecord, **kwargs)
        try:
            # try 3 times, wait 1-2 seconds between retries
            async for attempt in AsyncRetrying(reraise=True, stop=stop_after_attempt(3), wait=wait_random(1, 2)):
                with attempt:
                    tx_hash_obj = await loop.run_in_executor(None, partial_func)
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
            logging.info(
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

        if e_obj:
            logging.info(
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

        return tx_hash


@ray.remote
class CommitPayloadProcessorActor:
    def __init__(self):
        self.aioredis_pool = RedisPool(pool_size=10)
        formatter = logging.Formatter(
            u"%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s")
        stdout_handler = logging.StreamHandler(sys.stdout)
        stdout_handler.setLevel(logging.DEBUG)
        stdout_handler.setFormatter(formatter)

        stderr_handler = logging.StreamHandler(sys.stderr)
        stderr_handler.setLevel(logging.ERROR)
        stderr_handler.setFormatter(formatter)
        logging.basicConfig(level=logging.INFO, handlers=[stdout_handler, stderr_handler])

    async def single_payload_commit(self, msg_body):
        await self.aioredis_pool.populate()
        # semaphore: asyncio.BoundedSemaphore
        """
            - This function will take a pending payload and commit it to a smart contract
        """
        writer_redis_conn: aioredis.Redis = self.aioredis_pool.writer_redis_pool
        payload_commit_obj: PayloadCommit = PayloadCommit.parse_raw(msg_body)
        payload_commit_id = payload_commit_obj.commitId
        # payload_logger.debug(payload_data)
        snapshot = dict()
        core_payload = payload_commit_obj.payload
        project_id = payload_commit_obj.projectId
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
        result_tx_hash = await transaction_processor_actor.make_transaction.remote(
            snapshot_cid=snapshot_cid,
            token_hash=token_hash,
            payload_commit_id=payload_commit_id,
            last_tentative_block_height=payload_commit_obj.tentativeBlockHeight,
            project_id=payload_commit_obj.projectId,
            resubmission_block=payload_commit_obj.resubmissionBlock
        )
        if result_tx_hash:
            if not payload_commit_obj.resubmitted:
                last_snapshot_cid_key = redis_keys.get_last_snapshot_cid_key(project_id)
                logging.debug("Setting the last snapshot_cid as %s for project ID %s", snapshot_cid, project_id)
                _ = await writer_redis_conn.set(last_snapshot_cid_key, snapshot_cid)
            logging.debug('Setting tx hash %s against payload commit ID %s', result_tx_hash, payload_commit_id)
            await writer_redis_conn.set(redis_keys.get_payload_commit_key(payload_commit_id), result_tx_hash)


async def commit_message_cb(
    message: IncomingMessage
):
    async with message.process():
        # semaphore: asyncio.BoundedSemaphore
        """
            - This function will take a pending payload and commit it to a smart contract
        """
        # not going to wait on it, move on
        payload_commit_msg_processor_actor.single_payload_commit.remote(message.body)
        payload_logger.debug('Sent incoming payload commit message to processor actor: %s', message.body)


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


def verifier_crash_cb(fut: asyncio.Future):
    try:
        exc = fut.exception()
    except (asyncio.CancelledError, aioredis.ConnectionClosedError):
        payload_logger.error('Respawning commit payload task...')
        t = asyncio.ensure_future(payload_commit_queue_listener(asyncio.get_running_loop()))
        t.add_done_callback(verifier_crash_cb)
    except Exception as e:
        payload_logger.error('Commit payload task crashed')
        payload_logger.error(e, exc_info=True)


if __name__ == "__main__":
    ev_loop = asyncio.get_event_loop()
    transaction_processor_actor = TransactionActorAsync.options(
        max_concurrency=settings.max_payload_commits,
        lifetime='detached'
    ).remote()
    payload_commit_msg_processor_actor = CommitPayloadProcessorActor.options(
        max_concurrency=100,
        lifetime='detached'
    ).remote()
    payload_logger.debug("Starting the loop")
    f = asyncio.ensure_future(payload_commit_queue_listener(ev_loop))
    f.add_done_callback(verifier_crash_cb)
    try:
        ev_loop.run_forever()
    except:
        asyncio.get_event_loop().stop()

