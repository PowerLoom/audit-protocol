import aioredis
from fastapi import FastAPI, Request, Response, Header, BackgroundTasks
from utils.rabbitmq_utils import get_rabbitmq_connection, get_rabbitmq_channel
from config import settings
from utils import redis_keys
from utils import helper_functions
from utils import diffmap_utils
from utils import dag_utils
from utils.ipfs_async import client as ipfs_client
from utils.redis_conn import RedisPool
from aio_pika import ExchangeType, DeliveryMode, Message
from functools import partial
from aio_pika.pool import Pool
from data_models import PayloadCommit, PendingTransaction
import asyncio
import aiohttp
import logging
import sys
import json
import os



app = FastAPI()

formatter = logging.Formatter(u"%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s")

stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setLevel(logging.DEBUG)
# stdout_handler.setFormatter(formatter)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setLevel(logging.ERROR)
# stderr_handler.setFormatter(formatter)
rest_logger = logging.getLogger(__name__)
rest_logger.setLevel(logging.DEBUG)
rest_logger.addHandler(stdout_handler)
rest_logger.addHandler(stderr_handler)

REDIS_WRITER_CONN_CONF = {
    "host": settings.redis.host,
    "port": settings.redis.port,
    "password": settings.redis.password,
    "db": settings.redis.db
}

REDIS_READER_CONN_CONF = {
    "host": settings.redis_reader.host,
    "port": settings.redis_reader.port,
    "password": settings.redis_reader.password,
    "db": settings.redis_reader.db
}


@app.on_event('startup')
async def startup_boilerplate():
    app.rmq_connection_pool = Pool(get_rabbitmq_connection, max_size=5, loop=asyncio.get_running_loop())
    app.rmq_channel_pool = Pool(
        partial(get_rabbitmq_channel, app.rmq_connection_pool), max_size=20, loop=asyncio.get_running_loop()
    )
    try:
        os.stat(os.getcwd() + '/bloom_filter_objects')
    except:
        os.mkdir(os.getcwd() + '/bloom_filter_objects')

    try:
        os.stat(os.getcwd() + '/containers')
    except:
        os.mkdir(os.getcwd() + '/containers')
    app.aioredis_pool = RedisPool()
    await app.aioredis_pool.populate()
    app.writer_redis_pool = app.aioredis_pool.writer_redis_pool
    app.reader_redis_pool = app.aioredis_pool.reader_redis_pool
    app.aiohttp_client_session = aiohttp.ClientSession(
        timeout=aiohttp.ClientTimeout(
            sock_connect=settings.aiohtttp_timeouts.sock_connect,
            sock_read=settings.aiohtttp_timeouts.sock_read,
            connect=settings.aiohtttp_timeouts.connect
        )
    )


