from fastapi import FastAPI, Request, Response
import aioredis
import ipfshttpclient
from config import settings
import logging
from skydb import SkydbTable
import sys
import json
import os
import io

""" Powergate Imports """
from pygate_grpc.client import PowerGateClient

ipfs_client = ipfshttpclient.connect(settings.IPFS_URL)

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

REDIS_CONN_CONF = {
    "host": settings['REDIS']['HOST'],
    "port": settings['REDIS']['PORT'],
    "password": settings['REDIS']['PASSWORD'],
    "db": settings['REDIS']['DB']
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
    app.redis_pool = await aioredis.create_pool(
        address=(REDIS_CONN_CONF['host'], REDIS_CONN_CONF['port']),
        db=REDIS_CONN_CONF['db'],
        password=REDIS_CONN_CONF['password'],
        maxsize=5
    )

async def save_event_data(event_data:dict, redis_conn):
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
    _ = await redis_conn.hmset_dict(
        key=event_data_key,
        **fields
    )

    pending_blocks_key = f"projectId:{event_data['event_data']['projectId']}:pendingBlocks"
    _ = await redis_conn.zadd (
            key=pending_blocks_key, 
            member=event_data['event_data']['payloadCommitId'],
            score=int(event_data['event_data']['tentativeBlockHeight'])
        )

    return 0

async def create_dag_block(event_data:dict, redis_conn):

        txHash = event_data['txHash']
        project_id = event_data['event_data']['projectId']
        tentative_block_height = int(event_data['event_data']['tentativeBlockHeight'])
        snapshotCid = event_data['event_data']['snapshotCid']
        timestamp = event_data['event_data']['timestamp']

        if settings.block_storage == "FILECOIN":
            """ Get the filecoin token for the project Id """
            KEY = f"filecoinToken:{project_id}"
            token = await redis_conn.get(KEY)
            token = token.decode('utf-8')

        """ Get the last dag cid for the project_id """
        last_dag_cid_key = f"projectId:{project_id}:lastDagCid"
        last_dag_cid = await redis_conn.get(last_dag_cid_key)
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
                    'type': 'COLD_FILECOIN',
                }
        dag['txHash'] = txHash
        dag['timestamp'] = timestamp
        rest_logger.debug(dag)

        """ Convert dag structure to json and put it on ipfs dag """
        json_string = json.dumps(dag).encode('utf-8')
        dag_data = ipfs_client.dag.put(io.BytesIO(json_string))
        rest_logger.debug("The DAG cid is: ")
        rest_logger.debug(dag_data['Cid']['/'])

        if settings.block_storage == "FILECOIN":
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

            _ = await redis_conn.hmset_dict(
                key=block_filecoin_storage_key,
                **fields
            )
            
            rest_logger.debug("Pushed the block data to filecoin: ")
            rest_logger.debug("Job: "+job.jobId)

        if settings.METADATA_CACHE == 'skydb':
            ipfs_table = SkydbTable(
                table_name=f"{settings.dag_table_name}:{project_id}",
                columns=['cid'],
                seed=settings.seed,
                verbose=1
            )
            ipfs_table.add_row({'cid': dag_data['Cid']['/']})
        elif settings.METADATA_CACHE == 'redis':
            last_dag_cid_key = f"projectId:{project_id}:lastDagCid"
            await redis_conn.set(last_dag_cid_key, dag_data['Cid']['/'])
            await redis_conn.zadd(
                key=f'projectID:{project_id}:Cids',
                score=tentative_block_height,
                member=dag_data['Cid']['/']
            )
            await redis_conn.set(f'projectID:{project_id}:blockHeight', tentative_block_height)

        return dag_data['Cid']['/'], dag

