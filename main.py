from typing import Optional, Union
from fastapi import Depends, FastAPI, WebSocket, HTTPException, Security, Request, Response, BackgroundTasks, Cookie, \
    Query, WebSocketDisconnect
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from eth_utils import keccak
from maticvigil.EVCore import EVCore
import logging
import sys
import json
import aioredis
import io
import redis
import time
from skydb import SkydbTable
import ipfshttpclient
from config import settings
from pygate_grpc.client import PowerGateClient
from uuid import uuid4

print(settings.as_dict())
ipfs_client = ipfshttpclient.connect()

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

# setup CORS origins stuff
origins = ["*"]

redis_lock = redis.Redis()

app = FastAPI(docs_url=None, openapi_url=None, redoc_url=None)
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"]
)
app.mount('/static', StaticFiles(directory='static'), name='static')

REDIS_CONN_CONF = {
    "host": settings['REDIS']['HOST'],
    "port": settings['REDIS']['PORT'],
    "password": settings['REDIS']['PASSWORD'],
    "db": settings['REDIS']['DB']
}
#
STORAGE_CONFIG = {
    "hot": {
        "enabled": True,
        "allowUnfreeze": True,
        "ipfs": {
            "addTimeout": 30
        }
    },
    "cold": {
        "enabled": True,
        "filecoin": {
            "repFactor": 1,
            "dealMinDuration": 518400,
            "renew": {
            },
            "addr": "placeholderstring"
        }
    }
}


@app.on_event('startup')
async def startup_boilerplate():
    app.redis_pool = await aioredis.create_pool(
        address=(REDIS_CONN_CONF['host'], REDIS_CONN_CONF['port']),
        db=REDIS_CONN_CONF['db'],
        password=REDIS_CONN_CONF['password'],
        maxsize=5
    )
    app.evc = EVCore(verbose=True)
    app.contract = app.evc.generate_contract_sdk(
        contract_address=settings.audit_contract,
        app_name='auditrecords'
    )

async def get_project_token(request: Request):
    """ From the request body, extract the projectId and return back the powergate token."""
    req_args = await request.json()
    projectId = req_args['projectId']

    """ Intitalize powergate client """
    rest_logger.debug("Intitializing powergate client")
    powgate_client = PowerGateClient(settings.POWERGATE_CLIENT_ADDR,False)

    if settings.METADATA_CACHE == "redis":
        """ Create a redis connection """
        redis_conn_raw = await request.app.redis_pool.acquire()
        redis_conn = aioredis.Redis(redis_conn_raw)

        """ Check if there is a filecoin token for the project Id """
        KEY = f"filecoinToken:{projectId}"
        token = await redis_conn.get(KEY)
        if not token:
            user = powgate_client.admin.users.create()
            token = user.token
            _ = await redis_conn.set(KEY, token)
            rest_logger.debug("Created a token for projectId: "+str(projectId))
            rest_logger.debug("Token: "+token)

            """ Save the project Id on set """
            _ = await redis_conn.sadd("storedProjectIds", projectId)
        else:
            token = token.decode('utf-8')
            rest_logger.debug("Retrieved token: "+token+", for project Id: "+str(projectId))

        """ Release Redis connection pool"""
        request.app.redis_pool.release(redis_conn_raw)

        return token


@app.get('/requests/{requestId:str}')
async def request_status(
    request: Request,
    response: Response,
    requestId:str,
):
    """
        Given a requestId, return either the status of that request or retrieve all the payloads for that 
    """

    redis_conn_raw = await request.app.redis_pool.acquire()
    redis_conn = aioredis.Redis(redis_conn_raw)

    # Check if the request is already in the pending list
    requests_list_key = f"pendingRetrievalRequests"
    out = await redis_conn.sismember(requests_list_key, requestId)
    if out == True:
        return {'requestId': requestId, 'requestStatus':'Pending'}

    # Get all the retrieved files
    retrieval_files_key = f"retrievalRequestFiles:{requestId}"
    retrieved_files = await redis_conn.zrange(
        key=retrieval_files_key,
        start=0,
        stop=-1,
        withscores=True
    )

    data = {}
    files = {}
    for file_, block_height in retrieved_files:
        file_ = file_.decode('utf-8')
        cid = file_.split('/')[-1]
        block_height = int(block_height)

        dag_block = {
            'payloadFile':'/'+file_,
            'Height': block_height
        }
        files[cid] = dag_block

    data['requestId'] = requestId
    data['requestStatus'] = 'Completed'
    data['files'] = files
    response.status_code = 200
    return data