@app.post('/')
async def create_dag(
        request: Request,
        response: Response,
        x_hook_signature: str = Header(None),
):
    reader_redis_conn: aioredis.Redis = request.app.reader_redis_pool
    writer_redis_conn: aioredis.Redis = request.app.writer_redis_pool
    event_data = await request.json()
    # Verify the payload that has arrived.
    if x_hook_signature:
        is_safe = dag_utils.check_signature(event_data, x_hook_signature)
        if is_safe:
            rest_logger.debug("The arriving payload has been verified")
            pass
        else:
            rest_logger.debug("Recieved an wrong signature payload")
            return dict()
    if 'event_name' in event_data.keys():
        if event_data['event_name'] == 'RecordAppended':
            rest_logger.debug(event_data)

            """ Get data from the event """
            project_id = event_data['event_data']['projectId']
            tentative_block_height_event_data = int(event_data['event_data']['tentativeBlockHeight'])

            # Get the max block height(finalized after all error corrections and reorgs) for the project_id
            finalized_block_height_project = await helper_functions.get_block_height(
                project_id=project_id,
                reader_redis_conn=reader_redis_conn
            )

            rest_logger.debug(
                "Event Data Tentative Block Height: %s | Finalized Project %s Block Height: %s",
                tentative_block_height_event_data, project_id, finalized_block_height_project
            )
            rest_logger.debug(tentative_block_height_event_data)
            rest_logger.debug(finalized_block_height_project)

            tentative_block_height_cached = await helper_functions.get_tentative_block_height(
                project_id=project_id,
                reader_redis_conn=reader_redis_conn
            )
            # retrieve callback URL for project ID
            cb_url = await reader_redis_conn.get(f'powerloom:project:{project_id}:callbackURL')
            if tentative_block_height_event_data <= finalized_block_height_project:
                rest_logger.debug("Discarding event at height %s | %s", tentative_block_height_event_data, event_data)
                rest_logger.debug(tentative_block_height_event_data)
                redis_output = await dag_utils.discard_event(
                    project_id=project_id,
                    payload_commit_id=event_data['event_data']['payloadCommitId'],
                    tx_hash=event_data['txHash'],
                    payload_cid=event_data['event_data']['snapshotCid'],
                    tentative_block_height=tentative_block_height_event_data,
                )
                # rest_logger.debug("Redis operations output: ")
                # rest_logger.debug(redis_output)
                return dict()

            elif tentative_block_height_event_data > finalized_block_height_project + 1:
                rest_logger.debug("An out of order event arrived | Project ID %s | %s", project_id, event_data)

                discarded_transactions_key = f"projectID:{project_id}:discardedTransactions"

                rest_logger.debug("Checking if txHash %s is in the list of pending Transactions", event_data['txHash'])

                _ = await reader_redis_conn.zrangebyscore(
                    key=redis_keys.get_pending_transactions_key(project_id),
                    min=float('-inf'),
                    max=float('+inf'),
                    withscores=False
                )
                is_pending = False
                for k in _:
                    pending_tx_obj: PendingTransaction = PendingTransaction.parse_raw(k)
                    if pending_tx_obj.txHash == event_data['txHash']:
                        is_pending = True
                        break

                if not is_pending:
                    rest_logger.debug(
                        "Discarding event because tx Hash not in pending transactions | Project ID %s | %s",
                        project_id, event_data
                    )
                    _ = await dag_utils.clear_payload_commit_data(
                        project_id=project_id,
                        tx_hash=event_data['txHash'],
                        tentative_height_pending_tx_entry=0,  # this has no effect, and it is fine
                        payload_commit_id=event_data['event_data']['payloadCommitId']
                    )

                    _ = await writer_redis_conn.zadd(
                        key=discarded_transactions_key,
                        score=tentative_block_height_event_data,
                        member=event_data['txHash']
                    )

                    return dict()

                rest_logger.debug(
                    "Saving Event data for tentative height: %s | %s ",
                    tentative_block_height_event_data, event_data
                )
                await dag_utils.save_event_data(
                    event_data=event_data,
                    writer_redis_conn=writer_redis_conn,
                )
                # NOTE: tentative entries in pendingBlocks are a graduation from pendingTransactions set
                #       find out difference between zsets of pending txs and pending blocks
                #       this will tell us the tentative heights against which no callbacks have arrived,
                #       yet we have tx hash identifiers recorded
                #       set of pending txs tentative heights will always be superset of the same against pending blocks
                pending_confirmation_callbacks_txs = await reader_redis_conn.zrangebyscore(
                    key=redis_keys.get_pending_transactions_key(project_id),
                    min=finalized_block_height_project + 1,
                    max=tentative_block_height_event_data,
                    withscores=True
                )
                pending_dag_chain_insertion_blocks = await reader_redis_conn.zrangebyscore(
                    key=redis_keys.get_pending_blocks_key(project_id),
                    min=finalized_block_height_project + 1,
                    max=tentative_block_height_event_data,
                    withscores=True
                )
                # filter out resubmitted transactions so they are not counted against gaps and further resubmission
                # dont touch them if they are 'fresh' as well (lastTouchedBlock = 0)
                pending_confirmation_callbacks_txs_filtered = list(filter(
                    lambda x:
                        PendingTransaction.parse_raw(x[0]).lastTouchedBlock + 3 <= tentative_block_height_event_data or
                        PendingTransaction.parse_raw(x[0]).lastTouchedBlock == 0,
                    pending_confirmation_callbacks_txs
                ))
                rest_logger.info(
                    'Pending transactions qualified for resubmission on account of being unconfirmed or '
                    'not yet considered since the last resubmission attempt: %s',
                    pending_confirmation_callbacks_txs_filtered
                )
                callback_unseen_heights = set(
                    map(
                        lambda x: int(x[1]),
                        pending_confirmation_callbacks_txs_filtered
                    )
                ).difference(
                    set(
                        map(
                            lambda x: int(x[1]),
                            pending_dag_chain_insertion_blocks
                        )
                    )
                )
                callback_unseen_txs = set()
                if callback_unseen_heights:
                    callback_unseen_heights_l = sorted(callback_unseen_heights)
                    [callback_unseen_txs.add(x.decode('utf-8')) for x in
                     await reader_redis_conn.zrangebyscore(
                         key=redis_keys.get_pending_transactions_key(project_id),
                         min=min(callback_unseen_heights),
                         max=max(callback_unseen_heights),
                         withscores=False
                     )]
                    rest_logger.debug(
                        'Found following tentative block heights '
                        'pending confirmed callbacks: %s | Corresponding txs: %s',
                        callback_unseen_heights, callback_unseen_txs
                    )
                    # NOTE: if we do not find a pending entry against finalized height+1, it is a bigger problem.
                    #       Ideally such a situation should never arise.
                    if callback_unseen_heights_l[0] != finalized_block_height_project + 1:
                        # TODO: in such cases the snapshots being sent from a subscribed project need to be resubmitted
                        #       beginning from the snapshot corresponding to last finalized block height+1.
                        #       This may be reconstructed as audit protocol informs the last legible state data
                        #       from the tx input data stored against the expected tx hash
                        #       (to be referenced from payload commit ID details). Similar strategy would apply for
                        #       scenarios tagged with critical logs below as well
                        rest_logger.critical(
                            'Possible irrecoverable gap in DAG chain creation | Last finalized height: %s | No pending '
                            'entry found against callback confirmation for tentative height: %s | ' 
                            'pending confirmed callbacks: %s | Corresponding txs: %s',
                            finalized_block_height_project, finalized_block_height_project + 1,
                            callback_unseen_heights, callback_unseen_txs
                        )
                    else:
                        # check if pending tx confirmation callbacks to earliest known pending block is a
                        # continuous sequence when compared against the corresponding tentative block heights
                        sequential = True
                        failed_sequence_height = None
                        for idx, h in enumerate(callback_unseen_heights_l):
                            if idx == 0:
                                continue
                            if h - callback_unseen_heights_l[idx - 1] != 1:
                                sequential = False
                                # tentative block height
                                failed_sequence_height = callback_unseen_heights_l[idx - 1]
                                break
                        if callback_unseen_heights_l[-1] + 1 != int(pending_dag_chain_insertion_blocks[0][1]):
                            rest_logger.warning(
                                'Gap between max tentative block height against unreceived webhook callbacks and min '
                                'tentative block height against received callbacks but pending entry into chain '
                                '| %s | %s',
                                callback_unseen_heights_l,
                                list(map(lambda x: int(x[1]), pending_dag_chain_insertion_blocks))
                            )
                            sequential = False
                        if not sequential:
                            rest_logger.critical(
                                'Possible irrecoverable gap in DAG chain creation | Last finalized height: %s | Found'
                                ' possible gap in transactions pending callback at tentative height: %s while scanning for'
                                ' continuous pending range of of callbacks between tentative heights %s - %s |'
                                ' Tentative block heights without callbacks received: %s | Tentative blockheights'
                                ' pending insertion into DAG chain but callbacks were received: %s',
                                finalized_block_height_project, failed_sequence_height,
                                finalized_block_height_project + 1, tentative_block_height_event_data,
                                callback_unseen_heights_l, list(map(lambda x: int(x[1]), pending_dag_chain_insertion_blocks))
                            )
                        else:
                            if len(pending_dag_chain_insertion_blocks)+len(callback_unseen_heights_l) \
                                    >= settings.max_pending_payload_commits:
                                rest_logger.info(
                                    'Preparing to send out txs for reprocessing for the ones that have not '
                                    'received any callbacks (possibility of dropped txs) : %s',
                                    pending_confirmation_callbacks_txs_filtered
                                )
                                # map their tentative block height scores in the zset to the actual member entry in bytes
                                pending_confirmation_callbacks_txs_filtered_map = {
                                    int(x[1]): x[0] for x in pending_confirmation_callbacks_txs_filtered
                                }
                                # send out txs for reprocessing
                                async with request.app.rmq_channel_pool.acquire() as channel:
                                    exchange = await channel.declare_exchange(
                                        settings.rabbitmq.setup['core']['exchange'],
                                        ExchangeType.DIRECT
                                    )
                                    for queued_tentative_height_ in callback_unseen_heights_l:
                                        # get the tx hash from the filtered set of qualified pending transactions
                                        tx_hash_obj: PendingTransaction = PendingTransaction.parse_raw(
                                            pending_confirmation_callbacks_txs_filtered_map[queued_tentative_height_]
                                        )
                                        tx_hash = tx_hash_obj.txHash
                                        # fetch transaction input data
                                        s = await reader_redis_conn.get(
                                            redis_keys.get_pending_tx_input_data_key(tx_hash)
                                        )
                                        if not s:
                                            rest_logger.critical(
                                                'Possible irrecoverable gap in DAG chain creation | '
                                                'Did not find cached input data against past tx: %s | '
                                                'Last finalized height: %s'
                                                ' Tentative block heights without callbacks received: %s | Tentative blockheights'
                                                ' pending insertion into DAG chain but callbacks were received: %s',
                                                tx_hash,
                                                finalized_block_height_project,
                                                callback_unseen_heights_l,
                                                list(map(lambda x: int(x[1]), pending_dag_chain_insertion_blocks))
                                            )
                                            break
                                        # send back to payload commit service
                                        tx_commit_details = json.loads(s)
                                        payload_commit_obj = PayloadCommit(
                                            **{
                                                'projectId': project_id,
                                                'commitId': tx_commit_details['payloadCommitId'],
                                                'snapshotCID': tx_commit_details['snapshotCid'],
                                                'tentativeBlockHeight': tx_commit_details['tentativeBlockHeight'],
                                                'apiKeyHash': tx_commit_details['apiKeyHash'],
                                                'resubmitted': True,
                                                'resubmissionBlock': tentative_block_height_event_data
                                            }
                                        )
                                        message = Message(
                                            payload_commit_obj.json().encode('utf-8'),
                                            delivery_mode=DeliveryMode.PERSISTENT,
                                        )
                                        await exchange.publish(
                                            message=message,
                                            routing_key='commit-payloads'
                                        )
                                        rest_logger.debug(
                                            'Re-Published payload against commit ID %s , tentative block height %s for '
                                            'reprocessing by payload commit service | '
                                            'Previous tx not found for DAG entry: %s',
                                            tx_commit_details['payloadCommitId'], tx_commit_details['tentativeBlockHeight'],
                                            tx_hash
                                        )
                                        # NOTE: dont remove hash from pending transactions key, instead overwrite with
                                        #       resubmission attempt block number so the next time an event with
                                        #       higher tentative block height comes through it does not include this
                                        #       entry for a resubmission attempt (once payload commit service takes
                                        #       care of the resubmission, the pending entry is also updated with the
                                        #       new tx hash)
                                        pending_tx_entry = pending_confirmation_callbacks_txs_filtered_map[queued_tentative_height_]
                                        _ = await writer_redis_conn.zrem(
                                            key=redis_keys.get_pending_transactions_key(project_id),
                                            member=pending_tx_entry
                                        )
                                        if _:
                                            rest_logger.info(
                                                'Removed tx hash entry %s from pending transactions set',
                                                pending_tx_entry
                                            )
                                            dummy_tx_entry = PendingTransaction(
                                                txHash=tx_hash, lastTouchedBlock=tentative_block_height_event_data
                                            )
                                            rest_logger.info(
                                                'Replacing with dummy tx hash entry %s at height %s '
                                                'to avoid intermediate resubmission attempts',
                                                dummy_tx_entry, queued_tentative_height_
                                            )
                                            res = await writer_redis_conn.zadd(
                                                key=redis_keys.get_pending_transactions_key(project_id),
                                                member=dummy_tx_entry.json(),
                                                score=queued_tentative_height_
                                            )
                                            if res:
                                                rest_logger.info(
                                                    'Successfully replaced with dummy tx hash entry %s at height %s '
                                                    'to avoid intermediate resubmission attempts in pending txs set',
                                                    dummy_tx_entry, queued_tentative_height_
                                                )
                                            else:
                                                rest_logger.warning(
                                                    'Failed to replace with dummy tx hash entry %s at height %s '
                                                    'to avoid intermediate resubmission attempts in pending txs set',
                                                    dummy_tx_entry, queued_tentative_height_
                                                )
                                        else:
                                            rest_logger.warning(
                                                'Could not remove tx hash entry %s from pending transactions set',
                                                pending_tx_entry
                                            )

            elif tentative_block_height_event_data == finalized_block_height_project + 1:
                """
                    An event which is in-order has arrived. Create a dag block for this event
                    and process all other pending events for this project
                """

                _dag_cid, dag_block = await dag_utils.create_dag_block(
                    tx_hash=event_data['txHash'],
                    project_id=project_id,
                    tentative_block_height=tentative_block_height_event_data,
                    payload_cid=event_data['event_data']['snapshotCid'],
                    timestamp=event_data['event_data']['timestamp']
                )
                rest_logger.info('Created DAG block with CID %s at height %s', _dag_cid, tentative_block_height_event_data)
                # clear payload commit ID log on success
                _ = await dag_utils.clear_payload_commit_processing_logs(
                    project_id=project_id,
                    payload_commit_id=event_data['event_data']['payloadCommitId'],
                    writer_redis_conn=writer_redis_conn
                )
                if _:
                    rest_logger.debug(
                        'Cleared processing logs for payload commit ID: %s', event_data['event_data']['payloadCommitId']
                    )
                else:
                    rest_logger.debug(
                        'Not sure if processing logs cleared or nonexistent for payload commit ID: %s | '
                        'Redis return: %s',
                        event_data['event_data']['payloadCommitId'],
                        _
                    )
                # process diff at this new block height
                try:
                    diff_map = await diffmap_utils.calculate_diff(
                        dag_cid=_dag_cid,
                        dag=dag_block,
                        project_id=project_id,
                        ipfs_client=ipfs_client
                        # writer_redis_conn=writer_redis_conn
                    )
                except json.decoder.JSONDecodeError as jerr:
                    rest_logger.debug("There was an error while decoding the JSON data: %s", jerr, exc_info=True)
                else:
                    rest_logger.debug("The diff map retrieved: %s", diff_map)

                # send commit ID confirmation callback
                if cb_url:
                    await send_commit_callback(
                        aiohttp_session=request.app.aiohttp_client_session,
                        url=cb_url,
                        payload={
                            'commitID': event_data['event_data']['payloadCommitId'],
                            'projectID': project_id,
                            'status': True
                        }
                    )

                await dag_utils.clear_payload_commit_data(
                    project_id=project_id,
                    payload_commit_id=event_data['event_data']['payloadCommitId'],
                    tx_hash=event_data['txHash'],
                    tentative_height_pending_tx_entry=tentative_block_height_event_data
                )

                """ retrieve all list of all the payload_commit_ids from the pendingBlocks zset """
                all_pending_blocks = await reader_redis_conn.zrangebyscore(
                    key=redis_keys.get_pending_blocks_key(project_id=project_id),
                    min=tentative_block_height_event_data+1,
                    # `max` arg lef out  # defaults to inf
                    withscores=True
                )
                if len(all_pending_blocks) > 0:
                    rest_logger.info(
                        'Found blocks pending creation at following heights %s',
                        list(map(lambda x: x[1], all_pending_blocks))
                    )
                    """ There are some pending blocks left """
                    cur_max_height_project = tentative_block_height_event_data + 1
                    for _payload_commit_id, _tt_block_height in all_pending_blocks:
                        _payload_commit_id = _payload_commit_id.decode('utf-8')
                        _tt_block_height = int(_tt_block_height)

                        if _tt_block_height == cur_max_height_project + 1:
                            rest_logger.debug("Processing queued block at tentative_block_height: %s", _tt_block_height)

                            """ Retrieve the event data for the payload_commit_id """
                            event_data_key = f"eventData:{_payload_commit_id}"
                            out = await reader_redis_conn.hgetall(key=event_data_key)

                            _event_data = {k.decode('utf-8'): v.decode('utf-8') for k, v in out.items()}
                            _tx_hash = _event_data['txHash']

                            """ Create the dag block for this event """
                            _dag_cid, dag_block = await dag_utils.create_dag_block(
                                tx_hash=_event_data['txHash'],
                                project_id=project_id,
                                tentative_block_height=_tt_block_height,
                                payload_cid=_event_data['snapshotCid'],
                                timestamp=_event_data['timestamp'],
                            )
                            rest_logger.info('Created DAG block with CID %s at height %s', _dag_cid, _tt_block_height)

                            cur_max_height_project = _tt_block_height
                            try:
                                diff_map = await diffmap_utils.calculate_diff(
                                    dag_cid=_dag_cid,
                                    dag=dag_block,
                                    project_id=project_id,
                                    ipfs_client=ipfs_client
                                    # writer_redis_conn=writer_redis_conn
                                )
                            except json.decoder.JSONDecodeError as jerr:
                                rest_logger.debug("There was an error while decoding the JSON data: %s", jerr, exc_info=True)
                            else:
                                rest_logger.debug("The diff map retrieved: %s", diff_map)

                            await dag_utils.clear_payload_commit_data(
                                project_id=project_id,
                                payload_commit_id=_payload_commit_id,
                                tx_hash=_tx_hash,
                                tentative_height_pending_tx_entry=_tt_block_height
                            )

                            _ = await dag_utils.clear_payload_commit_processing_logs(
                                project_id=project_id,
                                payload_commit_id=_payload_commit_id,
                                writer_redis_conn=writer_redis_conn
                            )
                            if _:
                                rest_logger.debug(
                                    'Cleared processing logs for payload commit ID: %s',
                                    _payload_commit_id
                                )
                            else:
                                rest_logger.debug(
                                    'Not sure if processing logs cleared or nonexistent for payload commit ID: %s | '
                                    'Redis return: %s',
                                    _payload_commit_id,
                                    _
                                )
                            # send commit ID confirmation callback
                            if cb_url:
                                await send_commit_callback(
                                    aiohttp_session=request.app.aiohttp_client_session,
                                    url=cb_url,
                                    payload={
                                        'commitID': _payload_commit_id,
                                        'projectID': project_id,
                                        'status': True
                                    }
                                )
    response.status_code = 200
    return dict()


async def send_commit_callback(aiohttp_session: aiohttp.ClientSession, url, payload):
    if type(url) is bytes:
        url = url.decode('utf-8')
    try:
        async with aiohttp_session.post(url=url, json=payload) as resp:
            json_response = await resp.json()
    except Exception as e:
        rest_logger.error('Failed to push callback for commit ID')
        rest_logger.error({'url': url, 'payload': payload})
        rest_logger.error(e, exc_info=True)