from typing import Optional, Union
from fastapi import Depends, FastAPI, WebSocket, HTTPException, Security, Request, Response, BackgroundTasks, Cookie, \
    Query, WebSocketDisconnect
from fastapi.middleware.cors import CORSMiddleware
from functools import wraps
from fastapi.staticfiles import StaticFiles
from eth_utils import keccak
from maticvigil.EVCore import EVCore
from maticvigil.exceptions import EVBaseException
import logging
import sys
import json
import aioredis
import io
import redis
import time
import async_timeout
from skydb import SkydbTable
import ipfshttpclient
from config import settings
from pygate_grpc.client import PowerGateClient
from uuid import uuid4
import requests
import async_timeout
from redis_conn import setup_teardown_boilerplate

ipfs_client = ipfshttpclient.connect(settings.IPFS_URL)

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
        maxsize=100
    )
    app.evc = EVCore(verbose=True)
    app.contract = app.evc.generate_contract_sdk(
        contract_address=settings.audit_contract,
        app_name='auditrecords'
    )


@setup_teardown_boilerplate
async def get_project_token(
        request: Request = None,
        projectId: str = None,
        override_settings=False,
        redis_conn=None
):
    """ From the request body, extract the projectId and return back the powergate token."""
    if projectId is None:
        req_args = await request.json()
        projectId = req_args['projectId']

    """ Save the project Id on set """
    _ = await redis_conn.sadd("storedProjectIds", projectId)

    """ Intitalize powergate client """
    rest_logger.debug("Intitializing powergate client")
    if (settings.payload_storage != "FILECOIN") and (settings.block_storage != "FILECOIN") and (
            override_settings == False):
        return ""

    powgate_client = PowerGateClient(settings.POWERGATE_CLIENT_ADDR, False)

    if settings.METADATA_CACHE == "redis":
        """ Check if there is a filecoin token for the project Id """
        KEY = f"filecoinToken:{projectId}"
        token = await redis_conn.get(KEY)
        if not token:
            user = powgate_client.admin.users.create()
            token = user.token
            _ = await redis_conn.set(KEY, token)
            rest_logger.debug("Created a token for projectId: " + str(projectId))
            rest_logger.debug("Token: " + token)

        else:
            token = token.decode('utf-8')
            rest_logger.debug("Retrieved token: " + token + ", for project Id: " + str(projectId))

        """ Release Redis connection pool"""

        return token


async def retrieve_block_data(block_dag, redis_conn, data_flag=0):
    """
        A function which will get dag block from ipfs and also increment its hits
        Args:
            block_dag:str - The cid of the dag block that needs to be retrieved
            redis_conn - The redis_conn which will be used to access redis
            data_flag:int - This is a flag which can take three values:
                0 - Return only the dag block and not its payload data
                1 - Return the dag block along with its payload data
                2 - Return only the payload data 
    """

    assert data_flag in range(0, 3), f"The value of data: {data_flag} is invalid. It can take values: 0, 1 or 2"

    """ Increment the hits of that block """
    block_dag_hits_key = f"hitsDagBlock"
    r = await redis_conn.zincrby(block_dag_hits_key, 1.0, block_dag)
    rest_logger.debug("Block hit for: ")
    rest_logger.debug(block_dag)
    rest_logger.debug(r)

    """ Retrieve the DAG block from ipfs """
    block = ipfs_client.dag.get(block_dag).as_json()
    if data_flag == 0:
        return block

    payload = block['data']

    """ Get the payload Data """
    payload_data = await retrieve_payload_data(block['data']['cid'], redis_conn)
    payload['payload'] = payload_data

    if data_flag == 1:
        block['data'] = payload
        return block

    if data_flag == 2:
        return payload


async def retrieve_payload_data(payload_cid, redis_conn):
    """
        - Given a payload_cid, get its data from ipfs, at the same time increase its hit
    """
    payload_key = f"hitsPayloadData"
    r = await redis_conn.zincrby(payload_key, 1.0, payload_cid)
    rest_logger.debug("Payload Data hit for: ")
    rest_logger.debug(payload_cid)
    rest_logger.debug(r)

    """ Get the payload Data from ipfs """
    payload_data = ipfs_client.cat(payload_cid).decode('utf-8')
    return payload_data