@app.post('/commit_payload')
async def commit_payload(
        request: Request,
        response: Response,
        token: str = Depends(get_project_token)
):
    """"
        - This endpoint accepts the payload and adds it to ipfs.
        - The cid retrieved after adding to ipfs is committed to a Smart Contract and once the transaction goes
        through a webhook listener will catch that event and update the DAG with the latest block along with the timestamp
        of when the cid of the payload was committed.
    """
    req_args = await request.json()
    try:
        payload = req_args['payload']
        project_id = req_args['projectId']
        rest_logger.debug("Extracted payload and projectId from the request: ")
        rest_logger.debug(payload)
        rest_logger.debug(f"Payload data type: {type(payload)}")
        rest_logger.debug(f"ProjectId: {project_id}")
    except Exception as e:
        return {'error': "Either payload or projectId"}
    prev_dag_cid = ""
    prev_payload_cid = None
    block_height = 0
    ipfs_table = None
    redis_conn = None
    redis_conn_raw = None
    last_tentative_block_height = None
    last_tentative_block_height_key = f'projectID:{project_id}:tentativeBlockHeight'
    if settings.METADATA_CACHE == 'skydb':
        ipfs_table = SkydbTable(
            table_name=f"{settings.dag_table_name}:{project_id}",
            columns=['cid'],
            seed=settings.seed,
            verbose=1
        )
        if ipfs_table.index == 0:
            prev_dag_cid = ""
        else:
            prev_dag_cid = ipfs_table.fetch_row(row_index=ipfs_table.index - 1)['cid']
        block_height = ipfs_table.index

    elif settings.METADATA_CACHE == 'redis':
        rest_logger.debug("Fetching Data from Redis")
        """ Fetch the cid of latest DAG block along with the latest block height. """
        redis_conn_raw = await request.app.redis_pool.acquire()
        redis_conn = aioredis.Redis(redis_conn_raw)
        last_known_dag_cid_key = f'projectID:{project_id}:lastDagCid'
        r = await redis_conn.get(last_known_dag_cid_key)
        if r:
            """ Retrieve the height of the latest block only if the DAG projectId is not empty """
            prev_dag_cid = r.decode('utf-8')
            last_block_height_key = f'projectID:{project_id}:blockHeight'
            r2 = await redis_conn.get(last_block_height_key)
            if r2:
                block_height = int(r2)
        out = await redis_conn.get(last_tentative_block_height_key)
        rest_logger.debug(f"From Redis Last tentative block height: {out}")
        if out:
            last_tentative_block_height = int(out)
        

    rest_logger.debug(f"Got last tentative block height: {last_tentative_block_height}")

    """ 
        - If the prev_dag_cid is empty, then it means that this is the first block that
        will be added to the DAG of the projectId.
    """
    if prev_dag_cid != "":
        prev_payload_cid = ipfs_client.dag.get(prev_dag_cid).as_json()['Data']['Cid']
    payload_changed = False

    rest_logger.debug('Previous IPLD CID in the DAG: ')
    rest_logger.debug(prev_dag_cid)
    if not last_tentative_block_height:
        last_tentative_block_height = 0

    """ The DAG will be created in the Webhook listener script """
    # dag = settings.dag_structure.to_dict()

    """ Instead of adding payload to directly IPFS, stage it to Filecoin and get back the Cid"""
    if type(payload) is dict:
        payload = json.dumps(payload)
    payload = payload.encode('utf-8')
    powgate_client = PowerGateClient(settings.POWERGATE_CLIENT_ADDR, False)
    stage_res = powgate_client.data.stage_bytes(payload, token=token)
    """ Since the same data may come back for snapshotting, I have added override=True"""
    job = powgate_client.config.apply(stage_res.cid, override=True, token=token)
    snapshot_cid = stage_res.cid
    if snapshot_cid != prev_payload_cid:
        payload_changed = True
    payload_cid_key = f"projectID:{project_id}:payloadCids"
    _ = await redis_conn.zadd(
        key=payload_cid_key,
        score=last_tentative_block_height,
        member=snapshot_cid
    )

    """ Add the job id to redis. """
    KEY = f"JobStatus:{snapshot_cid}"
    _ = await redis_conn.set(key=KEY, value=job.jobId)
    rest_logger.debug("Pushed the payload to filecoin.")
    rest_logger.debug("Job Id: "+job.jobId)


    #""" Add the Payload to IPFS """
    #if type(payload) is dict:
    #    snapshot_cid = ipfs_client.add_json(payload)
    #else:
    #    try:
    #        snapshot_cid = ipfs_client.add_str(str(payload))
    #    except:
    #        response.status_code = 400
    #        return {'success': False, 'error': 'PayloadNotSuppported'}
    rest_logger.debug('Payload CID')
    rest_logger.debug(snapshot_cid)
    snapshot = dict()
    snapshot['Cid'] = snapshot_cid
    snapshot['Type'] = "COLD_FILECOIN"
    """ Check if the payload has changed. """
    if prev_payload_cid:
        if prev_payload_cid != snapshot['Cid']:
            payload_changed = True
    rest_logger.debug(snapshot)

    ipfs_cid = snapshot['Cid']
    token_hash = '0x' + keccak(text=json.dumps(snapshot)).hex()
    tx_hash_obj = request.app.contract.commitRecord(**dict(
        ipfsCid=ipfs_cid,
        apiKeyHash=token_hash,
    ))

    if settings.METADATA_CACHE == 'redis':
        """ Put this transaction hash on redis so that webhook listener can access it when listening to events"""
        hash_key = f"TRANSACTION:{tx_hash_obj[0]['txHash']}"
        hash_field = f"project_id"
        r = await redis_conn.hset(
                key=hash_key,
                field=hash_field,
                value=f"{project_id}"
            )
        hash_field = f"tentative_block_height"
        r = await redis_conn.hset(
                key=hash_key,
                field=hash_field,
                value=f"{last_tentative_block_height}"
            )

        hash_field = f"prev_dag_cid"
        r = await redis_conn.hset(
                key=hash_key,
                field=hash_field,
                value=prev_dag_cid,
            )
        _ = await redis_conn.set(last_tentative_block_height_key, last_tentative_block_height+1)
        request.app.redis_pool.release(redis_conn_raw)
    
    return {
        'Cid': snapshot['Cid'],
        'tentativeHeight': last_tentative_block_height,
        'payloadChanged': payload_changed,
    }



