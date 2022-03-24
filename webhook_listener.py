from fastapi import FastAPI, Request, Response, Header, BackgroundTasks
import aiohttp
import time
import aioredis
from config import settings
import logging
import sys
import json
import os

from utils import redis_keys
from utils import helper_functions
from utils import diffmap_utils
from utils import dag_utils
from utils.ipfs_async import client as ipfs_client
from utils.redis_conn import RedisPool


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
    reader_redis_conn = request.app.reader_redis_pool
    writer_redis_conn = request.app.writer_redis_pool
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

            # Get the max block height for the project_id
            max_block_height_project = await helper_functions.get_block_height(
                project_id=project_id,
                reader_redis_conn=reader_redis_conn
            )

            rest_logger.debug("Tentative Block Height and Block Height")
            rest_logger.debug(tentative_block_height_event_data)
            rest_logger.debug(max_block_height_project)

            tentative_block_height_cached = await helper_functions.get_tentative_block_height(
                project_id=project_id,
                reader_redis_conn=reader_redis_conn
            )
            # retrieve callback URL for project ID
            cb_url = await reader_redis_conn.get(f'powerloom:project:{project_id}:callbackURL')
            if (tentative_block_height_cached == max_block_height_project) or (tentative_block_height_event_data <= max_block_height_project):
                rest_logger.debug("Discarding event at height:")
                rest_logger.debug(tentative_block_height_event_data)
                redis_output = await dag_utils.discard_event(
                    project_id=project_id,
                    payload_commit_id=event_data['event_data']['payloadCommitId'],
                    tx_hash=event_data['txHash'],
                    payload_cid=event_data['event_data']['snapshotCid'],
                    tentative_block_height=tentative_block_height_event_data,
                )
                rest_logger.debug("Redis operations output: ")
                rest_logger.debug(redis_output)
                return dict()

            elif tentative_block_height_event_data > max_block_height_project + 1:
                rest_logger.debug("An out of order event arrived...")
                rest_logger.debug("Checking if that txHash is in the list of pending Transactions")

                discarded_transactions_key = f"projectID:{project_id}:discardedTransactions"
                pending_transactions_key = f"projectID:{project_id}:pendingTransactions"

                out = await reader_redis_conn.zscore(pending_transactions_key, event_data['txHash'])
                if out is None:
                    rest_logger.debug("Discarding the event...")
                    _ = await dag_utils.clear_payload_commit_data(
                        project_id=project_id,
                        tx_hash=event_data['txHash'],
                        payload_commit_id=event_data['event_data']['payloadCommitId']
                    )

                    _ = await writer_redis_conn.zadd(
                        key=discarded_transactions_key,
                        score=tentative_block_height_event_data,
                        member=event_data['txHash']
                    )

                    return dict()

                if (tentative_block_height_cached - max_block_height_project) >= settings.max_pending_events:

                    target_tt_block_height = max_block_height_project + 1

                    # FIXME: we take the zscore of the tx hash in `out` yet it is not used
                    #   instead, this is terrible practice to be shoving the arrived event data at the height
                    #   (last known block height+1)
                    rest_logger.debug("Creating DAG for the arrived event at height: ")
                    rest_logger.debug(target_tt_block_height)

                    _dag_cid, _dag_block = await dag_utils.create_dag_block(
                        tx_hash=event_data['txHash'],
                        project_id=project_id,
                        tentative_block_height=target_tt_block_height,
                        payload_cid=event_data['event_data']['snapshotCid'],
                        timestamp=event_data['event_data']['timestamp'],
                    )

                    """ Reset the tenatativeBlockHeight"""
                    rest_logger.debug("Resetting the tentativeBlockHeight to: ")
                    rest_logger.debug(target_tt_block_height)

                    tentative_block_height_key = redis_keys.get_tentative_block_height_key(project_id)
                    _ = await writer_redis_conn.set(tentative_block_height_key, target_tt_block_height)

                    await dag_utils.clear_payload_commit_data(
                        project_id=project_id,
                        tx_hash=event_data['txHash'],
                        payload_commit_id=event_data['event_data']['payloadCommitId']
                    )

                    """ Move the payload_commit_id's from pendingTransactions into discardedTransactions """
                    pending_transactions = await reader_redis_conn.zrangebyscore(
                        key=pending_transactions_key,
                        max=tentative_block_height_cached,
                        min=max_block_height_project + 1,
                        withscores=True
                    )

                    for _tx_hash, _tx_tt_height in pending_transactions:
                        _ = await writer_redis_conn.zadd(
                            key=discarded_transactions_key,
                            score=_tx_tt_height,
                            member=_tx_hash
                        )

                        _ = await writer_redis_conn.zrem(
                            key=pending_transactions_key,
                            member=_tx_hash
                        )

                    """ Remove the keys for discarded events """
                    pending_blocks_key = redis_keys.get_pending_blocks_key(project_id)
                    all_pending_blocks = await reader_redis_conn.zrangebyscore(
                        key=pending_blocks_key,
                        max=tentative_block_height_event_data,
                        min=max_block_height_project + 1,
                    )

                    for pending_block in all_pending_blocks:
                        _payload_commit_id = pending_block.decode('utf-8')

                        # Delete the payload commit id from pending blocks
                        await dag_utils.clear_payload_commit_data(
                            project_id=project_id,
                            payload_commit_id=_payload_commit_id,
                            tx_hash="",
                        )
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
                else:
                    rest_logger.debug("Saving Event data for height: ")
                    rest_logger.debug(tentative_block_height_event_data)
                    _ = await dag_utils.save_event_data(
                        event_data=event_data,
                        writer_redis_conn=writer_redis_conn,
                    )

            elif tentative_block_height_event_data == max_block_height_project + 1:
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

                await dag_utils.clear_payload_commit_data(
                    project_id=project_id,
                    payload_commit_id=event_data['event_data']['payloadCommitId'],
                    tx_hash=event_data['txHash'],
                )

                """ retrieve all list of all the payload_commit_ids from the pendingBlocks zset """
                all_pending_blocks = await reader_redis_conn.zrange(
                    key=redis_keys.get_pending_blocks_key(project_id=project_id),
                    start=0,
                    stop=-1,
                    withscores=True
                )
                if len(all_pending_blocks) > 0:
                    """ There are some pending blocks left """
                    for _payload_commit_id, _tt_block_height in all_pending_blocks:
                        _payload_commit_id = _payload_commit_id.decode('utf-8')
                        _tt_block_height = int(_tt_block_height)

                        if _tt_block_height == max_block_height_project + 1:
                            rest_logger.debug("Processing tentative_block_height: ")
                            rest_logger.debug(tentative_block_height_event_data)

                            """ Retrieve the event data for the payload_commit_id """
                            event_data_key = f"eventData:{_payload_commit_id}"
                            out = await reader_redis_conn.hgetall(key=event_data_key)

                            _event_data = {k.decode('utf-8'): v.decode('utf-8') for k, v in out.items()}
                            _tx_hash = _event_data['txHash']
                            # _event_data['event_data'] = {}
                            # _event_data['event_data'].update({k: v for k, v in _event_data.items()})
                            # _tx_hash = _event_data['event_data'].pop('txHash')

                            """ Create the dag block for this event """
                            _dag_cid, dag_block = await dag_utils.create_dag_block(
                                tx_hash=_event_data['txHash'],
                                project_id=project_id,
                                tentative_block_height=_tt_block_height,
                                payload_cid=_event_data['event_data']['snapshotCid'],
                                timestamp=_event_data['event_data']['timestamp'],
                            )
                            max_block_height_project = dag_block['height']
                            await dag_utils.clear_payload_commit_data(
                                project_id=project_id,
                                payload_commit_id=_payload_commit_id,
                                tx_hash=_tx_hash,
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
                        else:
                            """ Since there is a pending block creation, stop the loop """
                            rest_logger.debug("There is pending block creation left. Breaking out of the loop..")
                            break
                else:
                    """ There are no pending blocks in the chain """
                    try:
                        diff_map = await diffmap_utils.calculate_diff(
                            dag_cid=_dag_cid,
                            dag=dag_block,
                            project_id=project_id,
                            ipfs_client=ipfs_client
                            # writer_redis_conn=writer_redis_conn
                        )
                    except json.decoder.JSONDecodeError as jerr:
                        rest_logger.debug("There was an error while decoding the JSON data")
                        rest_logger.debug(jerr)
                    else:
                        rest_logger.debug("The diff map retrieved")
                        rest_logger.debug(diff_map)
                        pass
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