async def get_max_block_height(project_id: str, redis_conn):
    """
        - Given the projectId and redis_conn, get the prev_dag_cid, block height and
        tetative block height of that projectId from redis
    """
    rest_logger.debug("Fetching Data from Redis")

    """ Fetch the cid of latest DAG block along with the latest block height. """
    prev_dag_cid = None
    block_height = None
    last_tentative_block_height = None
    last_snapshot_cid = None

    last_known_dag_cid_key = f'projectID:{project_id}:lastDagCid'
    r = await redis_conn.get(last_known_dag_cid_key)

    last_tentative_block_height_key = f'projectID:{project_id}:tentativeBlockHeight'
    if r:
        """ Retrieve the height of the latest block only if the DAG projectId is not empty """
        prev_dag_cid = r.decode('utf-8')
        last_block_height_key = f'projectID:{project_id}:blockHeight'
        r2 = await redis_conn.get(last_block_height_key)
        if r2:
            block_height = int(r2)
    else:
        prev_dag_cid = ""

    last_snapshot_cid_key = f'projectID:{project_id}:lastSnapshotCid'
    out = await redis_conn.get(last_snapshot_cid_key)
    rest_logger.debug("The last snapshot cid was:")
    rest_logger.debug(out)
    if out:
        last_snapshot_cid = out.decode('utf-8')
    else:
        last_snapshot_cid = ""

    out = await redis_conn.get(last_tentative_block_height_key)
    rest_logger.debug("From Redis Last tentative block height: ")
    rest_logger.debug(out)
    if out:
        last_tentative_block_height = int(out)
    else:
        last_tentative_block_height = 0
    return (prev_dag_cid, block_height, last_tentative_block_height, last_snapshot_cid)


async def create_retrieval_request(project_id: str, from_height: int, to_height: int, data: int, redis_conn):
    request_id = str(uuid4())

    """ Setup the retrievalRequestInfo HashTable """
    retrieval_request_info_key = f"retrievalRequestInfo:{request_id}"
    fields = {
        'projectId': project_id,
        'to_height': to_height,
        'from_height': from_height,
        'data': data
    }

    _ = await redis_conn.hmset_dict(
        key=retrieval_request_info_key,
        **fields
    )
    requests_list_key = f"pendingRetrievalRequests"
    _ = await redis_conn.sadd(requests_list_key, request_id)

    return request_id


async def make_transaction(snapshot_cid, payload_commit_id, token_hash, last_tentative_block_height, project_id,
                           redis_conn, contract):
    """
        - Create a unqiue transaction_id associated with this transaction, 
        and add it to the set of pending transactions
    """
    e_obj = None
    async with async_timeout.timeout(5) as cm:
        try:
            tx_hash_obj = contract.commitRecord(**dict(
                payloadCommitId=payload_commit_id,
                snapshotCid=snapshot_cid,
                apiKeyHash=token_hash,
                projectId=project_id,
                tentativeBlockHeight=last_tentative_block_height,
            ))

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
            rest_logger.debug("The transaction went through successfully")
            rest_logger.debug(tx_hash_obj)

    if e_obj and cm.expired:
        rest_logger.debug("=" * 80)
        rest_logger.debug("The transaction was not succesfull")
        rest_logger.debug("Commit Payload failed to MaticVigil API")
        rest_logger.debug(e_obj)
        rest_logger.debug("=" * 80)

    pendingTransactionsKey = f"projectId:{project_id}:pendingBlockCreation"
    _ = await redis_conn.sadd(pendingTransactionsKey, payload_commit_id)


