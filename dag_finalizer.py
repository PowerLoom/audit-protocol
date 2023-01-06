from functools import wraps
from utils.rabbitmq_utils import get_rabbitmq_connection, get_rabbitmq_channel, get_rabbitmq_queue_name, get_rabbitmq_routing_key, get_rabbitmq_core_exchange
from utils import redis_keys
from utils import helper_functions
from async_ipfshttpclient.main import AsyncIPFSClientSingleton, AsyncIPFSClient
from aio_pika import DeliveryMode, Message, IncomingMessage
from utils import dag_utils
from utils.redis_conn import RedisPool
from functools import partial
from aio_pika.pool import Pool
from typing import Optional, Dict, List
from urllib.parse import urljoin
from data_models import (
    PayloadCommit, PendingTransaction, ProjectDAGChainSegmentMetadata, DAGFinalizerCallback, DAGFinalizerCBEventData,
    AuditRecordTxEventData, SnapshotterIssue, SnapshotterIssueSeverity
)
from tenacity import retry, wait_random_exponential, stop_after_attempt, retry_if_exception_type
from httpx import AsyncClient, Timeout, Limits, AsyncHTTPTransport
import httpx._exceptions as httpx_exceptions
from redis import asyncio as aioredis
from config import settings
if settings.use_consensus:
    from snapshot_consensus.data_models import EpochConsensusStatus, SubmissionResponse, SnapshotBase, EpochBase
import resource
import uvloop
import logging
import sys
import uuid
import time
import json
import asyncio

asyncio.set_event_loop_policy(uvloop.EventLoopPolicy())


def acquire_project_atomic_lock(fn):
    @wraps(fn)
    async def wrapper(self, *args, **kwargs):
        kwarg_event_data: DAGFinalizerCallback = args[0]
        project_id = kwarg_event_data.event_data.projectId
        try:
            await self._asyncio_lock_map[project_id].acquire()
        except KeyError:
            self._logger.debug('Creating asyncio lock for project ID %s', kwarg_event_data.event_data.projectId)
            self._asyncio_lock_map[project_id] = asyncio.Lock()
            await self._asyncio_lock_map[project_id].acquire()
        finally:
            try:
                self._logger.debug(
                    'Acquired asyncio lock for project ID %s for callback %s',
                    kwarg_event_data.event_data.projectId, kwarg_event_data
                )
                return await fn(self, *args, **kwargs)
            except Exception as e:
                self._logger.error(
                    'Exception while processing callback for project ID %s: %s | %s',
                    kwarg_event_data.event_data.projectId, kwarg_event_data, e,
                    exc_info=True
                )
            finally:
                self._asyncio_lock_map[project_id].release()
                self._logger.debug(
                    'Released asyncio lock for project ID %s for callback %s',
                    kwarg_event_data.event_data.projectId, kwarg_event_data
                )
    return wrapper


class CustomAdapter(logging.LoggerAdapter):
    """
    This example adapter expects the passed in dict-like object to have a
    'txHash' key, whose value in brackets is prepended to the log message.
    """

    def process(self, msg, kwargs):
        return '[%s] %s' % (self.extra['txHash'], msg), kwargs


