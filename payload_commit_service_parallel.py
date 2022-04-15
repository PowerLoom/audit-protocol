from config import settings
from data_models import PayloadCommit, PendingTransaction
from eth_utils import keccak
from maticvigil.EVCore import EVCore
from typing import List
from utils.redis_conn import RedisPool, get_writer_redis_conn, REDIS_WRITER_CONN_CONF
from utils import redis_keys
from utils.rabbitmq_utils import RabbitmqSelectLoopInteractor
from functools import partial
from maticvigil.exceptions import EVBaseException
from tenacity import Retrying, stop_after_attempt, wait_random
from multiprocessing.managers import SyncManager
import threading
from concurrent.futures import ThreadPoolExecutor
import uuid
import queue
import ipfshttpclient
import psutil
import multiprocessing
import redis
import signal
import time
import logging
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


def get_process_exception_handler_wrapper():
    def deco(func):
        def mod_func(self, *args, **kwargs):
            try:
                return func(self, *args, **kwargs)
            except Exception as e:
                if not (
                        isinstance(e, GracefulSIGTERMExit) or
                        isinstance(e, GracefulSIGDeathExit) or
                        isinstance(e, KeyboardInterrupt)
                ):
                    payload_logger.error("Worker {} going down. Exception encountered: {}".format(self._name, e),
                                         exc_info=True)
                    self._queue.put(self._name)

        return mod_func

    return deco


process_exception_handler = get_process_exception_handler_wrapper()


class GracefulSIGTERMExit(Exception):
    def __str__(self):
        return "SIGTERM Exit"

    def __repr__(self):
        return "SIGTERM Exit"


class GracefulSIGDeathExit(Exception):
    def __init__(self, signal_information):
        Exception.__init__(self)
        self._sig = signal_information

    def __str__(self):
        return f"{self._sig} Exit"

    def __repr__(self):
        return f"{self._sig} Exit"


def sigterm_handler(signum, frame):
    raise GracefulSIGTERMExit


def sigkill_handler(signum, frame):
    raise GracefulSIGDeathExit(signum)


def chunks(start_idx, stop_idx, n):
    run_idx = 0
    for i in range(start_idx, stop_idx + 1, n):
        # Create an index range for l of n items:
        begin_idx = i  # if run_idx == 0 else i+1
        if begin_idx == stop_idx:
            return
        end_idx = i + n - 1 if i + n - 1 <= stop_idx else stop_idx
        run_idx += 1
        yield begin_idx, end_idx, run_idx