@app.post('/commit_payload')
@setup_teardown_boilerplate
async def commit_payload(
        request: Request,
        response: Response,
        token: str = Depends(get_project_token),
        redis_conn=None
):
    """ 
        This endpoint handles the following cases
        - If there are no pending dag block creations, then commit the payload
        and return the snapshot-cid, tentative block height and the payload changed flag

        - If there are any pending dag block creations that are left, then Queue up 
        the payload and let a background service handle it.
        
        - If there are more than `N` no.of payloads pending, then trigger a mechanism to block
        further calls to this endpoint until all the pending payloads are committed. This
        number is specified in the settings.json file as 'max_pending_payload_commits'

    """
    req_args = await request.json()
    try:
        payload = req_args['payload']
        project_id = req_args['projectId']
        rest_logger.debug("Extracted payload and projectId from the request: ")
        rest_logger.debug(payload)
        rest_logger.debug("Payload data type: ")
        rest_logger.debug(type(payload))
        rest_logger.debug("ProjectId: ")
        rest_logger.debug(project_id)
    except Exception as e:
        return {'error': "Either payload or projectId"}
    prev_dag_cid = ""
    block_height = 0
    ipfs_table = None
    last_tentative_block_height = None
    last_tentative_block_height_key = f'projectID:{project_id}:tentativeBlockHeight'
    last_snapshot_cid = None
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
        """ 
        Retrieve the block height, last dag cid and tentative block height 
        for the project_id
        """
        prev_dag_cid, block_height, last_tentative_block_height, last_snapshot_cid = \
            await get_max_block_height(project_id, redis_conn)

    rest_logger.debug("Got last tentative block height: ")
    rest_logger.debug(last_tentative_block_height)

    """ 
        - If the prev_dag_cid is empty, then it means that this is the first block that
        will be added to the DAG of the projectId.
    """
    payload_changed = False

    rest_logger.debug('Previous IPLD CID in the DAG: ')
    rest_logger.debug(prev_dag_cid)
    last_tentative_block_height = last_tentative_block_height + 1
    """ The DAG will be created in the Webhook listener script """
    # dag = settings.dag_structure.to_dict()

    """ Instead of adding payload to directly IPFS, stage it to Filecoin and get back the Cid"""
    if type(payload) is dict:
        payload = json.dumps(payload)

    if settings.payload_storage == "FILECOIN":
        powgate_client = PowerGateClient(settings.POWERGATE_CLIENT_ADDR, False)
        stage_res = powgate_client.data.stage_bytes(payload, token=token)
        """ Since the same data may come back for snapshotting, I have added override=True"""
        job = powgate_client.config.apply(stage_res.cid, override=True, token=token)
        snapshot_cid = stage_res.cid
        """ Add the job id to redis. """
        KEY = f"jobStatus:{snapshot_cid}"
        _ = await redis_conn.set(key=KEY, value=job.jobId)
        rest_logger.debug("Pushed the payload to filecoin.")
        rest_logger.debug("Job Id: " + job.jobId)
    elif settings.payload_storage == "IPFS":
        if type(payload) is dict:
            snapshot_cid = ipfs_client.add_json(payload)
        else:
            try:
                snapshot_cid = ipfs_client.add_str(str(payload))
            except:
                response.status_code = 400
                return {'success': False, 'error': 'PayloadNotSuppported'}
    if snapshot_cid != last_snapshot_cid:
        payload_changed = True
    payload_cid_key = f"projectID:{project_id}:payloadCids"
    _ = await redis_conn.zadd(
        key=payload_cid_key,
        score=last_tentative_block_height,
        member=snapshot_cid
    )

    """ Create a unique identifier for this payload """
    payload_data = {
        'tentativeBlockHeight': last_tentative_block_height,
        'payloadCid': snapshot_cid,
        'projectId': project_id
    }
    payload_commit_id = '0x' + keccak(text=json.dumps(payload_data)).hex()
    rest_logger.debug("Created the unique payload commit id: ")
    rest_logger.debug(payload_commit_id)

    """ Check if there are any pending Block creations left or any pending payloads left to commit """
    pending_block_creations_key = f"projectId:{project_id}:pendingBlockCreation"
    pending_payload_commits_key = f"pendingPayloadCommits"
    pending_block_creations = await redis_conn.smembers(pending_block_creations_key)
    pending_payload_commits = await redis_conn.lrange(pending_payload_commits_key, 0, -1)
    if (len(pending_block_creations) > 0) or (len(pending_payload_commits) > 0):
        rest_logger.debug("There are pending block creations or pending payloads to commit...")
        rest_logger.debug(
            "Queuing up this payload to be processed once all blocks are created or all the pending payloads are committed.")

        """ Create the Hash table for Payload """
        payload_commit_key = f"payloadCommit:{payload_commit_id}"
        fields = {
            'snapshotCid': snapshot_cid,
            'projectId': project_id,
            'tentativeBlockHeight': last_tentative_block_height,
        }

        _ = await redis_conn.hmset_dict(
            key=payload_commit_key,
            **fields
        )

        """ Add this payload commit for pending payload commits list """
        _ = await redis_conn.lpush(pending_payload_commits_key, payload_commit_id)


    else:
        """ Since there are no pending payloads or block creations, commit this payload directly """
        rest_logger.debug('Payload CID')
        rest_logger.debug(snapshot_cid)
        snapshot = dict()
        snapshot['cid'] = snapshot_cid
        snapshot['type'] = "COLD_FILECOIN"
        """ Check if the payload has changed. """

        token_hash = '0x' + keccak(text=json.dumps(snapshot)).hex()
        _ = await make_transaction(
            snapshot_cid=snapshot_cid,
            token_hash=token_hash,
            payload_commit_id=payload_commit_id,
            last_tentative_block_height=last_tentative_block_height,
            project_id=project_id,
            redis_conn=redis_conn,
            contract=request.app.contract
        )

    _ = await redis_conn.set(last_tentative_block_height_key, last_tentative_block_height)
    last_snapshot_cid_key = f'projectID:{project_id}:lastSnapshotCid'
    rest_logger.debug("Setting the last snapshot_cid as: ")
    rest_logger.debug(snapshot_cid)
    _ = await redis_conn.set(last_snapshot_cid_key, snapshot_cid)

    return {
        'cid': snapshot_cid,
        'tentativeHeight': last_tentative_block_height,
        'payloadChanged': payload_changed,
    }