class DAGFinalizationCallbackProcessor:
    _rmq_connection_pool: Pool
    _rmq_channel_pool: Pool
    _aioredis_pool: RedisPool
    _writer_redis_pool: aioredis.Redis
    _reader_redis_pool: aioredis.Redis
    _httpx_client: AsyncClient
    _ipfs_singleton: AsyncIPFSClientSingleton
    _ipfs_writer_client: AsyncIPFSClient
    _ipfs_reader_client: AsyncIPFSClient

    def __init__(self):
        self._q = get_rabbitmq_queue_name('dag-processing')
        self._rmq_routing = get_rabbitmq_routing_key('dag-processing')
        formatter = logging.Formatter(
            u"%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s")
        stdout_handler = logging.StreamHandler(sys.stdout)
        stdout_handler.setLevel(logging.DEBUG)
        stdout_handler.setFormatter(formatter)

        stderr_handler = logging.StreamHandler(sys.stderr)
        stderr_handler.setLevel(logging.ERROR)
        stderr_handler.setFormatter(formatter)
        self._logger = logging.getLogger(__name__)
        self._logger.setLevel(logging.DEBUG)
        self._logger.addHandler(stdout_handler)
        self._logger.addHandler(stderr_handler)
        self._asyncio_lock_map: Dict[str, asyncio.Lock] = dict()

    async def _identify_prune_target(self, project_id, tentative_max_height):
        # should be triggered when finalized_height (mod) segment_size == 1
        writer_redis_conn: aioredis.Redis = self._writer_redis_pool
        reader_redis_conn: aioredis.Redis = self._reader_redis_pool
        # if dagSegments are not present, initialize for the entire chain
        # first run, do not divide between height 1 -> cur_max
        # back up entire segment, 1 - 2800, for example with expected segment length, 700
        # insertion of DAG block at 2801, should contain prevCid = null right there and not refer to 2800
        # next run, qualified target would be 2801 - 3500, and would be triggered at height 3501

        # store
        p_ = await reader_redis_conn.hgetall(
            name=redis_keys.get_project_dag_segments_key(project_id)
        )

        # get the DAG CID at tentative_max_height - 1 to set the endDAGCid field against the chain segment
        # in project state metadata
        last_dag_cid = await helper_functions.get_dag_cid(
            project_id=project_id,
            block_height=tentative_max_height - 1,
            reader_redis_conn=reader_redis_conn
        )

        begin_height = 1
        if len(p_) > 0:
            sorted_keys = sorted(list(p_.keys()), key=lambda x: (len(x), x))
            sorted_dict = {k: p_[k] for k in sorted_keys}
            last_dag_segment = ProjectDAGChainSegmentMetadata.parse_raw(sorted_dict[sorted_keys[-1]])
            begin_height = last_dag_segment.endHeight + 1
        new_project_dag_segment = ProjectDAGChainSegmentMetadata(
            beginHeight=begin_height,
            endHeight=tentative_max_height - 1,
            endDAGCID=last_dag_cid,
            storageType='pending'
        )
        await writer_redis_conn.hset(redis_keys.get_project_dag_segments_key(project_id),
                                     new_project_dag_segment.endHeight, new_project_dag_segment.json())

    async def _in_order_block_creation_and_state_update(
            self,
            dag_finalizer_callback_obj: DAGFinalizerCallback,
            post_finalization_pending_txs,
            parent_cid_height_diff,
            custom_logger_obj,
            reader_redis_conn: aioredis.Redis,
            writer_redis_conn: aioredis.Redis,
            _httpx_client: AsyncClient
    ):
        blocks_finalized = list()
        project_id = dag_finalizer_callback_obj.event_data.projectId
        top_level_tentative_height_cb = dag_finalizer_callback_obj.event_data.tentativeBlockHeight
        fetch_prev_cid_for_dag_block_creation = True
        if top_level_tentative_height_cb > settings.pruning.segment_size and \
                top_level_tentative_height_cb % settings.pruning.segment_size == 1:
            await self._identify_prune_target(project_id, top_level_tentative_height_cb)
            fetch_prev_cid_for_dag_block_creation = False
        dag_cid, dag_block = await dag_utils.create_dag_block_update_project_state(
            tx_hash=dag_finalizer_callback_obj.txHash, request_id=dag_finalizer_callback_obj.requestID,
            project_id=dag_finalizer_callback_obj.event_data.projectId,
            payload_commit_id=dag_finalizer_callback_obj.event_data.payloadCommitId,
            tentative_block_height_event_data=top_level_tentative_height_cb,
            snapshot_cid=dag_finalizer_callback_obj.event_data.snapshotCid,
            timestamp=dag_finalizer_callback_obj.event_data.timestamp, reader_redis_conn=reader_redis_conn,
            writer_redis_conn=writer_redis_conn,
            fetch_prev_cid_for_dag_block_creation=fetch_prev_cid_for_dag_block_creation,
            parent_cid_height_diff=parent_cid_height_diff, ipfs_write_client=self._ipfs_writer_client,
            httpx_client=_httpx_client, custom_logger_obj=custom_logger_obj)
        blocks_finalized.append(dag_block)
        all_qualified_dag_addition_txs = filter(
            lambda x: PendingTransaction.parse_raw(x[0]).lastTouchedBlock == -1,
            post_finalization_pending_txs
        )
        cur_max_height_project = top_level_tentative_height_cb
        for pending_tx_entry, _tt_block_height in all_qualified_dag_addition_txs:
            pending_tx_obj: PendingTransaction = PendingTransaction.parse_raw(pending_tx_entry)
            _tt_block_height = int(_tt_block_height)
            pending_q_fetch_prev_cid_for_dag_block_creation = True
            if _tt_block_height == cur_max_height_project + 1:
                if _tt_block_height > settings.pruning.segment_size and \
                        _tt_block_height % settings.pruning.segment_size == 1:
                    await self._identify_prune_target(project_id, _tt_block_height)
                    pending_q_fetch_prev_cid_for_dag_block_creation = False
                custom_logger_obj.info(
                    'Processing queued confirmed tx %s at tentative_block_height: %s',
                    pending_tx_obj, _tt_block_height
                )
                """ Create the dag block for this event """
                dag_cid, dag_block = await dag_utils.create_dag_block_update_project_state(
                    tx_hash=pending_tx_obj.event_data.txHash, request_id=pending_tx_obj.requestID,
                    project_id=project_id, payload_commit_id=pending_tx_obj.event_data.payloadCommitId,
                    tentative_block_height_event_data=_tt_block_height,
                    snapshot_cid=pending_tx_obj.event_data.snapshotCid,
                    timestamp=int(pending_tx_obj.event_data.timestamp), reader_redis_conn=reader_redis_conn,
                    writer_redis_conn=writer_redis_conn,
                    fetch_prev_cid_for_dag_block_creation=pending_q_fetch_prev_cid_for_dag_block_creation,
                    parent_cid_height_diff=parent_cid_height_diff, ipfs_write_client=self._ipfs_writer_client,
                    httpx_client=_httpx_client, custom_logger_obj=custom_logger_obj)
                cur_max_height_project = _tt_block_height
                blocks_finalized.append(dag_block)
        return blocks_finalized
        # return _dag_cid, dag_block

    @acquire_project_atomic_lock
    async def _payload_to_dag_processor_task(self, event_data: DAGFinalizerCallback):
        """ Get data from the event """
        project_id = event_data.event_data.projectId
        tx_hash = event_data.txHash
        request_id = event_data.requestID
        asyncio.current_task(asyncio.get_running_loop()).set_name('TxProcessor-' + tx_hash)
        custom_logger = CustomAdapter(self._logger, {'txHash': tx_hash})
        tentative_block_height_event_data = int(event_data.event_data.tentativeBlockHeight)
        writer_redis_conn: aioredis.Redis = self._writer_redis_pool
        reader_redis_conn: aioredis.Redis = self._reader_redis_pool

        # Get the max block height(finalized after all error corrections and reorgs) for the project_id
        finalized_block_height_project = await helper_functions.get_block_height(
            project_id=project_id,
            reader_redis_conn=reader_redis_conn
        )

        custom_logger.debug(
            "Event Data Tentative Block Height: %s | Finalized Project %s Block Height: %s",
            tentative_block_height_event_data, project_id, finalized_block_height_project
        )
        if tentative_block_height_event_data <= finalized_block_height_project:
            custom_logger.debug("Discarding event at height %s | %s", tentative_block_height_event_data, event_data)
            await dag_utils.discard_event(
                project_id=project_id,
                request_id=request_id,
                tentative_block_height=tentative_block_height_event_data,
                writer_redis_conn=writer_redis_conn
            )
        elif tentative_block_height_event_data > finalized_block_height_project + 1:
            custom_logger.debug(
                "An out of order event arrived at tentative height %s | Project ID %s | "
                "Current finalized height: %s",
                tentative_block_height_event_data, project_id, finalized_block_height_project
            )

            discarded_transactions_key = redis_keys.get_discarded_transactions_key(project_id)

            custom_logger.debug("Checking if requestID %s is in the list of pending Transactions", request_id)

            _ = await reader_redis_conn.zrangebyscore(
                name=redis_keys.get_pending_transactions_key(project_id),
                min=float('-inf'),
                max=float('+inf'),
                withscores=True
            )
            is_pending = False
            # will be used to reset lastTouchBlocked tag in pending set
            # we will use the raw bytes entry to safely address the member in the zset
            pending_tx_set_entry: Optional[bytes] = None
            for k in _:
                pending_tx_obj: PendingTransaction = PendingTransaction.parse_raw(k[0])
                if pending_tx_obj.requestID == request_id and int(k[1]) == event_data.event_data.tentativeBlockHeight:
                    is_pending = True
                    pending_tx_set_entry = k[0]
                    break

            if not is_pending:
                custom_logger.error(
                    "Discarding event because requestID %s not in pending transactions | Project ID %s | %s",
                    request_id, project_id, event_data
                )

                _ = await writer_redis_conn.zadd(
                    name=discarded_transactions_key,
                    mapping={request_id: tentative_block_height_event_data}
                )

            else:
                custom_logger.debug(
                    "Saving Event data for tentative height: %s | %s ",
                    tentative_block_height_event_data, event_data
                )
                await dag_utils.save_event_data(
                    event_data=event_data,
                    writer_redis_conn=writer_redis_conn,
                    pending_tx_set_entry=pending_tx_set_entry
                )

                # first, we get transactions in the 'pending' state
                pending_confirmation_callbacks_txs = await reader_redis_conn.zrangebyscore(
                    name=redis_keys.get_pending_transactions_key(project_id),
                    min=finalized_block_height_project + 1,
                    max=tentative_block_height_event_data,
                    withscores=True
                )
                # filter out prior resubmitted transactions not older than 10 blocks
                # or if they are 'fresh' as well (lastTouchedBlock = 0) yet 10 blocks have passed since confirmation
                # has arrived (the score of the pending tx entry is used for this, which is the tentative block
                # height at which the tx callback is supposed to arrive)
                # not to be considered in this set are the ones already enqueued for block creation
                # (lastTouchedBlock == -1)
                num_block_to_wait_for_resubmission = 10
                pending_confirmation_callbacks_txs_filtered = list(filter(
                    lambda x:
                    PendingTransaction.parse_raw(x[0]).lastTouchedBlock == 0 and
                    int(x[1]) + num_block_to_wait_for_resubmission <= tentative_block_height_event_data,
                    pending_confirmation_callbacks_txs
                ))
                custom_logger.info(
                    'Pending transactions qualified for resubmission on account of being unconfirmed or '
                    'not yet considered since the last resubmission attempt: %s',
                    pending_confirmation_callbacks_txs_filtered
                )
                # map these tx entries eligible for resubmission to their tentative block height scores
                if pending_confirmation_callbacks_txs_filtered:
                    pending_confirmation_callbacks_txs_filtered_map = {
                        int(x[1]): x[0] for x in pending_confirmation_callbacks_txs_filtered
                    }
                    custom_logger.info(
                        'Preparing to send out txs for reprocessing for the ones that have not '
                        'received any callbacks (possibility of dropped txs) : %s',
                        pending_confirmation_callbacks_txs_filtered
                    )
                    # send out txs for reprocessing
                    republished_txs = list()
                    replaced_dummy_tx_entries_in_pending_set = list()
                    async with self._rmq_channel_pool.acquire() as channel:
                        # to save a call to rabbitmq. we already initialize exchanges and queues beforehand
                        # always ensure exchanges and queues are initialized as part of launch sequence,
                        # not to be checked here
                        exchange = await channel.get_exchange(
                            name=get_rabbitmq_core_exchange(),
                            ensure=False
                        )
                        for queued_tentative_height_ in pending_confirmation_callbacks_txs_filtered_map.keys():
                            # get the tx hash from the filtered set of qualified pending transactions
                            single_pending_tx_entry = pending_confirmation_callbacks_txs_filtered_map[
                                queued_tentative_height_]
                            pending_tx_obj: PendingTransaction = PendingTransaction.parse_raw(
                                single_pending_tx_entry
                            )
                            request_id = pending_tx_obj.requestID
                            # fetch transaction input data
                            tx_commit_details = pending_tx_obj.event_data
                            if not tx_commit_details:
                                custom_logger.critical(
                                    'Possible irrecoverable gap in DAG chain creation %s | '
                                    'Did not find cached input data against past requestID: %s | '
                                    'Last finalized height: %s | '
                                    'Qualified txs for resubmission: %s',
                                    project_id, request_id,
                                    finalized_block_height_project,
                                    pending_confirmation_callbacks_txs_filtered_map
                                )
                                continue
                            # send once more to payload commit service
                            payload_commit_obj = PayloadCommit(
                                **{
                                    'projectId': project_id,
                                    'commitId': tx_commit_details.payloadCommitId,
                                    'snapshotCID': tx_commit_details.snapshotCid,
                                    'tentativeBlockHeight': tx_commit_details.tentativeBlockHeight,
                                    'apiKeyHash': tx_commit_details.apiKeyHash,
                                    'resubmitted': True,
                                    'requestID': request_id,
                                    'resubmissionBlock': tentative_block_height_event_data,
                                    'skipAnchorProof': tx_commit_details.skipAnchorProof
                                }
                            )
                            message = Message(
                                payload_commit_obj.json().encode('utf-8'),
                                delivery_mode=DeliveryMode.PERSISTENT,
                            )
                            await exchange.publish(
                                message=message,
                                routing_key=get_rabbitmq_routing_key('commit-payloads')
                            )
                            custom_logger.debug(
                                'Re-Published payload against commit ID %s , tentative block height %s for '
                                'reprocessing by payload commit service | '
                                'Tx confirmation for requestID not received : %s',
                                tx_commit_details.payloadCommitId, tx_commit_details.tentativeBlockHeight,
                                request_id
                            )
                            republished_txs.append({
                                'requestID': request_id,
                                'payloadCommitID': payload_commit_obj.commitId,
                                'unconfirmedTentativeHeight': tx_commit_details.tentativeBlockHeight,
                                'resubmittedAtConfirmedBlockHeight': tentative_block_height_event_data
                            })
                            # NOTE: dont remove hash from pending transactions key, instead overwrite with
                            # resubmission attempt block number so the next time an event with
                            # higher tentative block height comes through it does not include this
                            # entry for a resubmission attempt (once payload commit service takes
                            # care of the resubmission, the pending entry is also updated with the
                            # new tx hash)
                            custom_logger.info(
                                'Updating the requestID entry %s at height %s '
                                'to avoid intermediate resubmission attempts',
                                request_id, queued_tentative_height_
                            )
                            update_res = await dag_utils.update_pending_tx_block_touch(
                                pending_tx_set_entry=single_pending_tx_entry,
                                touched_at_block=tentative_block_height_event_data,
                                tentative_block_height=queued_tentative_height_,
                                project_id=project_id,
                                writer_redis_conn=writer_redis_conn,
                                event_data=pending_tx_obj.event_data
                            )
                            if update_res['status']:
                                custom_logger.info(
                                    'Removed requestID entry %s from pending transactions set',
                                    single_pending_tx_entry
                                )
                                custom_logger.info(
                                    'Successfully replaced the requestID entry at height %s '
                                    'to avoid intermediate resubmission attempts in pending txs set',
                                    queued_tentative_height_
                                )
                                replaced_dummy_tx_entries_in_pending_set.append({
                                    'atHeight': queued_tentative_height_
                                })
                            else:
                                if not update_res['results']['zrem']:
                                    custom_logger.warning(
                                        'Could not remove requestID entry %s from pending transactions set',
                                        single_pending_tx_entry
                                    )
                                if not update_res['results']['zadd']:
                                    custom_logger.warning(
                                        'Failed to replace the requestID entry at height %s '
                                        'to avoid intermediate resubmission attempts in pending txs set',
                                        queued_tentative_height_
                                    )

                # check if any self-healing is required/possible
                threshold_before_self_healing_check = 5
                if tentative_block_height_event_data > (
                        finalized_block_height_project + threshold_before_self_healing_check) - 1:
                    # we are looking for pending entry at height (finalized_height+1) which has lastTouchedBlock = -1
                    # yet DAG chain did not move ahead to (finalized_height+1).
                    # this can happen because of inconsistencies in updating the keys in the data store (Redis, atm)
                    # for eg: DAG block did get created in IPFS, the set() on finalized height of the project failed
                    immediate_tx_next_to_finalized_filter = list(filter(
                        lambda x:
                        PendingTransaction.parse_raw(x[0]).lastTouchedBlock == -1 and
                        int(x[1]) == finalized_block_height_project + 1,
                        pending_confirmation_callbacks_txs
                    ))
                    if len(immediate_tx_next_to_finalized_filter) == 1:
                        immediate_tx_pending_obj: PendingTransaction = PendingTransaction.parse_raw(
                            immediate_tx_next_to_finalized_filter[0][0]
                        )
                        custom_logger.info(
                            'Found a pending tx entry with last touched block=-1 at height %s | Finalized project '
                            '%s height: %s | Skipping further resubmission to create DAG chain entry and updating '
                            'project state...',
                            immediate_tx_next_to_finalized_filter[0][1], project_id, finalized_block_height_project
                        )
                        # simulate DAG finalization callback at this height and pass in to DAG block creation routine
                        dag_finalization_cb = DAGFinalizerCallback(
                            txHash=immediate_tx_pending_obj.txHash,
                            requestID=immediate_tx_pending_obj.requestID,
                            event_data=DAGFinalizerCBEventData.parse_obj(immediate_tx_pending_obj.event_data)
                        )
                        all_pending_tx_entries = await reader_redis_conn.zrangebyscore(
                            name=redis_keys.get_pending_transactions_key(project_id),
                            min=finalized_block_height_project + 1,
                            max=float('+inf'),
                            withscores=True
                        )
                        blocks_created = await self._in_order_block_creation_and_state_update(
                            dag_finalizer_callback_obj=dag_finalization_cb,
                            post_finalization_pending_txs=all_pending_tx_entries, parent_cid_height_diff=1,
                            custom_logger_obj=custom_logger, reader_redis_conn=reader_redis_conn,
                            writer_redis_conn=writer_redis_conn, _httpx_client=self._httpx_client)
                        custom_logger.info(
                            'Finished processing self healing DAG block insertion at height %s | '
                            'DAG blocks finalized in total: %s',
                            finalized_block_height_project + 1, blocks_created
                        )
                    elif len(immediate_tx_next_to_finalized_filter) > 1:
                        custom_logger.critical(
                            'Found multiple pending tx entry with last touched block=-1 at height %s immediately '
                            'following finalized project %s height: %s | Entries: %s',
                            finalized_block_height_project + 1, project_id, finalized_block_height_project,
                            immediate_tx_next_to_finalized_filter
                        )
                    elif len(immediate_tx_next_to_finalized_filter) == 0:
                        # missing pending entry at a DAG height == (finalized height + 1) even though callbacks
                        # have arrived for heights > (finalized height + 1)
                        if settings.use_consensus:
                            # epochs that were not submitted. get their epochEndHeight for consensus service query
                            # 1. find missing DAG blocks at event_height > tentative heights > (finalized height + 1)
                            missing_epochs_diag_filter = list(filter(
                                lambda x:
                                    PendingTransaction.parse_raw(x[0]).lastTouchedBlock == -1 and
                                    int(x[1]) > finalized_block_height_project + 1,
                                pending_confirmation_callbacks_txs
                            ))
                            earliest_pending_dag_height_next_to_finalized = int(missing_epochs_diag_filter[0][1])
                            custom_logger.info(
                                'Consensus Self Healing | Found missing pending entries between heights '
                                '%s - %s for project %s | '
                                'Finalized height: %s',
                                finalized_block_height_project+1, earliest_pending_dag_height_next_to_finalized - 1,
                                project_id,
                                finalized_block_height_project
                            )
                            # 2. calculate epoch end heights to be fetched from consensus service corresponding to
                            #    missing DAG height
                            project_epoch_size, project_first_epoch_end_height = await writer_redis_conn.mget(
                                keys=[redis_keys.get_project_epoch_size(project_id),
                                      redis_keys.get_project_first_epoch_end_height(project_id)]
                            )
                            project_epoch_size = int(project_epoch_size)
                            project_first_epoch_end_height = int(project_first_epoch_end_height)
                            # map missing tentative height to expected epochEndHeight
                            epochs_to_fetch = {
                                k: (k-1) * project_epoch_size + project_first_epoch_end_height
                                for k in range(
                                    finalized_block_height_project+1, earliest_pending_dag_height_next_to_finalized
                                )
                            }
                            consensus_snapshots_fetch_tasks = [
                                self._httpx_wrap_call(
                                    url=urljoin(settings.consensus_config.service_url, '/epochStatus'),
                                    json_body=SnapshotBase(
                                        epoch=EpochBase(begin=y - project_epoch_size + 1, end=y),
                                        projectID=project_id,
                                        instanceID=settings.instance_id
                                    ).dict()
                                )
                                for x, y in epochs_to_fetch.items()
                            ]
                            # 3. query consensus service for snapshots and create DAG block
                            consensus_snapshots_response = await asyncio.gather(
                                *consensus_snapshots_fetch_tasks,
                                return_exceptions=True
                            )
                            # fill in pending tx entries according to consensus snapshot CID returned
                            # or set data to null for the epochs where no consensus or submission were found
                            tentative_height_to_cid_map = dict()
                            for tentative_height, each_consensus_response in zip(
                                    epochs_to_fetch.keys(), consensus_snapshots_response
                            ):
                                custom_logger.debug('tentative height: %s  consensus service response: %s', tentative_height, each_consensus_response)
                                if isinstance(consensus_snapshots_response, Exception):
                                    custom_logger.warning(
                                        'Consensus Self Healing | Project %s | Exception fetching snapshot from '
                                        'consensus service against epoch end height %s, expected tentative height: %s ',
                                        project_id,
                                        epochs_to_fetch[tentative_height],
                                        tentative_height,
                                        consensus_snapshots_response
                                    )
                                    tentative_height_to_cid_map[tentative_height] = f'null_{epochs_to_fetch[tentative_height]}'
                                    continue
                                try:
                                    parsed_snapshot_response = SubmissionResponse.parse_obj(each_consensus_response)
                                except Exception as e:
                                    custom_logger.warning(
                                        'Consensus Self Healing | Project %s | Exception converting response to '
                                        'data model against epoch end height %s, expected tentative height: %s | %s',
                                        project_id,
                                        epochs_to_fetch[tentative_height],
                                        tentative_height,
                                        e,
                                        exc_info=True
                                    )
                                    tentative_height_to_cid_map[tentative_height] = f'null_{epochs_to_fetch[tentative_height]}'
                                    continue

                                custom_logger.info(
                                    'Consensus Self Healing | Project %s | Fetched snapshot CID %s from '
                                    'consensus service against epoch end height %s, expected tentative height: %s',
                                    project_id,
                                    parsed_snapshot_response.finalizedSnapshotCID, epochs_to_fetch[tentative_height],
                                    tentative_height
                                )

                                if parsed_snapshot_response.status == EpochConsensusStatus.consensus_achieved:
                                    tentative_height_to_cid_map[
                                        tentative_height] = parsed_snapshot_response.finalizedSnapshotCID
                                else:
                                    custom_logger.warning(
                                        'Consensus Self Healing | Project %s| Consensus service reports inconsistent status '
                                        'against epoch end height %s, expected tentative height: %s',
                                        project_id,
                                        epochs_to_fetch[tentative_height],
                                        tentative_height,
                                        parsed_snapshot_response
                                    )
                                    tentative_height_to_cid_map[tentative_height] = f'null_{epochs_to_fetch[tentative_height]}'
                            dummy_tx_hash = '0x' + '0' * 160
                            dummy_api_hash = '0x' + '0' * 256
                            dummy_payload_commit_id = '0x' + '0' * 256
                            await writer_redis_conn.zadd(
                                name=redis_keys.get_payload_cids_key(project_id),
                                mapping={str(v): k for k, v in tentative_height_to_cid_map.items()}
                            )
                            for i, (t, cid) in enumerate(tentative_height_to_cid_map.items()):
                                # add new pending tx entries against missing epochs
                                pending_tx_entry_missing_epoch = PendingTransaction(
                                    txHash=dummy_tx_hash,
                                    requestID=str(uuid.uuid4()),
                                    lastTouchedBlock=-1,
                                    event_data=AuditRecordTxEventData(
                                        txHash=dummy_tx_hash,
                                        apiKeyHash=dummy_api_hash,
                                        timestamp=int(time.time()),
                                        payloadCommitId=dummy_payload_commit_id,
                                        snapshotCid=cid,
                                        tentativeBlockHeight=t,
                                        skipAnchorProof=False,
                                        projectId=project_id
                                    )
                                )
                                await writer_redis_conn.zadd(
                                    name=redis_keys.get_pending_transactions_key(project_id),
                                    mapping={
                                        pending_tx_entry_missing_epoch.json(): t
                                    }
                                )
                                custom_logger.info(
                                    'Consensus Self Healing| Project %s | Added pending tx entry at height %s '
                                    'against epoch %s: %s',
                                    project_id,
                                    t,
                                    epochs_to_fetch[t],
                                    pending_tx_entry_missing_epoch
                                )
                                if i == 0:
                                    first_pending_entry = pending_tx_entry_missing_epoch
                            dummy_event_data = DAGFinalizerCBEventData(
                                apiKeyHash=dummy_api_hash,
                                tentativeBlockHeight=min(epochs_to_fetch.keys()),
                                projectId=project_id,
                                snapshotCid=tentative_height_to_cid_map[min(epochs_to_fetch.keys())],
                                payloadCommitId=dummy_payload_commit_id,
                                timestamp=int(time.time())
                            )
                            dag_finalization_cb = DAGFinalizerCallback(
                                txHash=dummy_tx_hash,
                                requestID=first_pending_entry.requestID,
                                event_data=dummy_event_data
                            )
                            post_pending_tx_entries = await reader_redis_conn.zrangebyscore(
                                name=redis_keys.get_pending_transactions_key(project_id),
                                min=min(epochs_to_fetch.keys()) + 1,
                                max=float('+inf'),
                                withscores=True
                            )
                            blocks_created = await self._in_order_block_creation_and_state_update(
                                dag_finalizer_callback_obj=dag_finalization_cb,
                                post_finalization_pending_txs=post_pending_tx_entries, parent_cid_height_diff=1,
                                custom_logger_obj=custom_logger, reader_redis_conn=reader_redis_conn,
                                writer_redis_conn=writer_redis_conn, _httpx_client=self._httpx_client)
                            custom_logger.info(
                                'Consensus Self Healing| Project %s | '
                                'Finished processing self healing DAG block insertion beginning height %s | '
                                'DAG blocks finalized in total: %s',
                                project_id, finalized_block_height_project + 1,
                                blocks_created
                            )
                            await self._dispatch_healing_notifications(
                                project_id,
                                tentative_height_to_cid_map,
                                epochs_to_fetch
                            )
                        else:
                            custom_logger.critical(
                                'Standalone system | Missing pending tx entry at tentative height %s following '
                                'finalized height %s for project %s',
                                finalized_block_height_project + 1,
                                finalized_block_height_project,
                                project_id
                            )
        elif tentative_block_height_event_data == finalized_block_height_project + 1:
            """
                An event which is in-order has arrived. Create a dag block for this event
                and process all other pending events for this project
            """
            all_pending_tx_entries = await reader_redis_conn.zrangebyscore(
                name=redis_keys.get_pending_transactions_key(project_id),
                min=float('-inf'),
                max=float('+inf'),
                withscores=True
            )
            # custom_logger.debug('All pending transactions for project %s in key %s : %s',
            #                   project_id, redis_keys.get_pending_transactions_key(project_id),
            #                   all_pending_tx_entries)
            is_pending = False
            for k in all_pending_tx_entries:
                pending_tx_obj: PendingTransaction = PendingTransaction.parse_raw(k[0])
                if pending_tx_obj.requestID == request_id and int(k[1]) == event_data.event_data.tentativeBlockHeight:
                    is_pending = True
                    break

            if not is_pending:
                custom_logger.error(
                    "Discarding event because requestID %s not in pending transactions | Project ID %s | %s",
                    request_id, project_id, event_data
                )
                discarded_transactions_key = redis_keys.get_discarded_transactions_key(project_id)
                _ = await writer_redis_conn.zadd(
                    name=discarded_transactions_key,
                    mapping={request_id: tentative_block_height_event_data}
                )
            else:
                blocks_created = await self._in_order_block_creation_and_state_update(
                    dag_finalizer_callback_obj=event_data, post_finalization_pending_txs=filter(
                        lambda x: int(x[1]) > tentative_block_height_event_data,
                        all_pending_tx_entries
                    ), parent_cid_height_diff=1, custom_logger_obj=custom_logger, reader_redis_conn=reader_redis_conn,
                    writer_redis_conn=writer_redis_conn, _httpx_client=self._httpx_client)
                custom_logger.info(
                    'Finished processing in order DAG block insertion at height %s | '
                    'DAG blocks finalized in total: %s',
                    tentative_block_height_event_data, blocks_created
                )

    async def _dispatch_healing_notifications(self, project_id, tentative_height_cid_map, tentative_height_epoch_match):
        null_assigned_epochs = list(map(
            lambda x: tentative_height_epoch_match[x[0]],
            filter(
                lambda x: 'null' in x[1],
                tentative_height_cid_map.items()
            )
        ))
        cid_finalized_epochs = list(map(
            lambda x: tentative_height_epoch_match[x[0]],
            filter(
                lambda x: 'null' not in x[1],
                tentative_height_cid_map.items()
            )
        ))
        tasks = [
            self._httpx_client.post(
                url=urljoin(f'http://{settings.host}:{settings.port}', '/reportIssue'),
                json=SnapshotterIssue(
                    instanceID=settings.instance_id,
                    severity=SnapshotterIssueSeverity.medium if idx == 1 else SnapshotterIssueSeverity.high,
                    issueType='MISSED_SNAPSHOT' if idx == 1 else 'SKIP_EPOCH',
                    projectID=project_id,
                    epochs=k,
                    timeOfReporting=int(time.time())
                ).dict()
            ) for idx, k in enumerate([null_assigned_epochs, cid_finalized_epochs])
        ]
        await asyncio.gather(*tasks)

    async def _on_rabbitmq_message(self, message: IncomingMessage):
        event_data = DAGFinalizerCallback.parse_raw(message.body)
        asyncio.ensure_future(self._payload_to_dag_processor_task(event_data))
        await message.ack()

    @retry(
        reraise=True, wait=wait_random_exponential(multiplier=1, max=20),
        # Randomly wait up to 2^x * 1 seconds between each retry until the range reaches 60 seconds
        # then randomly up to 20 seconds afterwards
        stop=stop_after_attempt(5),
        retry=retry_if_exception_type(httpx_exceptions.TransportError) | retry_if_exception_type(json.JSONDecodeError)
    )
    async def _httpx_wrap_call(self, url, json_body):
        resp = await self._httpx_client.post(
            url=url,
            json=json_body
        )
        return resp.json()

    async def main(self):
        ev_loop = asyncio.get_running_loop()
        self._rmq_connection_pool = Pool(get_rabbitmq_connection, max_size=100, loop=ev_loop)
        self._rmq_channel_pool = Pool(
            partial(get_rabbitmq_channel, self._rmq_connection_pool), max_size=200, loop=ev_loop
        )

        self._aioredis_pool = RedisPool()
        await self._aioredis_pool.populate()
        self._writer_redis_pool = self._aioredis_pool.writer_redis_pool
        self._reader_redis_pool = self._aioredis_pool.reader_redis_pool
        self._async_transport = AsyncHTTPTransport(
            limits=Limits(max_connections=300, max_keepalive_connections=100, keepalive_expiry=30)
        )
        self._httpx_client = AsyncClient(
            timeout=Timeout(timeout=5.0),
            follow_redirects=False,
            transport=self._async_transport
        )
        self._ipfs_singleton = AsyncIPFSClientSingleton()
        await self._ipfs_singleton.init_sessions()
        self._ipfs_writer_client = self._ipfs_singleton._ipfs_write_client
        self._ipfs_reader_client = self._ipfs_singleton._ipfs_read_client
        async with self._rmq_channel_pool.acquire() as channel:
            await channel.set_qos(500)
            q_obj = await channel.get_queue(
                name=self._q,
                ensure=False
            )
            self._logger.debug(f'Consuming queue {self._q} with routing key {self._rmq_routing}...')
            await q_obj.consume(self._on_rabbitmq_message)


if __name__ == '__main__':
    soft, hard = resource.getrlimit(resource.RLIMIT_NOFILE)
    resource.setrlimit(resource.RLIMIT_NOFILE, (settings.rlimit['file_descriptors'], hard))
    dag_finalization_cb_processor = DAGFinalizationCallbackProcessor()
    asyncio.ensure_future(dag_finalization_cb_processor.main())
    ev_loop = asyncio.get_event_loop()
    try:
        ev_loop.run_forever()
    finally:
        ev_loop.close()