@app.get('/{projectId:str}/payloads')
async def get_payloads(
        request: Request,
        response: Response,
        projectId: str,
        from_height: int = Query(None),
        to_height: int = Query(None),
        data: Optional[str] = Query(None)
):
    ipfs_table = None
    max_block_height = None
    redis_conn_raw = None
    redis_conn = None
    user_token = None
    powgate_client = PowerGateClient(settings.POWERGATE_CLIENT_ADDR,False)
    if settings.METADATA_CACHE == 'skydb':
        ipfs_table = SkydbTable(table_name=f"{settings.dag_table_name}:{projectId}",
                                columns=['cid'],
                                seed=settings.seed,
                                verbose=1)
        max_block_height = ipfs_table.index - 1
    elif settings.METADATA_CACHE == 'redis':
        redis_conn_raw = await request.app.redis_pool.acquire()
        redis_conn = aioredis.Redis(redis_conn_raw)
        h = await redis_conn.get(f'projectID:{projectId}:blockHeight')
        user_token = await redis_conn.get(f'filecoinToken:{projectId}')
        if not user_token:
            response.status_code = 403
            return {'error':'Forbidden'}
        user_token = user_token.decode('utf-8')
        if not h:
            max_block_height = 0
        else:
            max_block_height = int(h.decode('utf-8')) - 1
    if data:
        if data.lower() == 'true':
            data = True
        else:
            data = False

    if (from_height < 0) or (to_height > max_block_height) or (from_height > to_height):
        return {'error': 'Invalid Height'}

    if from_height < max_block_height - settings.max_ipfs_blocks:
        """ Create a request Id and start a retrieval request """
        request_id = str(uuid4())

        """ Setup the retrievalRequestInfo Hashset """
        key = f"retrievalRequestInfo:{request_id}"
        _ = await redis_conn.hset(
            key=key,
            field='projectId',
            value=projectId
        )
        _ = await redis_conn.hset(
            key=key,
            field='to_height',
            value=to_height
        )
        _ = await redis_conn.hset(
            key=key,
            field='from_height',
            value=from_height
        )
        _data = '1' if data else '0'
        _ = await redis_conn.hset(
            key=key,
            field='data',
            value=_data
        )
        requests_list_key = f"pendingRetrievalRequests"
        _ = await redis_conn.sadd(requests_list_key, request_id)
        request.app.redis_pool.release(redis_conn_raw)

        return {'requestId': request_id}

    blocks = list()
    current_height = to_height
    prev_dag_cid = ""
    prev_payload_cid = None
    idx = 0
    while current_height >= from_height:
        rest_logger.debug("Fetching block at height: " + str(current_height))
        if not prev_dag_cid:
            if settings.METADATA_CACHE == 'skydb':
                prev_dag_cid = ipfs_table.fetch_row(row_index=current_height)['cid']
            elif settings.METADATA_CACHE == 'redis':
                project_cids_key_zset = f'projectID:{projectId}:Cids'
                r = await redis_conn.zrangebyscore(
                    key=project_cids_key_zset,
                    min=current_height,
                    max=current_height,
                    withscores=False
                )
                if r:
                    prev_dag_cid = r[0].decode('utf-8')
                else:
                    return {'error': 'NoRecordsFound'}
        block = None
        block = ipfs_client.dag.get(prev_dag_cid).as_json()
        rest_logger.debug("Block Retrieved: ")
        rest_logger.debug(block)
        formatted_block = dict()
        formatted_block['dagCid'] = prev_dag_cid
        formatted_block.update({k: v for k, v in block.items()})
        formatted_block['prevDagCid'] = formatted_block.pop('prevCid')
        if data:
            if current_height < max_block_height - settings.max_ipfs_blocks :
                payload_data = powgate_client.data.get(block['Data']['Cid'],token=user_token).decode('utf-8')
                formatted_block['Data']['payload'] = payload_data
            else:
                formatted_block['Data']['payload'] = ipfs_client.cat(block['Data']['Cid']).decode('utf-8')
        if prev_payload_cid:
            if prev_payload_cid != block['Data']['Cid']:
                blocks[idx-1]['payloadChanged'] = True
                diff_key = f"CidDiff:{prev_payload_cid}:{block['Data']['Cid']}"
                diff_b = await redis_conn.get(diff_key)
                diff_map = dict()
                if not diff_b:
                    # diff not cached already
                    rest_logger.debug('Diff not cached | New CID | Old CID')
                    rest_logger.debug(blocks[idx-1]['Data']['Cid'])
                    rest_logger.debug(block['Data']['Cid'])
                    if 'payload' in formatted_block['Data'].keys():
                        prev_data = formatted_block['Data']['payload']
                    else:
                        if current_height < max_block_height - settings.max_ipfs_blocks:
                            prev_data = powgate_client.data.get(block['Data']['Cid'], token=user_token).decode(
                                'utf-8')
                        else:
                            prev_data = ipfs_client.cat(block['Data']['Cid']).decode('utf-8')
                    rest_logger.debug("Got the payload data: ")
                    rest_logger.debug(prev_data)
                    prev_data = json.loads(prev_data)
                    if 'payload' in blocks[idx-1]['Data'].keys():
                        cur_data = blocks[idx-1]['Data']['payload']
                    else:
                        cur_data = ipfs_client.cat(blocks[idx-1]['Data']['Cid']).decode('utf-8')
                    cur_data = json.loads(cur_data)
                    # calculate diff
                    for k, v in cur_data.items():
                        if k not in prev_data.keys():
                            rest_logger.info('Ignoring key in older payload as it is not present')
                            rest_logger.info(k)
                            blocks[idx - 1]['payloadChanged'] = False
                            break
                        if v != prev_data[k]:
                            diff_map[k] = {'old': prev_data[k], 'new': v}
                    if len(diff_map):
                        rest_logger.debug('Found diff in first time calculation')
                        rest_logger.debug(diff_map)
                    # cache in redis
                    await redis_conn.set(diff_key, json.dumps(diff_map))
                else:
                    diff_map = json.loads(diff_b)
                    rest_logger.debug('Found Diff in Cache! | New CID | Old CID | Diff')
                    rest_logger.debug(blocks[idx - 1]['Data']['Cid'])
                    rest_logger.debug(block['Data']['Cid'])
                    rest_logger.debug(diff_map)
                blocks[idx-1]['diff'] = diff_map
            else:
                blocks[idx-1]['payloadChanged'] = False
        prev_payload_cid = block['Data']['Cid']
        blocks.append(formatted_block)
        prev_dag_cid = formatted_block['prevDagCid']
        current_height = current_height - 1
        idx += 1
    if settings.METADATA_CACHE == 'redis':
        request.app.redis_pool.release(redis_conn_raw)
    return blocks


