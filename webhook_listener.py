from fastapi import FastAPI, Request, Response, Header
import aioredis
from config import settings
import logging
from skydb import SkydbTable
from redis_conn import inject_reader_redis_conn, inject_writer_redis_conn
import sys
import json
import os
import io
from ipfs_async import client as ipfs_client
from utils import process_payloads_for_diff
from utils import preprocess_dag
import hmac

""" Powergate Imports """
from pygate_grpc.client import PowerGateClient

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
    "host": settings['REDIS']['HOST'],
    "port": settings['REDIS']['PORT'],
    "password": settings['REDIS']['PASSWORD'],
    "db": settings['REDIS']['DB']
}

REDIS_READER_CONN_CONF = {
    "host": settings['REDIS_READER']['HOST'],
    "port": settings['REDIS_READER']['PORT'],
    "password": settings['REDIS_READER']['PASSWORD'],
    "db": settings['REDIS_READER']['DB']
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
    app.writer_redis_pool = await aioredis.create_pool(
        address=(REDIS_WRITER_CONN_CONF['host'], REDIS_WRITER_CONN_CONF['port']),
        db=REDIS_WRITER_CONN_CONF['db'],
        password=REDIS_WRITER_CONN_CONF['password'],
        maxsize=50
    )
    app.reader_redis_pool = await aioredis.create_pool(
        address=(REDIS_READER_CONN_CONF['host'], REDIS_READER_CONN_CONF['port']),
        db=REDIS_READER_CONN_CONF['db'],
        password=REDIS_READER_CONN_CONF['password'],
        maxsize=50
    )


async def save_event_data(event_data: dict, writer_redis_conn):
    """
        - Given event_data, save the txHash, timestamp, projectId, snapshotCid, tentativeBlockHeight
        onto a redis HashTable with key: eventData:{payloadCommitId}
        - And then add the payload_commit_id to a zset with key: projectId:{projectId}:pendingBlocks
        with score being the tentativeBlockHeight
    """

    event_data_key = f"eventData:{event_data['event_data']['payloadCommitId']}"
    fields = {
        'txHash': event_data['txHash'],
        'projectId': event_data['event_data']['projectId'],
        'timestamp': event_data['event_data']['timestamp'],
        'snapshotCid': event_data['event_data']['snapshotCid'],
        'tentativeBlockHeight': event_data['event_data']['tentativeBlockHeight']
    }
    _ = await writer_redis_conn.hmset_dict(
        key=event_data_key,
        **fields
    )

    pending_blocks_key = f"projectID:{event_data['event_data']['projectId']}:pendingBlocks"
    _ = await writer_redis_conn.zadd(
        key=pending_blocks_key,
        member=event_data['event_data']['payloadCommitId'],
        score=int(event_data['event_data']['tentativeBlockHeight'])
    )

    return 0


async def create_dag_block(event_data: dict, reader_redis_conn, writer_redis_conn):
    txHash = event_data['txHash']
    project_id = event_data['event_data']['projectId']
    tentative_block_height = int(event_data['event_data']['tentativeBlockHeight'])
    snapshotCid = event_data['event_data']['snapshotCid']
    timestamp = event_data['event_data']['timestamp']

    """ Get the last dag cid for the project_id """
    last_dag_cid_key = f"projectID:{project_id}:lastDagCid"
    last_dag_cid = await reader_redis_conn.get(last_dag_cid_key)
    rest_logger.debug("Got last Dag Cid: ")
    rest_logger.debug(last_dag_cid)
    if last_dag_cid:
        last_dag_cid = last_dag_cid.decode('utf-8')
    else:
        last_dag_cid = ""

    """ Fill up the dag """
    dag = settings.dag_structure
    dag['height'] = tentative_block_height
    dag['prevCid'] = last_dag_cid
    dag['data'] = {
        'cid': snapshotCid,
        'type': 'HOT_IPFS',
    }
    dag['txHash'] = txHash
    dag['timestamp'] = timestamp
    rest_logger.debug(dag)

    """ Convert dag structure to json and put it on ipfs dag """
    json_string = json.dumps(dag).encode('utf-8')
    dag_data = await ipfs_client.dag.put(io.BytesIO(json_string))
    rest_logger.debug("The DAG cid is: ")
    rest_logger.debug(dag_data['Cid']['/'])

    if settings.block_storage == "FILECOIN":
        """ Get the filecoin token for the project Id """
        KEY = f"filecoinToken:{project_id}"
        token = await reader_redis_conn.get(KEY)
        token = token.decode('utf-8')

        """ Put the dag json on to filecoin """
        powgate_client = PowerGateClient(settings.POWERGATE_CLIENT_ADDR, False)
        staged_res = powgate_client.data.stage_bytes(json_string, token=token)
        job = powgate_client.config.apply(staged_res.cid, override=False, token=token)

        block_filecoin_storage_key = f"blockFilecoinStorage:{project_id}:{tentative_block_height}"
        fields = {
            'blockStageCid': staged_res.cid,
            'blockDagCid': dag_data['Cid']['/'],
            'jobId': job.jobId
        }

        _ = await writer_redis_conn.hmset_dict(
            key=block_filecoin_storage_key,
            **fields
        )

        rest_logger.debug("Pushed the block data to filecoin: ")
        rest_logger.debug("Job: ")
        rest_logger.debug(job.jobId)

    if settings.METADATA_CACHE == 'skydb':
        ipfs_table = SkydbTable(
            table_name=f"{settings.dag_table_name}:{project_id}",
            columns=['cid'],
            seed=settings.seed,
            verbose=1
        )
        ipfs_table.add_row({'cid': dag_data['Cid']['/']})
    elif settings.METADATA_CACHE == 'redis':
        last_dag_cid_key = f"projectID:{project_id}:lastDagCid"
        await writer_redis_conn.set(last_dag_cid_key, dag_data['Cid']['/'])
        await writer_redis_conn.zadd(
            key=f'projectID:{project_id}:Cids',
            score=tentative_block_height,
            member=dag_data['Cid']['/']
        )
        await writer_redis_conn.set(f'projectID:{project_id}:blockHeight', tentative_block_height)

    return dag_data['Cid']['/'], dag


def check_signature(core_payload, signature):
    """
    Given the core_payload, check the signature generated by the webhook listener
    """

    _sign_rebuilt = hmac.new(
        key=settings.API_KEY.encode('utf-8'),
        msg=json.dumps(core_payload).encode('utf-8'),
        digestmod='sha256'
    ).hexdigest()

    rest_logger.debug("Signature the came in the header: ")
    rest_logger.debug(signature)
    rest_logger.debug("Signature rebuilt from core payload: ")
    rest_logger.debug(_sign_rebuilt)

    return _sign_rebuilt == signature


async def calculate_diff(dag_cid: str, dag: dict, project_id: str, reader_redis_conn, writer_redis_conn):
    # cache last seen diffs
    dag_height = dag['height']
    rest_logger.debug("DAG at point A:")
    rest_logger.debug(dag)
    if dag['prevCid']:
        payload_cid = dag['data']['cid']
        _prev_dag = await ipfs_client.dag.get(dag['prevCid'])
        prev_dag = _prev_dag.as_json()
        prev_dag = preprocess_dag(prev_dag)
        prev_payload_cid = prev_dag['data']['cid']
        if prev_payload_cid != payload_cid:
            diff_map = dict()
            _prev_data = await ipfs_client.cat(prev_payload_cid)
            prev_data = _prev_data.decode('utf-8')
            rest_logger.debug(prev_data)
            rest_logger.debug(type(prev_data))
            prev_data = json.loads(prev_data)
            _payload = await ipfs_client.cat(payload_cid)
            payload = _payload.decode('utf-8')
            rest_logger.debug(payload)
            rest_logger.debug(type(payload))
            payload = json.loads(payload)
            rest_logger.debug('Before payload clean up')
            rest_logger.debug({'cur_payload': payload, 'prev_payload': prev_data})
            result = await process_payloads_for_diff(
                project_id,
                prev_data,
                payload,
                reader_redis_conn
            )
            rest_logger.debug('After payload clean up and comparison if any')
            rest_logger.debug(result)
            cur_data_copy = result['cur_copy']
            prev_data_copy = result['prev_copy']

            for k, v in cur_data_copy.items():
                if k not in result['payload_changed']:
                    if k not in prev_data_copy.keys():
                        prev_data_copy[k] = None
                    if v != prev_data_copy.get(k):
                        diff_map[k] = {'old': prev_data.get(k), 'new': payload.get(k)}
                else:
                    if result['payload_changed'][k]:
                        diff_map[k] = {'old': prev_data.get(k), 'new': payload.get(k)}

            if diff_map:
                rest_logger.debug("DAG at point B:")
                rest_logger.debug(dag)
                diff_data = {
                    'cur': {
                        'height': dag_height,
                        'payloadCid': payload_cid,
                        'dagCid': dag_cid,
                        'txHash': dag['txHash'],
                        'timestamp': dag['timestamp']
                    },
                    'prev': {
                        'height': prev_dag['height'],
                        'payloadCid': prev_payload_cid,
                        'dagCid': dag['prevCid']
                        # this will be used to fetch the previous block timestamp from the DAG
                    },
                    'diff': diff_map
                }
                if settings.METADATA_CACHE == 'redis':
                    diff_snapshots_cache_zset = f'projectID:{project_id}:diffSnapshots'
                    await writer_redis_conn.zadd(
                        diff_snapshots_cache_zset,
                        score=int(dag['height']),
                        member=json.dumps(diff_data)
                    )
                    latest_seen_snapshots_htable = 'auditprotocol:lastSeenSnapshots'
                    await writer_redis_conn.hset(
                        latest_seen_snapshots_htable,
                        project_id,
                        json.dumps(diff_data)
                    )

                return diff_data

    return {}


@app.post('/')
@inject_writer_redis_conn
@inject_reader_redis_conn
async def create_dag(
        request: Request,
        response: Response,
        x_hook_signature: str = Header(None),
        reader_redis_conn=None,
        writer_redis_conn=None,
):
    event_data = await request.json()
    # Verify the payload that has arrived.
    if x_hook_signature:
        is_safe = check_signature(event_data, x_hook_signature)
        if is_safe:
            rest_logger.debug("The arriving payload has been verified")
        else:
            rest_logger.debug("Recieved an wrong signature payload")
            return dict()
    if 'event_name' in event_data.keys():
        if event_data['event_name'] == 'RecordAppended':
            rest_logger.debug(event_data)

            """ Create a Redis Connection """
            rest_logger.debug("Creating a redis connection....")

            """ Get data from the event """
            project_id = event_data['event_data']['projectId']
            tentative_block_height = event_data['event_data']['tentativeBlockHeight']

            # Get the max block height for the project_id
            max_block_height = await reader_redis_conn.get(f"projectID:{project_id}:blockHeight")
            if max_block_height:
                max_block_height = int(max_block_height)
            else:
                max_block_height = 0

            rest_logger.debug("Tentative Block Height and Block Height")
            rest_logger.debug(tentative_block_height)
            rest_logger.debug(max_block_height)
            if tentative_block_height > max_block_height + 1:
                """ 
                Since there are events that are yet to come, put this event data 
                into redis and once the required event arrives, complete the block creation
                """
                rest_logger.debug("There are pending block creations left. Caching the event data for: ")
                _ = await save_event_data(event_data, writer_redis_conn=writer_redis_conn)

            elif tentative_block_height == max_block_height + 1:
                """
                    An event which is in-order has arrived. Create a dag block for this event
                    and process all other pending events for this project
                """
                pending_blocks_key = f"projectID:{project_id}:pendingBlocks"
                pending_block_creations_key = f"projectID:{project_id}:pendingBlockCreation"

                _dag_cid, dag_block = await create_dag_block(event_data, writer_redis_conn=writer_redis_conn,
                                                             reader_redis_conn=reader_redis_conn)
                max_block_height = dag_block['height']
                """ Remove this block from pending block creations SET """
                _ = await writer_redis_conn.srem(pending_block_creations_key,
                                                 event_data['event_data']['payloadCommitId'])
                _ = await writer_redis_conn.zrem(pending_blocks_key, event_data['event_data']['payloadCommitId'])

                """ Delete this payload commit hash field"""
                payload_commit_key = f"payloadCommit:{event_data['event_data']['payloadCommitId']}"
                _ = await writer_redis_conn.delete(payload_commit_key)

                """ retrieve all list of all the payload_commit_ids from the pendingBlocks set """
                all_pending_blocks = await reader_redis_conn.zrange(
                    key=pending_blocks_key,
                    start=0,
                    stop=-1,
                    withscores=True
                )
                if len(all_pending_blocks) > 0:
                    """ There are some pending blocks left """
                    for _payload_commit_id, _tt_block_height in all_pending_blocks:
                        _payload_commit_id = _payload_commit_id.decode('utf-8')
                        _tt_block_height = int(_tt_block_height)

                        if _tt_block_height == max_block_height + 1:
                            rest_logger.debug("Processing:")
                            rest_logger.debug("payload_commit_id: ")
                            rest_logger.debug(_payload_commit_id)
                            rest_logger.debug("tentative_block_height: ")
                            rest_logger.debug(tentative_block_height)

                            """ Retrieve the event data for the payload_commit_id """
                            event_data_key = f"eventData:{_payload_commit_id}"
                            out = await reader_redis_conn.hgetall(key=event_data_key)

                            _event_data = {k.decode('utf-8'): v.decode('utf-8') for k, v in out.items()}
                            _event_data['event_data'] = {}
                            _event_data['event_data'].update({k: v for k, v in _event_data.items()})
                            _event_data['event_data'].pop('txHash')

                            """ Create the dag block for this event """
                            _dag_cid, dag_block = await create_dag_block(
                                _event_data,
                                writer_redis_conn=writer_redis_conn,
                                reader_redis_conn=reader_redis_conn
                            )
                            max_block_height = dag_block['height']

                            """ Remove the dag block from pending block creations list """
                            _ = await writer_redis_conn.srem(pending_block_creations_key, _payload_commit_id)
                            _ = await writer_redis_conn.zrem(pending_blocks_key, _payload_commit_id)

                            """ Delete the event_data """
                            _ = await writer_redis_conn.delete(event_data_key)

                            """ Delete the payload commit data """
                            payload_commit_key = f"payloadCommit:{_payload_commit_id}"
                            _ = await writer_redis_conn.delete(payload_commit_key)


                        else:
                            """ Since there is a pending block creation, stop the loop """
                            rest_logger.debug("There is pending block creation left. Breaking out of the loop..")
                            break
                else:
                    """ There are no pending blocks in the chain """
                    try:
                        diff_map = await calculate_diff(
                            dag_cid=_dag_cid,
                            dag=dag_block,
                            project_id=project_id,
                            reader_redis_conn=reader_redis_conn,
                            writer_redis_conn=writer_redis_conn
                        )
                    except json.decoder.JSONDecodeError as jerr:
                        rest_logger.debug("There was an error while decoding the JSON data")
                        rest_logger.debug(jerr)
                    else:
                        rest_logger.debug("The diff map retrieved")
                        rest_logger.debug(diff_map)

    response.status_code = 200
    return dict()
