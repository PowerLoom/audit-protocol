from exceptions import SelfExitException, GenericExitOnSignal
from multiprocessing import Process, log_to_stderr
from utils.redis_conn import REDIS_CONN_CONF, get_redis_conn_from_pool
from utils.rabbitmq_utils import RabbitmqSelectLoopInteractor, get_rabbitmq_queue_name

from typing import Dict, Union
from config import settings
from uuid import uuid4
from utils.helper_functions import cleanup_children_procs
from utils import diffmap_utils
from data_models import DiffCalculationRequest, DAGBlock
from async_ipfshttpclient.main import AsyncIPFSClient
import redis
import multiprocessing
import logging
import logging.handlers
import json
import time
from signal import signal, pause, SIGINT, SIGTERM, SIGQUIT, SIGCHLD, SIG_DFL
import os
import sys

formatter = logging.Formatter('%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s')
diff_service_main_logger = logging.getLogger(__name__)
diff_service_main_logger.setLevel(logging.DEBUG)

stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setLevel(logging.DEBUG)
stdout_handler.setFormatter(formatter)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setLevel(logging.ERROR)
stderr_handler.setFormatter(formatter)

diff_service_main_logger.addHandler(stdout_handler)
diff_service_main_logger.addHandler(stderr_handler)


class DiffCalculationCallbackWorker(Process):
    def __init__(self, name, unique_id, **kwargs):
        Process.__init__(self, name=name, **kwargs)
        self._id = unique_id
        self._worker_name = name+self._id
        self._shutdown_initiated = False
        self._queue_name = get_rabbitmq_queue_name('diff-requests')

    def signal_handler(self, signum, frame):
        if signum in [SIGINT, SIGTERM, SIGQUIT] and not self._shutdown_initiated:
            self._shutdown_initiated = True
            self.rabbitmq_interactor.stop()

    def callback(self, dont_use_ch, method, properties, body):
        self.rabbitmq_interactor._channel.basic_ack(delivery_tag=method.delivery_tag)
        if settings.calculate_diff is False:
            return
        try:
            command: DiffCalculationRequest = DiffCalculationRequest.parse_raw(body)
        except:
            diff_service_main_logger.info('Error converting incoming request into data model: %s', body)
            return
        diff_service_main_logger.debug(body)

        dag_block = DAGBlock(
            height=command.tentative_block_height,
            prevCid={'/': command.lastDagCid},
            data={'cid': {'/': command.payloadCid}, 'type': 'HOT_IPFS'},
            txHash=command.txHash,
            timestamp=command.timestamp
        )
        redis_conn = redis.Redis(connection_pool=self._redis_conn_pool)
        try:
            diff_map = diffmap_utils.calculate_diff(
                dag_cid=command.dagCid,
                dag=dag_block,
                project_id=command.project_id,
                ipfs_client=self._ipfs_client,
                writer_redis_conn=redis_conn
            )
        except json.decoder.JSONDecodeError as jerr:
            diff_service_main_logger.debug("There was an error while decoding the JSON data: %s", jerr, exc_info=True)
        except Exception as e:
            diff_service_main_logger.debug(
                "There was an exception while calculating diff at %s: %s",
                command.tentative_block_height, e, exc_info=True
            )
        else:
            diff_service_main_logger.debug(
                "The diff map calculated and cached | Project %s | At Height %s: %s",
                command.project_id, command.tentative_block_height, diff_map
            )

    def run(self) -> None:
        for signame in [SIGINT, SIGTERM, SIGQUIT]:
            signal(signame, self.signal_handler)

        # TODO: what is this async IPFS client doing in sync code?
        #       cant initialize any way
        #       all operations are async
        self._ipfs_client = AsyncIPFSClient(addr=settings.ipfs_url)
        self._redis_conn_pool = redis.BlockingConnectionPool(**REDIS_CONN_CONF, max_connections=20)
        self.rabbitmq_interactor: RabbitmqSelectLoopInteractor = RabbitmqSelectLoopInteractor(
            consume_queue_name=self._queue_name,
            consume_callback=self.callback,
            consumer_worker_name=self._worker_name
        )
        diff_service_main_logger.debug(
            'Diff Calculation worker %s starting RabbitMQ consumer on queue %s',
            self._id, self._queue_name
        )
        self.rabbitmq_interactor.run()
        diff_service_main_logger.debug('RabbitMQ interactor ioloop ended...')


class ProcessCore(Process):
    def __init__(self, name, **kwargs):
        Process.__init__(self, name=name, **kwargs)
        self._spawned_processes_map: Dict[str, Union[Process, None]] = dict()
        self._shutdown_initiated = False

    def signal_handler(self, signum, frame):
        if signum == SIGCHLD and not self._shutdown_initiated:
            pid, status = os.waitpid(-1, os.WNOHANG | os.WUNTRACED | os.WCONTINUED)
            if os.WIFCONTINUED(status) or os.WIFSTOPPED(status):
                return
            if os.WIFSIGNALED(status) or os.WIFEXITED(status):
                diff_service_main_logger.debug(
                    'Received process crash notification for diff calculation worker process PID: %s', pid
                )
                for k, v in self._spawned_processes_map.items():
                    # k is the unique ID
                    if v != -1 and v.pid == pid:
                        diff_service_main_logger.debug('RESPAWNING: diff calculation worker against ID %s', k)
                        proc_obj = DiffCalculationCallbackWorker(
                            name='AuditProtocol|DiffService|Worker|',
                            unique_id=k
                        )
                        proc_obj.start()
                        diff_service_main_logger.debug(
                            'RESPAWNED: diff calculation worker against ID %s with PID: %s', k, proc_obj.pid
                        )
                        self._spawned_processes_map[k] = proc_obj
        elif signum in [SIGINT, SIGTERM, SIGQUIT]:
            self._shutdown_initiated = True

    @cleanup_children_procs
    def run(self) -> None:

        for signame in [SIGINT, SIGTERM, SIGQUIT, SIGCHLD]:
            signal(signame, self.signal_handler)
        # launch worker processes
        try:
            workers = multiprocessing.cpu_count()
        except NotImplementedError:
            workers = 4
        for _ in range(workers):
            unique_id = str(uuid4())[:5]
            self._spawned_processes_map[unique_id] = DiffCalculationCallbackWorker(
                name='AuditProtocol|DiffService|Worker|',
                unique_id=unique_id
            )
        for w in self._spawned_processes_map.values():
            w.start()
        while not self._shutdown_initiated:
            time.sleep(2)
        raise SelfExitException


if __name__ == '__main__':
    p = ProcessCore(name='AuditProtocol|DiffService|Core')
    p.start()
    while p.is_alive():
        diff_service_main_logger.debug('Process hub core is still alive. waiting on it to join...')
        try:
            p.join()
        except:
            pass