@app.get('/requests/{request_id:str}')
@setup_teardown_boilerplate
async def request_status(
        request: Request,
        response: Response,
        request_id: str,
        redis_conn=None
):
    """
        Given a request_id, return either the status of that request or retrieve all the payloads for that
    """


    # Check if the request is already in the pending list
    requests_list_key = f"pendingRetrievalRequests"
    out = await redis_conn.sismember(requests_list_key, request_id)
    if out == 1:
        return {'requestId': request_id, 'requestStatus': 'Pending'}

    # Get all the retrieved files
    retrieval_files_key = f"retrievalRequestFiles:{request_id}"
    retrieved_files = await redis_conn.zrange(
        key=retrieval_files_key,
        start=0,
        stop=-1,
        withscores=True
    )

    if not retrieved_files:
        response.status_code = 400
        return {'error': 'Invalid requestId'}

    data = {}
    files = []
    for file_, block_height in retrieved_files:
        file_ = file_.decode('utf-8')
        cid = file_.split('/')[-1]
        block_height = int(block_height)

        dag_block = {
            'dagCid': cid,
            'payloadFile': '/' + file_,
            'height': block_height
        }
        files.append(dag_block)

    data['requestId'] = request_id
    data['requestStatus'] = 'Completed'
    data['files'] = files
    response.status_code = 200
    return data


# TODO: get API key/token specific updates corresponding to projects committed with those credentials
@app.get('/projects/updates')
@setup_teardown_boilerplate
async def get_latest_project_updates(
        request: Request,
        response: Response,
        namespace: str = Query(default=None),
        maxCount: int = Query(default=20),
        redis_conn=None
):
    project_diffs_snapshots = list()
    if settings.METADATA_CACHE == 'redis':
        h = await redis_conn.hgetall('auditprotocol:lastSeenSnapshots')
        if len(h) < 1:
            return {}
        for i, d in h.items():
            project_id = i.decode('utf-8')
            diff_data = json.loads(d)
            each_project_info = {
                'projectId': project_id,
                'diff_data': diff_data
            }
            if namespace and namespace in project_id:
                project_diffs_snapshots.append(each_project_info)
            if not namespace:
                try:
                    project_id_int = int(project_id)
                except:
                    pass
                else:
                    project_diffs_snapshots.append(each_project_info)
    sorted_project_diffs_snapshots = sorted(project_diffs_snapshots, key=lambda x: x['diff_data']['cur']['timestamp'],
                                            reverse=True)
    return sorted_project_diffs_snapshots[:maxCount]


