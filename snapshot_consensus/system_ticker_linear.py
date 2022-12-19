from signal import SIGINT, SIGTERM, SIGQUIT, signal
import signal
import time
import queue
import threading
import logging
import sys
from functools import wraps
from time import sleep
from multiprocessing import Process
from .conf import settings
from utils.redis_conn import RedisPool
from .helpers.redis_keys import get_system_ticker_linear_last_epoch, get_system_ticker_linear_epoch_history
from setproctitle import setproctitle
from exceptions import GenericExitOnSignal
from .helpers.rpc_helper import ConstructRPC
from .helpers.message_models import RPCNodesObject
import asyncio
import json


def chunks(start_idx, stop_idx, n):
    run_idx = 0
    for i in range(start_idx, stop_idx + 1, n):
        # Create an index range for l of n items:
        begin_idx = i  # if run_idx == 0 else i+1
        if begin_idx == stop_idx + 1:
            return
        end_idx = i + n - 1 if i + n - 1 <= stop_idx else stop_idx
        run_idx += 1
        yield begin_idx, end_idx, run_idx


def redis_cleanup(fn):
    @wraps(fn)
    async def wrapper(self, *args, **kwargs):
        try:
            await fn(self, *args, **kwargs)
        except (GenericExitOnSignal, KeyboardInterrupt):
            try:
                self._logger.debug('Waiting for pushing latest epoch to Redis')

                await self.writer_redis_pool.set(get_system_ticker_linear_last_epoch(), self.last_sent_block)

                self._logger.debug('Shutting down after sending out last epoch with end block height as %s,'
                                    ' starting blockHeight to be used during next restart is %s'
                                    ,self.last_sent_block
                                    , self.last_sent_block+1)
            except Exception as E:
                self._logger.error('Error while saving last state: %s', E)
        except Exception as E:
            self._logger.error('Error while running process: %s', E)
        finally:
            self._logger.debug('Shutting down')
            if not self._simulation_mode:
                sys.exit(0)
    return wrapper