@app.get('/{projectId}/payloads/height')
async def payload_height(request: Request, response: Response, projectId: str):
    max_block_height = -1
    if settings.METADATA_CACHE == 'skydb':
        ipfs_table = SkydbTable(table_name=f"{settings.dag_table_name}:{projectId}",
                                columns=['cid'],
                                seed=settings.seed)
        max_block_height = ipfs_table.index - 1
    elif settings.METADATA_CACHE == 'redis':
        redis_conn_raw = await request.app.redis_pool.acquire()
        redis_conn = aioredis.Redis(redis_conn_raw)
        h = await redis_conn.get(f'projectID:{projectId}:blockHeight')
        if not h:
            max_block_height = 0
        else:
            max_block_height = int(h.decode('utf-8')) - 1
        request.app.redis_pool.release(redis_conn_raw)

    return {"height": max_block_height}


@app.get('/{projectId}/payload/{block_height}')
async def get_block(request: Request,
                    response: Response,
                    projectId: str,
                    block_height: int,
                    ):
    if settings.METADATA_CACHE == 'skydb':
        ipfs_table = SkydbTable(table_name=f"{settings.dag_table_name}:{projectId}",
                                columns=['cid'],
                                seed=settings.seed,
                                verbose=1)

        if (block_height > ipfs_table.index - 1) or (block_height < 0):
            response.status_code = 400
            return {'error': 'Invalid block Height'}

        else:
            row = ipfs_table.fetch_row(row_index=block_height)
            block = ipfs_client.dag.get(row['cid']).as_json()
            return {row['cid']: block}
    elif settings.METADATA_CACHE == 'redis':
        redis_conn_raw = await request.app.redis_pool.acquire()
        redis_conn = aioredis.Redis(redis_conn_raw)
        max_block_height = await redis_conn.get(f"projectID:{projectId}:blockHeight")
        if not max_block_height:
            response.status_code = 400
            return {'error': 'Block does not exist at this block height'}
        max_block_height = int(max_block_height.decode('utf-8'))-1
        rest_logger.debug(max_block_height)
        if block_height > max_block_height:
            response.status_code = 400
            return {'error': 'Invalid Block Height'}

        if block_height < max_block_height - settings.max_ipfs_blocks:
            rest_logger.debug(f"Block at {block_height} is being fetched")
            request_id = str(uuid4())

            from_height = block_height
            to_height = block_height
            data = False
            """ Setup the retrievalRequestInfo Hashset """
            key = f"retrievalRequestInfo:{request_id}"
            _ = await redis_conn.hset(
                key=key,
                field='projectId',
                value=projectId
            )
            _ = await redis_conn.hset(
                key=key,
                field='to_height',
                value=to_height
            )
            _ = await redis_conn.hset(
                key=key,
                field='from_height',
                value=from_height
            )
            _data = '1' if data else '0'
            _ = await redis_conn.hset(
                key=key,
                field='data',
                value=_data
            )
            requests_list_key = f"pendingRetrievalRequests"
            _ = await redis_conn.sadd(requests_list_key, request_id)

            return {'requestId': request_id}



        project_cids_key_zset = f'projectID:{projectId}:Cids'
        r = await redis_conn.zrangebyscore(
            key=project_cids_key_zset,
            min=block_height,
            max=block_height,
            withscores=False
        )
        prev_dag_cid = r[0].decode('utf-8')
        if block_height < max_block_height - settings.max_ipfs_blocks:
            rest_logger.debug("Fetching data from Filecoin....")
            """ Intitalized the powergate Client """
            user_token = None
            powgate_client = PowerGateClient(settings.POWERGATE_CLIENT_ADDR, False)
            user_token = await redis_conn.get(f'filecoinToken:{projectId}')
            user_token = user_token.decode('utf-8')
            """ Get the blockStagedCid for that block"""
            KEY = f"blockFilecoinStorage:{projectId}:{block_height}"
            staged_cid = await redis_conn.hget(
                key=KEY,
                field="blockStageCid"
            )
            staged_cid = staged_cid.decode('utf-8')
            data = powgate_client.data.get(staged_cid, token=user_token)
            data = data.decode('utf-8')
            block = json.loads(data)
        else:
            block = ipfs_client.dag.get(prev_dag_cid).as_json()
        request.app.redis_pool.release(redis_conn_raw)
        return {prev_dag_cid: block}