async def calculate_diff(dag_cid:str, dag:dict, project_id:str, redis_conn):
    # cache last seen diffs
    global ipfs_client
    if dag['prevCid']:
        payload_cid = dag['data']['cid']
        prev_dag = ipfs_client.dag.get(dag['prevCid']).as_json()
        prev_payload_cid = prev_dag['data']['cid']
        if prev_payload_cid != payload_cid:
            diff_map = dict()
            prev_data = ipfs_client.cat(prev_payload_cid).decode('utf-8')
            rest_logger.debug(prev_data)
            rest_logger.debug(type(prev_data))
            prev_data = json.loads(prev_data)
            # rest_logger.info('Previous payload returned')
            # rest_logger.info(prev_data)
            payload = ipfs_client.cat(payload_cid).decode('utf-8')
            rest_logger.debug(payload)
            rest_logger.debug(type(payload))
            payload = json.loads(payload)
            # rest_logger.info('Current payload')
            # rest_logger.info(prev_data)
            for k, v in payload.items():
                if k not in prev_data.keys():
                    prev_data[k] = None
                if v != prev_data[k]:
                    diff_map[k] = {'old': prev_data[k], 'new': v}
            diff_data = {
                'cur': {
                    'height': dag['height'],
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
                await redis_conn.zadd(
                    diff_snapshots_cache_zset,
                    score=dag['height'],
                    member=json.dumps(diff_data)
                )
                latest_seen_snapshots_htable = 'auditprotocol:lastSeenSnapshots'
                await redis_conn.hset(
                    latest_seen_snapshots_htable,
                    project_id,
                    json.dumps(diff_data)
                )

            return diff_data

    return {}


@app.post('/')
async def create_dag(
        request: Request,
        response: Response,
    ):
    event_data = await request.json()
    #rest_logger.debug(event_data)
    if 'event_name' in event_data.keys():
        if event_data['event_name'] == 'RecordAppended':
            #rest_logger.debug(event_data)

            """ Create a Redis Connection """
            rest_logger.debug("Creating a redis connection....")
            redis_conn_raw = await request.app.redis_pool.acquire()
            redis_conn = aioredis.Redis(redis_conn_raw)

            """ Get data from the event """
            project_id = event_data['event_data']['projectId']
            tentative_block_height = event_data['event_data']['tentativeBlockHeight']

            # Get the max block height for the project_id
            max_block_height = await redis_conn.get(f"projectID:{project_id}:blockHeight")
            if max_block_height:
                max_block_height = int(max_block_height)
            else:
                max_block_height = 0

            rest_logger.debug("Tentative Block Height and Block Height")
            rest_logger.debug(tentative_block_height)
            rest_logger.debug(max_block_height)
            if tentative_block_height > max_block_height+1:
                """ 
                Since there are events that are yet to come, put this event data 
                into redis and once the required event arrives, complete the block creation
                """
                rest_logger.debug("There are pending block creations left. Caching the event data for: ")
                rest_logger.debug("Transaction: "+event_data['txHash']+", PayloadCommitId: "+event_data['event_data']['payloadCommitId'])
                _ = await save_event_data(event_data, redis_conn)

            elif tentative_block_height == max_block_height+1:
                """
                    An event which is in-order has arrived. Create a dag block for this event
                    and process all other pending events for this project
                """
                pending_blocks_key = f"projectId:{project_id}:pendingBlocks"
                pending_block_creations_key = f"projectId:{project_id}:pendingBlockCreation"

                _dag_cid, dag_block = await create_dag_block(event_data, redis_conn)
                max_block_height = dag_block['height']
                """ Remove this block from pending block creations SET """
                _ = await redis_conn.srem(pending_block_creations_key, event_data['event_data']['payloadCommitId'])
                _ = await redis_conn.zrem(pending_blocks_key, event_data['event_data']['payloadCommitId'])

                """ retrieve all list of all the payload_commit_ids from the pendingBlocks set """
                all_pending_blocks = await redis_conn.zrange(
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

                        if _tt_block_height == max_block_height+1:
                            rest_logger.debug("Processing:")
                            rest_logger.debug("payload_commit_id: "+_payload_commit_id)
                            rest_logger.debug("tentative block height: "+str(_tt_block_height))

                            """ Retrieve the event data for the payload_commit_id """ 
                            event_data_key = f"eventData:{_payload_commit_id}"
                            out = await redis_conn.hgetall(key=event_data_key)

                            _event_data = {k.decode('utf-8'):v.decode('utf-8') for k,v in out.items()}
                            fields_to_delete = list(_event_data.keys())
                            _event_data['event_data'] = {}
                            _event_data['event_data'].update({k:v for k,v in _event_data.items()})
                            _event_data['event_data'].pop('txHash')

                            """ Create the dag block for this event """
                            _dag_cid, dag_block = await create_dag_block(_event_data, redis_conn)
                            max_block_height = dag_block['height']

                            """ Remove the dag block from pending block creations list """
                            _ = await redis_conn.srem(pending_block_creations_key, _payload_commit_id)
                            _ = await redis_conn.zrem(pending_blocks_key, _payload_commit_id)

                            """ Delete the event_data """
                            _ = await redis_conn.hdel(event_data_key, *fields_to_delete)

                        else:
                            """ Since there is a pending block creation, stop the loop """
                            rest_logger.debug("There is pending block creation left. Breaking out of the loop..")
                            break
                else:
                    """ There are no pending blocks in the chaing """
                    try:
                        diff_map = await calculate_diff(_dag_cid, dag_block, project_id, redis_conn)
                    except json.decoder.JSONDecodeError as jerr:
                        rest_logger.debug("There was an error while decoding the JSON data")
                        rest_logger.debug(jerr)
                    else:
                        rest_logger.debug("The diff map retrieved")
                        rest_logger.debug(diff_map)

            request.app.redis_pool.release(redis_conn_raw)

    response.status_code = 200
    return {}