class LinearTickerProcess(Process):
    def __init__(self, name, simulation_mode=False):
        Process.__init__(self, name=name)
        self._logger = logging.getLogger(self.name)

        stdout_handler = logging.StreamHandler(sys.stdout)
        stderr_handler = logging.StreamHandler(sys.stderr)
        stderr_handler.setLevel(logging.ERROR)

        self._logger.setLevel(logging.DEBUG)
        self._logger.addHandler(stdout_handler)
        self._logger.addHandler(stderr_handler)

        self._simulation_mode = simulation_mode
        self._shutdown_initiated = False
        self.last_sent_block = 0
        

    async def setup(self, **kwargs):
        if self._simulation_mode:
            self._logger.debug('Simulation mode is on')
            self.aioredis_pool = RedisPool(writer_redis_conf=settings.test_redis)
        else:
            self.aioredis_pool = RedisPool(writer_redis_conf=settings.test_redis)
        await self.aioredis_pool.populate()
        self.reader_redis_pool = self.aioredis_pool.reader_redis_pool
        self.writer_redis_pool = self.aioredis_pool.writer_redis_pool
        self.redis_thread: threading.Thread
        
        self._end = kwargs.get('end')

    def _generic_exit_handler(self, signum, sigframe):
        if signum in [SIGINT, SIGTERM, SIGQUIT] and not self._shutdown_initiated:
            self._shutdown_initiated = True
            raise GenericExitOnSignal


    @redis_cleanup
    async def init(self, begin_block_epoch:int =0):
        for signame in [signal.SIGINT, signal.SIGTERM, signal.SIGQUIT]:
            signal.signal(signame, self._generic_exit_handler)
        last_block_data_redis = await self.writer_redis_pool.get(name=get_system_ticker_linear_last_epoch())
        if last_block_data_redis:
            # Can't provide begin block which previous state is present in redis
            if begin_block_epoch != 0:
                self._logger.debug(f'Attempting to start from {begin_block_epoch} but last block {last_block_data_redis.decode("utf-8")} found in Redis.')
                self._logger.debug('Either use clean_slate_ticker.py to reset the Redis state or remove the begin_block_epoch argument.')
                if not self._simulation_mode:
                    sys.exit(0)
                else:
                    return
            else:
                self._logger.debug('Begin block not given, attempting starting from Redis')
                begin_block_epoch = int(last_block_data_redis.decode("utf-8"))+1
                self._logger.debug(f'Found last epoch block : {begin_block_epoch} in Redis. Starting from checkpoint.')

        setproctitle(self.name)
            
        end_block_epoch = self._end
        # Sleep only 1 second to speed up simulation
        if self._simulation_mode:
            sleep_secs_between_chunks = 1
        else:
            sleep_secs_between_chunks = 60
        rpc_obj = ConstructRPC(network_id=settings.chain.chain_id)
        rpc_urls = []
        for node in settings.chain.rpc.nodes:
            self._logger.debug("node %s",node.url)
            rpc_urls.append(node.url)
        rpc_nodes_obj = RPCNodesObject(
            NODES=rpc_urls,
            RETRY_LIMIT=settings.chain.rpc.retry
        )
        generated_block_counter = 0
        self._logger.debug('Starting %s', Process.name)
        while True if not self._simulation_mode else generated_block_counter < 10:
            try:
                cur_block = rpc_obj.rpc_eth_blocknumber(rpc_nodes=rpc_nodes_obj)
            except Exception as ex:
                self._logger.error(
                    "Unable to fetch latest block number due to RPC failure %s. Retrying after %s seconds.",ex,settings.chain.epoch.block_time)
                sleep(settings.chain.epoch.block_time)
                continue
            else:
                self._logger.debug('Got current head of chain: %s', cur_block)
                if not begin_block_epoch:
                    self._logger.debug('Begin of epoch not set')
                    begin_block_epoch = cur_block
                    self._logger.debug('Set begin of epoch to current head of chain: %s', cur_block)
                    self._logger.debug('Sleeping for: %s seconds', settings.chain.epoch.block_time)
                    sleep(settings.chain.epoch.block_time)
                else:
                    # self._logger.debug('Picked begin of epoch: %s', begin_block_epoch)
                    end_block_epoch = cur_block - settings.chain.epoch.head_offset
                    if not (end_block_epoch - begin_block_epoch + 1) >= settings.chain.epoch.height:
                        sleep_factor = settings.chain.epoch.height - ((end_block_epoch - begin_block_epoch) + 1)
                        self._logger.debug('Current head of source chain estimated at block %s after offsetting | '
                                                   '%s - %s does not satisfy configured epoch length. '
                                                   'Sleeping for %s seconds for %s blocks to accumulate....',
                                                   end_block_epoch, begin_block_epoch, end_block_epoch,
                                                   sleep_factor*settings.chain.epoch.block_time, sleep_factor
                                                   )
                        time.sleep(sleep_factor * settings.chain.epoch.block_time)
                        continue
                    self._logger.debug('Chunking blocks between %s - %s with chunk size: %s', begin_block_epoch,
                                               end_block_epoch, settings.chain.epoch.height)
                    for epoch in chunks(begin_block_epoch, end_block_epoch, settings.chain.epoch.height):
                        if epoch[1] - epoch[0] + 1 < settings.chain.epoch.height:
                            self._logger.debug(
                                'Skipping chunk of blocks %s - %s as minimum epoch size not satisfied | Resetting chunking'
                                ' to begin from block %s',
                                epoch[0], epoch[1], epoch[0]
                            )
                            begin_block_epoch = epoch[0]
                            break
                        epoch_block = {'begin': epoch[0], 'end': epoch[1]}
                        generated_block_counter += 1
                        if self._simulation_mode and generated_block_counter >= 10:
                            break
                        self._logger.debug('Epoch of sufficient length found: %s', epoch_block)
                        
                        await self.writer_redis_pool.set(name=get_system_ticker_linear_last_epoch(), value=epoch_block['end'])

                        await self.writer_redis_pool.zadd(
                            name=get_system_ticker_linear_epoch_history(),
                            mapping={json.dumps(epoch_block['end']): int(time.time())}
                        )

                        self.last_sent_block = epoch_block['end']
                        self._logger.debug('Waiting to push next epoch in %d seconds...', sleep_secs_between_chunks)
                        # fixed wait
                        sleep(sleep_secs_between_chunks)
                    else:
                        begin_block_epoch = end_block_epoch + 1


def main(begin_block:int=0, simulation_mode=False):
    """Spin up the ticker process in event loop"""
    ticker_process = LinearTickerProcess(name="PowerLoom|EpochTracker|Linear", simulation_mode=simulation_mode)
    kwargs = dict()
    asyncio.run(ticker_process.setup(**kwargs))
    asyncio.run(ticker_process.init(begin_block))


if __name__ == '__main__':

    args = sys.argv
    if len(args) > 1:
        begin_block = int(args[1])
        main(begin_block=begin_block)
    else:
        main()