def make_transaction_sync(
    snapshot_cid,
    payload_commit_id,
    token_hash,
    last_tentative_block_height,
    project_id,
    writer_redis_conn: redis.Redis,
    audit_record_store_contract,
    resubmission_block=0,
):
    """
        - Create a unique transaction_id associated with this transaction,
        and add it to the set of pending transactions
    """

    e_obj = None
    kwargs = dict(
        payloadCommitId=payload_commit_id,
        snapshotCid=snapshot_cid,
        apiKeyHash=token_hash,
        projectId=project_id,
        tentativeBlockHeight=last_tentative_block_height,
    )

    try:
        # try 3 times, wait 1-2 seconds between retries
        for attempt in Retrying(reraise=True, stop=stop_after_attempt(3), wait=wait_random(1, 2)):
            with attempt:
                tx_hash_obj = audit_record_store_contract.commitRecord(**kwargs)
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
        _ = writer_redis_conn.zadd(
            name=pending_transaction_key,
            mapping={tx_hash_pending_entry.json(): last_tentative_block_height}
        )
        payload_logger.info(
            "Successful transaction %s committed to AuditRecord contract against payload commit ID %s",
            tx_hash_obj, payload_commit_id
        )
        service_log_update = {
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
        payload_logger.debug(service_log_update)
        # writer_redis_conn.zadd(
        #     name=redis_keys.get_payload_commit_id_process_logs_zset_key(
        #         project_id=project_id, payload_commit_id=payload_commit_id
        #     ),
        #     mapping={json.dumps(service_log_update): int(time.time())}
        # )

    if e_obj:
        payload_logger.info(
            "=" * 80 +
            "Commit Payload to AuditRecord contract failed. Tx was not successful against commit ID %s\n%s" +
            "=" * 80,
            payload_commit_id, e_obj
        )
        service_log_update = {
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
        payload_logger.debug(service_log_update)
        # writer_redis_conn.zadd(
        #     name=redis_keys.get_payload_commit_id_process_logs_zset_key(
        #         project_id=project_id, payload_commit_id=payload_commit_id
        #     ),
        #     mapping={json.dumps(service_log_update): int(time.time())
        #     }
        # )
        writer_redis_conn.close()
        return None

    writer_redis_conn.close()
    return tx_hash


def single_payload_commit_sync(payload_commit_obj: PayloadCommit, audit_record_store_contract):
    writer_redis_conn: redis.Redis = redis.Redis(**REDIS_WRITER_CONN_CONF, single_connection_client=True)
    payload_commit_id = payload_commit_obj.commitId
    # payload_logger.debug(payload_data)
    snapshot = dict()
    core_payload = payload_commit_obj.payload
    project_id = payload_commit_obj.projectId
    snapshot_cid = None
    if core_payload:
        ipfs_client_sync = ipfshttpclient.connect(settings.ipfs_url)
        try:
            # try 3 times, wait 1-2 seconds between retries
            for attempt in Retrying(reraise=True, stop=stop_after_attempt(3), wait=wait_random(1, 2)):
                with attempt:
                    if type(core_payload) is dict:
                        snapshot_cid = ipfs_client_sync.add_json(core_payload)
                    else:
                        try:
                            core_payload = json.dumps(core_payload)
                        except:
                            pass
                        snapshot_cid = ipfs_client_sync.add_str(str(core_payload))
                    if snapshot_cid:
                        break
        except Exception as e:
            service_log_update = {
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
            # writer_redis_conn.zadd(
            #     name=redis_keys.get_payload_commit_id_process_logs_zset_key(
            #         project_id=project_id, payload_commit_id=payload_commit_id
            #     ),
            #     mapping={json.dumps(service_log_update): int(time.time())}
            # )
            payload_logger.debug(service_log_update)
            return
        else:
            if snapshot_cid:
                service_log_update = {
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
                # writer_redis_conn.zadd(
                #     name=redis_keys.get_payload_commit_id_process_logs_zset_key(
                #         project_id=project_id, payload_commit_id=payload_commit_id
                #     ),
                #     mapping={json.dumps(service_log_update): int(time.time())
                #     }
                # )
                payload_logger.debug(service_log_update)
    else:
        snapshot_cid = payload_commit_obj.snapshotCID
    payload_cid_key = redis_keys.get_payload_cids_key(project_id)
    _ = writer_redis_conn.zadd(
        name=payload_cid_key,
        mapping={snapshot_cid: int(payload_commit_obj.tentativeBlockHeight)}
    )
    snapshot['cid'] = snapshot_cid
    snapshot['type'] = "HOT_IPFS"
    if not payload_commit_obj.apiKeyHash:
        token_hash = '0x' + keccak(text=json.dumps(snapshot)).hex()
    else:
        token_hash = payload_commit_obj.apiKeyHash
    result_tx_hash = make_transaction_sync(
        snapshot_cid=snapshot_cid,
        token_hash=token_hash,
        payload_commit_id=payload_commit_id,
        last_tentative_block_height=payload_commit_obj.tentativeBlockHeight,
        project_id=payload_commit_obj.projectId,
        resubmission_block=payload_commit_obj.resubmissionBlock,
        audit_record_store_contract=audit_record_store_contract,
        writer_redis_conn=writer_redis_conn
    )
    if result_tx_hash:
        if not payload_commit_obj.resubmitted:
            last_snapshot_cid_key = redis_keys.get_last_snapshot_cid_key(project_id)
            payload_logger.debug("Setting the last snapshot_cid as %s for project ID %s", snapshot_cid, project_id)
            _ = writer_redis_conn.set(last_snapshot_cid_key, snapshot_cid)
        payload_logger.debug('Setting tx hash %s against payload commit ID %s', result_tx_hash, payload_commit_id)
    writer_redis_conn.close()


class PayloadCommitRequestsEntry(multiprocessing.Process):
    def __init__(self, name, worker_id, sync_q: queue.Queue):
        multiprocessing.Process.__init__(self, name=name)
        self._name = name
        self._worker_id = worker_id
        self._queue = sync_q
        self._payload_commit_rabbitmq_queue_name = 'audit-protocol-commit-payloads'
        self.rabbitmq_interactor = None
        self._qd_msgs: List[PayloadCommit] = list()
        self._audit_record_store_contract = evc.generate_contract_sdk(
            contract_address=settings.audit_contract,
            app_name='auditrecords'
        )

    def _commit_payload_processing_callback(self, dont_use_ch, method, properties, body):
        self.rabbitmq_interactor._channel.basic_ack(delivery_tag=method.delivery_tag)
        if body:
            payload_logger.debug(body)
            payload_commit_obj: PayloadCommit = PayloadCommit.parse_raw(body)
            if payload_commit_obj.payload:
                payload_logger.debug('Recd. incoming payload commit message at tentative DAG height %s '
                                     'for project %s in chain height range %s to internal processing queue',
                                     payload_commit_obj.tentativeBlockHeight, payload_commit_obj.projectId,
                                     payload_commit_obj.payload.get('chainHeightRange')
                                     )
            else:
                if payload_commit_obj.resubmitted:
                    payload_logger.debug('Recd. incoming payload commit message at tentative DAG height %s '
                                         'for project %s for resubmission at block %s to internal processing queue',
                                         payload_commit_obj.tentativeBlockHeight, payload_commit_obj.projectId,
                                         payload_commit_obj.resubmissionBlock
                                         )
                else:
                    payload_logger.debug('Recd. incoming payload commit message at tentative DAG height %s '
                                         'for project %s',
                                         payload_commit_obj.tentativeBlockHeight, payload_commit_obj.projectId
                                         )
            t = threading.Thread(
                target=single_payload_commit_sync,
                daemon=True,
                kwargs={
                    'payload_commit_obj': payload_commit_obj,
                    'audit_record_store_contract': self._audit_record_store_contract
                }
            )
            t.start()

    @process_exception_handler
    def run(self) -> None:
        self.rabbitmq_interactor: RabbitmqSelectLoopInteractor = RabbitmqSelectLoopInteractor(
            consume_queue_name=self._payload_commit_rabbitmq_queue_name,
            consume_callback=self._commit_payload_processing_callback
        )
        payload_logger.debug(
            'Starting RabbitMQ consumer %s with worker ID %s on queue %s',
            self._name, self._worker_id, self._payload_commit_rabbitmq_queue_name
        )
        self.rabbitmq_interactor.run()


class InternalProcessMonitor(multiprocessing.Process):
    def __init__(self, name, q: queue.Queue):
        multiprocessing.Process.__init__(self, name=name)
        self._name = name
        self._queue = q

    def _reap_children(self, timeout=3):
        def on_terminate(proc):
            payload_logger.error("process {} terminated with exit code {}".format(proc, proc.returncode))

        procs = psutil.Process().children()
        # send SIGTERM
        for p in procs:
            p.terminate()
        gone, alive = psutil.wait_procs(procs, timeout=timeout, callback=on_terminate)
        if alive:
            # send SIGKILL
            for p in alive:
                payload_logger.error("process {} survived SIGTERM; trying SIGKILL" % p)
                p.kill()
            gone, alive = psutil.wait_procs(alive, timeout=timeout, callback=on_terminate)
            if alive:
                # give up
                for p in alive:
                    payload_logger.error("process {} survived SIGKILL; giving up" % p)

    def run(self):
        try:
            payload_logger.debug("Starting Internal process monitor...")
            jobs = []
            try:
                workers = multiprocessing.cpu_count() * 2
                # workers = 4
            except NotImplementedError:
                workers = 1
            jobs = [
                PayloadCommitRequestsEntry(name="PayloadCommitsEntryPoint-" + str(_id), worker_id=str(uuid.uuid4())[:5],
                                           sync_q=self._queue)
                for _id in range(workers)]
            for w in jobs:
                w.start()
            while True:
                crashed = self._queue.get()  # forever blocking
                payload_logger.error("Received crashed worker notification: {} ".format(crashed))
                if "PayloadCommitsEntryPoint" in crashed:
                    jobs = [j for j in jobs if j.is_alive()]
                    payload_logger.error("Bringing up worker {}".format(crashed))
                    j = PayloadCommitRequestsEntry(name=crashed, sync_q=self._queue, worker_id=str(uuid.uuid4())[:5])
                    j.start()
                    jobs.append(j)
        except KeyboardInterrupt:
            payload_logger.error("Internal Process Monitor received SIGINT. Shutting down")
            for w in jobs:
                w.terminate()
                w.join()
        except Exception as e:
            payload_logger.error(
                "%s: Unexpected exception in process monitor. Reaping all children", psutil.Process().name(),
                exc_info=True)
            self._reap_children()


def super_manager_init():
    signal.signal(signal.SIGINT, signal.SIG_IGN)
    signal.signal(signal.SIGTERM, signal.SIG_IGN)
    signal.signal(signal.SIGHUP, signal.SIG_IGN)
    signal.signal(signal.SIGQUIT, signal.SIG_IGN)
    payload_logger.debug("{}: Initialized SyncManager for shared queue".format(psutil.Process().name()))


if __name__ == "__main__":
    evc = EVCore(verbose=True)

    signal.signal(signal.SIGTERM, sigterm_handler)
    signal.signal(signal.SIGHUP, sigkill_handler)
    signal.signal(signal.SIGQUIT, sigkill_handler)

    super_manager = SyncManager()
    super_manager.start(super_manager_init)
    p_m_q = super_manager.Queue()
    process_monitor = InternalProcessMonitor("pm", p_m_q)
    try:
        process_monitor.start()
        process_monitor.join()
    except KeyboardInterrupt:
        payload_logger.error("__main__ received SIGINT. Going down...")
    except GracefulSIGTERMExit as e:
        payload_logger.error(f"__main__ received SIGTERM signal: {e}")
    except GracefulSIGDeathExit as e:
        payload_logger.error(f"__main__ received SHUTDOWN signal: {e}")
    finally:
        super_manager.shutdown()
        process_monitor.join()