@app.get('/{projectId:str}/payloads/cachedDiffs/count')
@setup_teardown_boilerplate
async def get_payloads_diff_counts(
        request: Request,
        response: Response,
        projectId: str,
        from_height: int = Query(default=1),
        to_height: int = Query(default=-1),
        maxCount: int = Query(default=10),
        redis_conn=None
):
    max_block_height = 0
    if settings.METADATA_CACHE == 'redis':
        h = await redis_conn.get(f'projectID:{projectId}:blockHeight')
        if not h:
            max_block_height = 0
        else:
            max_block_height = int(h.decode('utf-8'))

    if to_height == -1:
        to_height = max_block_height
    if (from_height <= 0) or (to_height > max_block_height) or (from_height > to_height):
        return {'error': 'Invalid Height'}
    if settings.METADATA_CACHE == 'redis':
        diff_snapshots_cache_zset = f'projectID:{projectId}:diffSnapshots'
        r = await redis_conn.zcard(diff_snapshots_cache_zset)
        if not r:
            return {'count': 0}
        else:
            try:
                return {'count': int(r)}
            except:
                return {'count': None}


# TODO: get API key/token specific updates corresponding to projects committed with those credentials

@app.get('/projects/updates')
@setup_teardown_boilerplate
async def get_latest_project_updates(
        request: Request,
        response: Response,
        namespace: str = Query(default=None),
        maxCount: int = Query(default=20),
        redis_conn=None
):
    project_diffs_snapshots = list()
    if settings.METADATA_CACHE == 'redis':
        h = await redis_conn.hgetall('auditprotocol:lastSeenSnapshots')
        if len(h) < 1:
            return {}
        for i, d in h.items():
            project_id = i.decode('utf-8')
            diff_data = json.loads(d)
            each_project_info = {
                'projectId': project_id,
                'diff_data': diff_data
            }
            if namespace and namespace in project_id:
                project_diffs_snapshots.append(each_project_info)
            if not namespace:
                try:
                    project_id_int = int(project_id)
                except:
                    pass
                else:
                    project_diffs_snapshots.append(each_project_info)
    sorted_project_diffs_snapshots = sorted(project_diffs_snapshots, key=lambda x: x['diff_data']['cur']['timestamp'],
                                            reverse=True)
    return sorted_project_diffs_snapshots[:maxCount]


@app.get('/{projectId:str}/payloads/cachedDiffs')
@setup_teardown_boilerplate
async def get_payloads_diffs(
        request: Request,
        response: Response,
        projectId: str,
        from_height: int = Query(default=1),
        to_height: int = Query(default=-1),
        maxCount: int = Query(default=10),
        redis_conn=None
):
    max_block_height = 0
    if settings.METADATA_CACHE == 'redis':
        h = await redis_conn.get(f'projectID:{projectId}:blockHeight')
        if not h:
            max_block_height = 0
        else:
            max_block_height = int(h.decode('utf-8'))

    if to_height == -1:
        to_height = max_block_height
    if (from_height <= 0) or (to_height > max_block_height) or (from_height > to_height):
        return {'error': 'Invalid Height'}
    extracted_count = 0
    diff_response = list()
    if settings.METADATA_CACHE == 'redis':
        diff_snapshots_cache_zset = f'projectID:{projectId}:diffSnapshots'
        r = await redis_conn.zrevrangebyscore(
            key=diff_snapshots_cache_zset,
            min=from_height,
            max=to_height,
            withscores=False
        )
        for diff in r:
            if extracted_count >= maxCount:
                break
            diff_response.append(json.loads(diff))
            extracted_count += 1
    return {
        'count': extracted_count,
        'diffs': diff_response
    }