@app.get('/{projectId:str}/payload/{block_height:int}/data')
async def get_block_data(
        request: Request,
        response: Response,
        projectId: str,
        block_height: int,
):
    if settings.METADATA_CACHE == 'skydb':
        ipfs_table = SkydbTable(table_name=f"{settings.dag_table_name}:{projectId}",
                                columns=['cid'],
                                seed=settings.seed,
                                verbose=1)
        if (block_height > ipfs_table.index - 1) or (block_height < 0):
            return {'error': 'Invalid block Height'}
        row = ipfs_table.fetch_row(row_index=block_height)
        block = ipfs_client.dag.get(row['cid']).as_json()
        block['Data']['payload'] = ipfs_client.cat(block['Data']['Cid']).decode()
        return {row['cid']: block['Data']}

    elif settings.METADATA_CACHE == "redis":
        redis_conn_raw = await request.app.redis_pool.acquire()
        redis_conn = aioredis.Redis(redis_conn_raw)
        max_block_height = await redis_conn.get(f"projectID:{projectId}:blockHeight")
        if not max_block_height:
            response.status_code = 400
            return {'error': 'Invalid Block Height'}
        max_block_height = int(max_block_height.decode('utf-8'))-1
        if block_height > max_block_height:
            response.status_code = 400
            return {'error': 'Invalid Block Height'}

        if block_height < max_block_height - settings.max_ipfs_blocks:
            rest_logger.debug(f"Block at {block_height} is being fetched")
            request_id = str(uuid4())

            from_height = block_height
            to_height = block_height
            data = True
            """ Setup the retrievalRequestInfo Hashset """
            key = f"retrievalRequestInfo:{request_id}"
            _ = await redis_conn.hset(
                key=key,
                field='projectId',
                value=projectId
            )
            _ = await redis_conn.hset(
                key=key,
                field='to_height',
                value=to_height
            )
            _ = await redis_conn.hset(
                key=key,
                field='from_height',
                value=from_height
            )
            _data = '2' # Get data only and not the block itself
            _ = await redis_conn.hset(
                key=key,
                field='data',
                value=_data
            )
            requests_list_key = f"pendingRetrievalRequests"
            _ = await redis_conn.sadd(requests_list_key, request_id)

            return {'requestId': request_id}

        project_cids_key_zset = f'projectID:{projectId}:Cids'
        r = await redis_conn.zrangebyscore(
            key=project_cids_key_zset,
            min=block_height,
            max=block_height,
            withscores=False
        )
        prev_dag_cid = r[0].decode('utf-8')
        if block_height < max_block_height - settings.max_ipfs_blocks:
            """ Intitalized the powergate Client """
            user_token = None
            powgate_client = PowerGateClient(settings.POWERGATE_CLIENT_ADDR, False)
            user_token = await redis_conn.get(f'filecoinToken:{projectId}')
            user_token = user_token.decode('utf-8')
            """ Get the blockStagedCid for that block"""
            KEY = f"blockFilecoinStorage:{projectId}:{block_height}"
            staged_cid = await redis_conn.hget(
                key=KEY,
                field="blockStageCid"
            )
            staged_cid = staged_cid.decode('utf-8')
            data = powgate_client.data.get(staged_cid, token=user_token)
            data = data.decode('utf-8')
            block = json.loads(data)
        else:
            block = ipfs_client.dag.get(prev_dag_cid).as_json()
        payload = block['Data']
        if block_height < max_block_height - settings.max_ipfs_blocks:
            payload_data = powgate_client.data.get(block['Data']['Cid'], token=user_token).decode(
                'utf-8')
        else:
            payload_data = ipfs_client.cat(block['Data']['Cid']).decode('utf-8')
        payload['payload'] = payload_data
        request.app.redis_pool.release(redis_conn_raw)
        return {prev_dag_cid: payload}