@app.get('/{projectId:str}/payloads')
@setup_teardown_boilerplate
async def get_payloads(
        request: Request,
        response: Response,
        projectId: str,
        from_height: int = Query(None),
        to_height: int = Query(None),
        data: Optional[str] = Query(None),
        redis_conn=None
):
    ipfs_table = None
    max_block_height = None
    redis_conn_raw = None
    if settings.METADATA_CACHE == 'skydb':
        ipfs_table = SkydbTable(table_name=f"{settings.dag_table_name}:{projectId}",
                                columns=['cid'],
                                seed=settings.seed,
                                verbose=1)
        max_block_height = ipfs_table.index - 1
    elif settings.METADATA_CACHE == 'redis':
        h = await redis_conn.get(f'projectID:{projectId}:blockHeight')
        if not h:
            max_block_height = 0
        else:
            max_block_height = int(h.decode('utf-8'))
    if data:
        if data.lower() == 'true':
            data = True
        else:
            data = False

    if (from_height <= 0) or (to_height > max_block_height) or (from_height > to_height):
        response.status_code = 400
        return {'error': 'Invalid Height'}

    if from_height < max_block_height - settings.max_ipfs_blocks:
        """ Create a request Id and start a retrieval request """
        _data = '1' if data else '0'
        request_id = await create_retrieval_request(
            projectId,
            from_height,
            to_height,
            _data,
            redis_conn)

        return {'requestId': request_id}

    blocks = list()  # Will hold the list of blocks in range from_height, to_height
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
        # block = ipfs_client.dag.get(prev_dag_cid).as_json()
        data_flag = 1 if data else 0
        block = await retrieve_block_data(prev_dag_cid, redis_conn, data_flag=data_flag)
        rest_logger.debug("Block Retrieved: ")
        rest_logger.debug(block)
        formatted_block = dict()
        formatted_block['dagCid'] = prev_dag_cid
        formatted_block.update({k: v for k, v in block.items()})
        formatted_block['prevDagCid'] = formatted_block.pop('prevCid')

        # Get the diff_map between the current and previous snapshot
        if prev_payload_cid:
            if prev_payload_cid != block['data']['cid']:
                blocks[idx - 1]['payloadChanged'] = True
                diff_key = f"CidDiff:{prev_payload_cid}:{block['data']['cid']}"
                diff_b = await redis_conn.get(diff_key)
                diff_map = dict()
                if not diff_b:
                    # diff not cached already
                    rest_logger.debug('Diff not cached | New CID | Old CID')
                    rest_logger.debug(blocks[idx - 1]['data']['cid'])
                    rest_logger.debug(block['data']['cid'])

                    """ If the payload is not yet retrieved, then get if from ipfs """
                    if 'payload' in formatted_block['data'].keys():
                        prev_data = formatted_block['data']['payload']
                    else:
                        prev_data = await retrieve_payload_data(block['data']['cid'], redis_conn)
                    rest_logger.debug("Got the payload data: ")
                    rest_logger.debug(prev_data)
                    prev_data = json.loads(prev_data)

                    if 'payload' in blocks[idx - 1]['data'].keys():
                        cur_data = blocks[idx - 1]['data']['payload']
                    else:
                        cur_data = await retrieve_payload_data(block['data']['cid'], redis_conn)
                    cur_data = json.loads(cur_data)

                    # calculate diff
                    for k, v in cur_data.items():
                        if k not in prev_data.keys():
                            rest_logger.info('Ignoring key in older payload as it is not present')
                            rest_logger.info(k)
                            blocks[idx - 1]['payloadChanged'] = False
                            continue
                        if v != prev_data[k]:
                            diff_map[k] = {'old': prev_data[k], 'new': v}

                    if len(diff_map):
                        rest_logger.debug('Found diff in first time calculation')
                        rest_logger.debug(diff_map)
                        blocks[idx - 1]['payloadChanged'] = True
                    # cache in redis
                    await redis_conn.set(diff_key, json.dumps(diff_map))
                else:
                    diff_map = json.loads(diff_b)
                    rest_logger.debug('Found Diff in Cache! | New CID | Old CID | Diff')
                    rest_logger.debug(blocks[idx - 1]['data']['cid'])
                    rest_logger.debug(block['data']['cid'])
                    rest_logger.debug(diff_map)
                blocks[idx - 1]['diff'] = diff_map
            else:  # If the cid of current snapshot is the same as that of the previous snapshot
                blocks[idx - 1]['payloadChanged'] = False
        prev_payload_cid = block['data']['cid']
        blocks.append(formatted_block)
        prev_dag_cid = formatted_block['prevDagCid']
        current_height = current_height - 1
        idx += 1
    return blocks


@app.get('/{projectId}/payloads/height')
@setup_teardown_boilerplate
async def payload_height(
        request: Request,
        response: Response,
        projectId: str,
        redis_conn=None
):
    max_block_height = -1
    if settings.METADATA_CACHE == 'skydb':
        ipfs_table = SkydbTable(table_name=f"{settings.dag_table_name}:{projectId}",
                                columns=['cid'],
                                seed=settings.seed)
        max_block_height = ipfs_table.index - 1
    elif settings.METADATA_CACHE == 'redis':
        h = await redis_conn.get(f'projectID:{projectId}:blockHeight')
        if not h:
            max_block_height = 0
        else:
            max_block_height = int(h.decode('utf-8'))
        rest_logger.debug(max_block_height)

    return {"height": max_block_height}


@app.get('/{projectId}/payload/{block_height}')
@setup_teardown_boilerplate
async def get_block(
        request: Request,
        response: Response,
        projectId: str,
        block_height: int,
        redis_conn=None
):
    """ This endpoint is responsible for retrieving only the dag block and not the payload """
    if settings.METADATA_CACHE == 'skydb':
        ipfs_table = SkydbTable(table_name=f"{settings.dag_table_name}:{projectId}",
                                columns=['cid'],
                                seed=settings.seed,
                                verbose=1)

        if (block_height > ipfs_table.index - 1) or (block_height <= 0):
            response.status_code = 400
            return {'error': 'Invalid block Height'}

        else:
            row = ipfs_table.fetch_row(row_index=block_height)
            block = ipfs_client.dag.get(row['cid']).as_json()
            return {row['cid']: block}
    elif settings.METADATA_CACHE == 'redis':
        max_block_height = await redis_conn.get(f"projectID:{projectId}:blockHeight")
        if not max_block_height:
            response.status_code = 400
            return {'error': 'Block does not exist at this block height'}
        max_block_height = int(max_block_height.decode('utf-8'))
        rest_logger.debug(max_block_height)
        if (block_height > max_block_height) or (block_height <= 0):
            response.status_code = 400
            return {'error': 'Invalid Block Height'}

        if block_height < max_block_height - settings.max_ipfs_blocks:
            rest_logger.debug("Block being fetched at height: ")
            rest_logger.debug(block_height)

            from_height = block_height
            to_height = block_height
            _data = 0  # This means, fetch only the DAG Block
            request_id = await create_retrieval_request(
                projectId,
                from_height,
                to_height,
                _data,
                redis_conn)

            return {'requestId': request_id}

        """ Access the block at block_height """
        project_cids_key_zset = f'projectID:{projectId}:Cids'
        r = await redis_conn.zrangebyscore(
            key=project_cids_key_zset,
            min=block_height,
            max=block_height,
            withscores=False
        )

        prev_dag_cid = r[0].decode('utf-8')

        block = await retrieve_block_data(prev_dag_cid, redis_conn, data_flag=0)

        return {prev_dag_cid: block}


@app.get('/{projectId:str}/payload/{block_height:int}/data')
@setup_teardown_boilerplate
async def get_block_data(
        request: Request,
        response: Response,
        projectId: str,
        block_height: int,
        redis_conn=None
):
    if settings.METADATA_CACHE == 'skydb':
        ipfs_table = SkydbTable(table_name=f"{settings.dag_table_name}:{projectId}",
                                columns=['cid'],
                                seed=settings.seed,
                                verbose=1)
        if (block_height > ipfs_table.index - 1) or (block_height <= 0):
            return {'error': 'Invalid block Height'}
        row = ipfs_table.fetch_row(row_index=block_height)
        block = ipfs_client.dag.get(row['cid']).as_json()
        block['data']['payload'] = ipfs_client.cat(block['data']['cid']).decode()
        return {row['cid']: block['data']}

    elif settings.METADATA_CACHE == "redis":
        max_block_height = await redis_conn.get(f"projectID:{projectId}:blockHeight")
        if not max_block_height:
            response.status_code = 400
            return {'error': 'Invalid Block Height'}
        max_block_height = int(max_block_height.decode('utf-8'))
        if (block_height > max_block_height) or (block_height <= 0):
            response.status_code = 400
            return {'error': 'Invalid Block Height'}

        if block_height < max_block_height - settings.max_ipfs_blocks:
            rest_logger.debug("Block is being fetched at height: ")
            rest_logger.debug(block_height)
            from_height = block_height
            to_height = block_height
            _data = 2  # Get data only and not the block itself
            request_id = await create_retrieval_request(
                projectId,
                from_height,
                to_height,
                _data,
                redis_conn
            )

            return {'requestId': request_id}

        project_cids_key_zset = f'projectID:{projectId}:Cids'
        r = await redis_conn.zrangebyscore(
            key=project_cids_key_zset,
            min=block_height,
            max=block_height,
            withscores=False
        )
        prev_dag_cid = r[0].decode('utf-8')

        payload = await retrieve_block_data(prev_dag_cid, redis_conn, 2)

        """ Return the payload data """
        return {prev_dag_cid: payload}
