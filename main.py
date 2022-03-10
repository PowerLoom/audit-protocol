from typing import Optional
from fastapi import Depends, FastAPI, Request, Response, Query
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from eth_utils import keccak
from maticvigil.EVCore import EVCore
from maticvigil.exceptions import EVBaseException
import logging
import sys
import json
import aioredis
import redis
import time
from config import settings
from pygate_grpc.client import PowerGateClient
from uuid import uuid4
import requests
import async_timeout
from utils.redis_conn import inject_reader_redis_conn, inject_writer_redis_conn
from utils.ipfs_async import client as ipfs_client
from utils.diffmap_utils import process_payloads_for_diff, preprocess_dag
from data_models import ContainerData
from pydantic import ValidationError
import asyncio
from functools import partial
from utils import helper_functions
from utils import redis_keys
from utils.redis_conn import REDIS_WRITER_CONN_CONF, REDIS_READER_CONN_CONF
from utils import retrieval_utils


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
    # app.evc = EVCore(verbose=True)
    # app.contract = app.evc.generate_contract_sdk(
    #     contract_address=settings.audit_contract,
    #     app_name='auditrecords'
    # )


@inject_reader_redis_conn
@inject_writer_redis_conn
async def get_project_token(
        request: Request = None,
        projectId: str = None,
        override_settings=False,
        reader_redis_conn=None,
        writer_redis_conn=None

):
    """ From the request body, extract the projectId and return back the powergate token."""
    if projectId is None:
        try:
            req_args = await request.json()
        except json.JSONDecodeError as jerr:
            rest_logger.error(jerr, exc_info=True)
            return {'error': "The payload has to be a json structured payload."}
        projectId = req_args['projectId']

    """ Save the project Id on set """
    _ = await writer_redis_conn.sadd(redis_keys.get_stored_project_ids_key(), projectId)

    if (settings.payload_storage != "FILECOIN") and (settings.block_storage != "FILECOIN") and (
            override_settings is False):
        return ""

    powgate_client = PowerGateClient(settings.powergate_client_addr, False)

    """ Check if there is a filecoin token for the project Id """
    filecoin_token_key = redis_keys.get_filecoin_token_key(project_id=projectId)
    token = await reader_redis_conn.get(filecoin_token_key)
    if not token:
        user = powgate_client.admin.users.create()
        token = user.token
        _ = await writer_redis_conn.set(filecoin_token_key, token)
        rest_logger.debug("Created a token for projectId: ")
        rest_logger.debug(token)

    else:
        token = token.decode('utf-8')
        rest_logger.debug("Retrieved token: ")
        rest_logger.debug(token)
        rest_logger.debug("projectID: ")
        rest_logger.debug(projectId)

    return token


async def get_max_block_height(project_id: str, reader_redis_conn):
    """
        - Given the projectId and redis_conn, get the prev_dag_cid, block height and
        tetative block height of that projectId from redis
    """
    prev_dag_cid = await helper_functions.get_last_dag_cid(project_id)
    block_height = await helper_functions.get_block_height(project_id)
    last_tentative_block_height = await helper_functions.get_tentative_block_height(
        project_id=project_id,
    )
    last_payload_cid = await helper_functions.get_last_payload_cid(project_id)
    return prev_dag_cid, block_height, last_tentative_block_height, last_payload_cid


async def create_retrieval_request(project_id: str, from_height: int, to_height: int, data: int, writer_redis_conn):
    request_id = str(uuid4())

    """ Setup the retrievalRequestInfo HashTable """
    retrieval_request_info_key = redis_keys.get_retrieval_request_info_key(request_id=request_id)
    fields = {
        'projectId': project_id,
        'to_height': to_height,
        'from_height': from_height,
        'data': data
    }

    _ = await writer_redis_conn.hmset_dict(
        key=retrieval_request_info_key,
        **fields
    )
    requests_list_key = f"pendingRetrievalRequests"
    _ = await writer_redis_conn.sadd(requests_list_key, request_id)

    return request_id


async def make_transaction(snapshot_cid, payload_commit_id, token_hash, last_tentative_block_height, project_id,
                           writer_redis_conn, contract):
    """
        - Create a unqiue transaction_id associated with this transaction, 
        and add it to the set of pending transactions
    """
    e_obj = None
    try:
        loop = asyncio.get_event_loop()
    except Exception as e:
        rest_logger.warning("There was an error while trying to get event loop")
        rest_logger.error(e, exc_info=True)
        return -1
    kwargs = dict(
        payloadCommitId=payload_commit_id,
        snapshotCid=snapshot_cid,
        apiKeyHash=token_hash,
        projectId=project_id,
        tentativeBlockHeight=last_tentative_block_height,
    )
    partial_func = partial(contract.commitRecord, **kwargs)
    try:
        async with async_timeout.timeout(5) as cm:
            try:

                tx_hash_obj = await loop.run_in_executor(None, partial_func)

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
    except asyncio.exceptions.CancelledError as cerr:
        rest_logger.debug("Transcation task cancelled")
        return -1
    except asyncio.exceptions.TimeoutError as terr:
        rest_logger.debug("The transaction timed-out")
        return -1

    if e_obj or cm.expired:
        rest_logger.debug("=" * 80)
        rest_logger.debug("The transaction was not succesfull")
        rest_logger.debug("Commit Payload failed to MaticVigil API")
        rest_logger.debug(e_obj)
        rest_logger.debug("=" * 80)
        return -1

    pending_transaction_key = f"projectID:{project_id}:pendingTransactions"
    tx_hash = tx_hash_obj[0]['txHash']

    _ = await writer_redis_conn.zadd(
            key=pending_transaction_key,
            score=int(last_tentative_block_height),
            member=tx_hash
     )
    return 1


@app.post('/commit_payload')
@inject_reader_redis_conn
@inject_writer_redis_conn
async def commit_payload(
        request: Request,
        response: Response,
        token: str = Depends(get_project_token),
        reader_redis_conn=None,
        writer_redis_conn=None
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

    last_tentative_block_height_key = redis_keys.get_tentative_block_height_key(project_id)

    """ 
    Retrieve the block height, last dag cid and tentative block height 
    for the project_id
    """
    prev_dag_cid, block_height, last_tentative_block_height, last_snapshot_cid = \
        await get_max_block_height(project_id, reader_redis_conn=reader_redis_conn)

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

    """ Instead of adding payload to directly IPFS, stage it to Filecoin and get back the Cid"""
    if type(payload) is dict:
        payload = json.dumps(payload)

    if type(payload) is dict:
        snapshot_cid = await ipfs_client.add_json(payload)
    else:
        try:
            snapshot_cid = await ipfs_client.add_str(str(payload))
        except:
            response.status_code = 400
            return {'success': False, 'error': 'PayloadNotSuppported'}

    if last_snapshot_cid != "":
        if snapshot_cid != last_snapshot_cid:
            payload_changed = True
    payload_cid_key = redis_keys.get_payload_cids_key(project_id)
    _ = await writer_redis_conn.zadd(
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
    # salt with commit time
    payload_commit_id = '0x' + keccak(text=json.dumps(payload_data)+str(time.time())).hex()
    rest_logger.debug("Created the unique payload commit id: ")
    rest_logger.debug(payload_commit_id)

    """ Create the Hash table for Payload """
    payload_commit_key = f"payloadCommit:{payload_commit_id}"
    fields = {
        'snapshotCid': snapshot_cid,
        'projectId': project_id,
        'commitId': payload_commit_id,
        'tentativeBlockHeight': last_tentative_block_height,
    }

    _ = await writer_redis_conn.hmset_dict(
        key=payload_commit_key,
        **fields
    )

    """ Add this payload commit for pending payload commits list """
    pending_payload_commits_key = redis_keys.get_pending_payload_commits_key()
    _ = await writer_redis_conn.lpush(pending_payload_commits_key, payload_commit_id)

    _ = await writer_redis_conn.set(last_tentative_block_height_key, last_tentative_block_height)
    last_snapshot_cid_key = redis_keys.get_last_snapshot_cid_key(project_id)
    rest_logger.debug("Setting the last snapshot_cid as: ")
    rest_logger.debug(snapshot_cid)
    _ = await writer_redis_conn.set(last_snapshot_cid_key, snapshot_cid)

    return {
        'cid': snapshot_cid,
        'tentativeHeight': last_tentative_block_height,
        'commitId': payload_commit_id,
        'payloadChanged': payload_changed,
    }


@app.post('/{projectId:str}/diffRules')
@inject_writer_redis_conn
async def configure_project(
        request: Request,
        response: Response,
        projectId: str,
        writer_redis_conn=None
):
    req_args = await request.json()
    """
    {
        "rules": 
            [
                {
                    "ruleType": "ignore",    
                    "field": "trail",
                    "fieldType": "list",
                    "listMemberType": "map",
                    "ignoreMemberFields": ["chainHeight"]
                }
            ]
    }
    """
    rules = req_args['rules']
    await writer_redis_conn.set(redis_keys.get_diff_rules_key(projectId), json.dumps(rules))
    rest_logger.debug('Set diff rules for project ID')
    rest_logger.debug(projectId)
    rest_logger.debug(rules)


@app.post('/{projectId:str}/confirmations/callback')
@inject_writer_redis_conn
async def configure_project(
        request: Request,
        response: Response,
        projectId: str,
        writer_redis_conn=None
):
    req_json = await request.json()
    await writer_redis_conn.set(f'powerloom:project:{projectId}:callbackURL', req_json['callbackURL'])
    response.status_code = 200
    return {'success': True}


@app.get('/{projectId:str}/getDiffRules')
@inject_reader_redis_conn
async def get_diff_rules(
        request: Request,
        response: Response,
        projectId: str,
        reader_redis_conn=None,
):
    """ This endpoint returs the diffRules set against a projectId """
    out = await helper_functions.check_project_exists(project_id=projectId)
    if out == 0:
        return {'error': 'The projectId provided does not exist'}

    diff_rules_key = redis_keys.get_diff_rules_key(projectId)
    out = await reader_redis_conn.get(diff_rules_key)
    if out is None:
        """ For projectId's who dont have any diffRules, return empty dict"""
        return dict()
    rest_logger.debug(out)
    rules = json.loads(out.decode('utf-8'))
    return rules


@app.get('/requests/{request_id:str}')
@inject_reader_redis_conn
async def request_status(
        request: Request,
        response: Response,
        request_id: str,
        reader_redis_conn=None,
):
    """
        Given a request_id, return either the status of that request or retrieve all the payloads for that
    """


    # Check if the request is already in the pending list
    requests_list_key = f"pendingRetrievalRequests"
    out = await reader_redis_conn.sismember(requests_list_key, request_id)
    if out == 1:
        return {'requestId': request_id, 'requestStatus': 'Pending'}

    # Get all the retrieved files
    retrieval_files_key = redis_keys.get_retrieval_request_files_key(request_id)
    retrieved_files = await reader_redis_conn.zrange(
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
@inject_reader_redis_conn
async def get_latest_project_updates(
        request: Request,
        response: Response,
        namespace: str = Query(default=None),
        maxCount: int = Query(default=20),
        reader_redis_conn=None
):
    project_diffs_snapshots = list()
    h = await reader_redis_conn.hgetall(redis_keys.get_last_seen_snapshots_key())
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
@inject_reader_redis_conn
async def get_payloads_diff_counts(
        request: Request,
        response: Response,
        projectId: str,
        maxCount: int = Query(default=10),
        reader_redis_conn=None
):

    out = await helper_functions.check_project_exists(project_id=projectId)
    if out == 0:
        return {'error': 'The projectId provided does not exist'}

    diff_snapshots_cache_zset = redis_keys.get_diff_snapshots_key(projectId)
    r = await reader_redis_conn.zcard(diff_snapshots_cache_zset)
    if not r:
        return {'count': 0}
    else:
        try:
            return {'count': int(r)}
        except:
            return {'count': None}


# TODO: get API key/token specific updates corresponding to projects committed with those credentials


@app.get('/{projectId:str}/payloads/cachedDiffs')
@inject_reader_redis_conn
async def get_payloads_diffs(
        request: Request,
        response: Response,
        projectId: str,
        from_height: int = Query(default=1),
        to_height: int = Query(default=-1),
        maxCount: int = Query(default=10),
        reader_redis_conn=None
):
    out = await helper_functions.check_project_exists(project_id=projectId)
    if out == 0:
        return {'error': 'The projectId provided does not exist'}

    max_block_height = 0
    max_block_height = await helper_functions.get_block_height(
        project_id=projectId,
        reader_redis_conn=reader_redis_conn
    )

    if to_height == -1:
        to_height = max_block_height
        rest_logger.debug("Max Block Height: ")
        rest_logger.debug(max_block_height)
    if (from_height <= 0) or (to_height > max_block_height) or (from_height > to_height):
        return {'error': 'Invalid Height'}
    extracted_count = 0
    diff_response = list()
    diff_snapshots_cache_zset = redis_keys.get_diff_snapshots_key(projectId)
    r = await reader_redis_conn.zrevrangebyscore(
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
@inject_reader_redis_conn
@inject_writer_redis_conn
async def get_payloads(
        request: Request,
        response: Response,
        projectId: str,
        from_height: int = Query(default=1),
        diffs: Optional[str] = Query(default='true'),  # FIXME: default flag behavior unexpected
        to_height: int = Query(default=-1),
        data: Optional[str] = Query(None),
        reader_redis_conn=None,
        writer_redis_conn=None
):
    """
        - Given the from and to_height do the following steps:
            - Check if there is any intersection between any of the previously
            cached spans.

            - If there is no overlap, then generate the requestID and return it

            - If there is an overlap, then do:
                - If there is any data point that needs to be fetched from a container that
                is not cached on local system
                    -  generate a requestID and return it

                - Split the data into two separate spans: container_fetch_data and ipfs_fetch_data

                    - Now fetch the data present in IPFS and cached containers and put them together
                    to hold data for entire span

                    - return the data

                    - save the span and add a timeout to it.

    """
    out = await helper_functions.check_project_exists(project_id=projectId)

    if out == 0:
        return {'error': 'The projectId provided does not exist'}

    max_block_height = await helper_functions.get_block_height(
        project_id=projectId,
        reader_redis_conn=reader_redis_conn
    )

    if data:
        if data.lower() == 'true':
            data = True
        else:
            data = False

    if diffs:
        # rest_logger.debug('Diffs flag value: %s', diffs)
        if diffs.lower() == 'true':
            diffs = True
            # rest_logger.debug('Converting diff map flag to: %s', diffs)
        else:
            diffs = False
            # rest_logger.debug('Converting diff map flag to: %s', diffs)

    if to_height == -1:
        to_height = max_block_height

    if (from_height <= 0) or (to_height > max_block_height) or (from_height > to_height):
        response.status_code = 400
        return {'error': 'Invalid Height'}

    last_pruned_height = await helper_functions.get_last_pruned_height(project_id=projectId)
    rest_logger.debug('Last pruned height: %s', last_pruned_height)
    rest_logger.debug("Checking max overlap...")

    # TODO: review logic around span, overlap, cached blocks etc. It's a complete shitpile at the moment.
    # for eg: check_overlap() is called once more from fetch_blocks(). Why?
    # ( a span is supposed to be a LRU cache of sorts adjusted within a moving window as DAG blocks keep piling up)
    max_overlap, max_span_id, each_height_spans = await retrieval_utils.check_overlap(
        from_height=from_height,
        to_height=to_height,
        project_id=projectId
    )
    # rest_logger.debug("Max overlap, Max Span ID, Each height spans:")
    # rest_logger.debug(max_overlap)
    # rest_logger.debug(max_span_id)
    # rest_logger.debug(each_height_spans)

    if (max_overlap == 0.0) and (from_height <= last_pruned_height):
        rest_logger.debug("Creating a retrieval request")
        _data = 1 if data else 0
        request_id = await create_retrieval_request(
            project_id=projectId,
            from_height=from_height,
            to_height=to_height,
            data=_data,
            writer_redis_conn=writer_redis_conn)

        return {'requestId': request_id}

    # Get the containers required and the cached values
    containers, cached = await retrieval_utils.check_containers(
        from_height=from_height,
        to_height=to_height,
        each_height_spans=each_height_spans,
        project_id=projectId
    )

    rest_logger.debug("Containers Required: ")
    rest_logger.debug(containers)

    if len(containers) > 0:
        rest_logger.debug("Creating a retrieval request")
        _data = 1 if data else 0
        request_id = await create_retrieval_request(
            project_id=projectId,
            from_height=from_height,
            to_height=to_height,
            data=_data,
            writer_redis_conn=writer_redis_conn)

        return {'requestId': request_id}

    dag_blocks = await retrieval_utils.fetch_blocks(
        from_height=from_height,
        to_height=to_height,
        project_id=projectId,
        data_flag=data
    )

    current_height = to_height
    prev_dag_cid = ""
    prev_payload_cid = None
    idx = 0
    blocks = list()
    while current_height >= from_height:
        # rest_logger.debug("Fetching block at height: %s", current_height)
        if not prev_dag_cid:
            project_cids_key_zset = redis_keys.get_dag_cids_key(projectId)
            r = await reader_redis_conn.zrangebyscore(
                key=project_cids_key_zset,
                min=current_height,
                max=current_height,
                withscores=False
            )
            if r:
                prev_dag_cid = r[0].decode('utf-8')
            else:
                return {'error': 'NoRecordsFound'}
        data_flag = 1 if data else 0
        # NOTE: not yet clear why the earlier call to retrieval_utils.fetch_blocks() would not populate `dag_blocks` map
        if dag_blocks.get(prev_dag_cid) is None:
            # rest_logger.debug("Fetching block from IPFS")
            block = await retrieval_utils.retrieve_block_data(prev_dag_cid, writer_redis_conn=writer_redis_conn, data_flag=data_flag)
        else:
            # rest_logger.debug("Block already fetched")
            block = dag_blocks.get(prev_dag_cid)
        # rest_logger.debug("Block Retrieved: ")
        # rest_logger.debug(block)
        formatted_block = dict()
        formatted_block['dagCid'] = prev_dag_cid
        formatted_block.update({k: v for k, v in block.items()})
        formatted_block['prevDagCid'] = formatted_block.pop('prevCid')

        # Get the diff_map between the current and previous snapshot
        # rest_logger.debug('Diff flag set as: %s', diffs)
        if diffs:
            if prev_payload_cid:
                if prev_payload_cid != block['data']['cid']:
                    blocks[idx - 1]['payloadChanged'] = False
                    diff_key = f"CidDiff:{prev_payload_cid}:{block['data']['cid']}"
                    diff_b = await reader_redis_conn.get(diff_key)
                    diff_map = dict()
                    if not diff_b:
                        # diff not cached already
                        # rest_logger.debug('Diff not cached | New CID | Old CID')
                        # rest_logger.debug(blocks[idx - 1]['data']['cid'])
                        # rest_logger.debug(block['data']['cid'])

                        """ If the payload is not yet retrieved, then get if from ipfs """
                        if 'payload' in formatted_block['data'].keys():
                            prev_data = formatted_block['data']['payload']
                        else:
                            prev_data = await retrieval_utils.retrieve_payload_data(block['data']['cid'], writer_redis_conn=writer_redis_conn)
                        # rest_logger.debug("Got the payload data: ")
                        # rest_logger.debug(prev_data)

                        if 'payload' in blocks[idx - 1]['data'].keys():
                            cur_data = blocks[idx - 1]['data']['payload']
                        else:
                            cur_data = await retrieval_utils.retrieve_payload_data(
                                blocks[idx-1]['data']['cid'],
                                writer_redis_conn=writer_redis_conn
                            )

                        result = await process_payloads_for_diff(
                            project_id=projectId,
                            prev_data=prev_data,
                            cur_data=cur_data,
                        )
                        # rest_logger.debug('After payload clean up and comparison if any')
                        # rest_logger.debug(result)
                        cur_data_copy = result['cur_copy']
                        prev_data_copy = result['prev_copy']

                        # calculate diff
                        for k, v in cur_data_copy.items():
                            if k in result['payload_changed'] and result['payload_changed'][k]:
                                diff_map[k] = {
                                    'old': prev_data.get(k),
                                    'new': cur_data.get(k)
                                }

                        if len(diff_map):
                            rest_logger.debug('Found diff in first time calculation')
                            rest_logger.debug(diff_map)
                        # cache in redis
                        await writer_redis_conn.set(diff_key, json.dumps(diff_map))
                    else:
                        diff_map = json.loads(diff_b)
                        # rest_logger.debug('Found Diff in Cache! | New CID | Old CID | Diff')
                        # rest_logger.debug(blocks[idx - 1]['data']['cid'])
                        # rest_logger.debug(block['data']['cid'])
                        # rest_logger.debug(diff_map)
                    blocks[idx - 1]['diff'] = diff_map
                    if len(diff_map.keys()):
                        blocks[idx - 1]['payloadChanged'] = True
                else:  # If the cid of current snapshot is the same as that of the previous snapshot
                    blocks[idx - 1]['payloadChanged'] = False
            prev_payload_cid = block['data']['cid']
            blocks.append(formatted_block)
            prev_dag_cid = formatted_block['prevDagCid']
            current_height = current_height - 1
            idx += 1
    return blocks


@app.get('/{projectId}/payloads/height')
@inject_reader_redis_conn
async def payload_height(
        request: Request,
        response: Response,
        projectId: str,
        reader_redis_conn=None
):

    max_block_height = -1
    max_block_height = await helper_functions.get_block_height(
        project_id=projectId,
        reader_redis_conn=reader_redis_conn
    )
    rest_logger.debug(max_block_height)

    return {"height": max_block_height}


@app.get('/{projectId}/payload/{block_height}')
@inject_reader_redis_conn
@inject_writer_redis_conn
async def get_block(
        request: Request,
        response: Response,
        projectId: str,
        block_height: int,
        reader_redis_conn=None,
        writer_redis_conn=None

):
    out = await helper_functions.check_project_exists(project_id=projectId)
    if out == 0:
        return {'error': 'The projectId provided does not exist'}

    """ This endpoint is responsible for retrieving only the dag block and not the payload """
    max_block_height = await helper_functions.get_block_height(
        project_id=projectId,
        reader_redis_conn=reader_redis_conn
    )
    if max_block_height == 0:
        response.status_code = 400
        return {'error': 'Block does not exist at this block height'}

    rest_logger.debug(max_block_height)
    if (block_height > max_block_height) or (block_height <= 0):
        response.status_code = 400
        return {'error': 'Invalid Block Height'}

    last_pruned_height = await helper_functions.get_last_pruned_height(
        project_id=projectId,
        reader_redis_conn=reader_redis_conn,
        writer_redis_conn=writer_redis_conn
    )

    if block_height < last_pruned_height:
        rest_logger.debug("Block being fetched at height: ")
        rest_logger.debug(block_height)

        from_height = block_height
        to_height = block_height
        _data = 0  # This means, fetch only the DAG Block
        request_id = await create_retrieval_request(
            project_id=projectId,
            from_height=from_height,
            to_height=to_height,
            data=_data,
            writer_redis_conn=writer_redis_conn,
        )

        return {'requestId': request_id}

    """ Access the block at block_height """
    project_cids_key_zset = f'projectID:{projectId}:Cids'
    r = await reader_redis_conn.zrangebyscore(
        key=project_cids_key_zset,
        min=block_height,
        max=block_height,
        withscores=False
    )

    prev_dag_cid = r[0].decode('utf-8')

    block = await retrieval_utils.retrieve_block_data(prev_dag_cid, writer_redis_conn=writer_redis_conn, data_flag=0)

    return {prev_dag_cid: block}


@app.get('/{projectId:str}/payload/{block_height:int}/data')
@inject_reader_redis_conn
@inject_writer_redis_conn
async def get_block_data(
        request: Request,
        response: Response,
        projectId: str,
        block_height: int,
        writer_redis_conn=None,
        reader_redis_conn=None,
):
    out = await helper_functions.check_project_exists(project_id=projectId)
    if out == 0:
        return {'error': 'The projectId provided does not exist'}

    max_block_height = await helper_functions.get_block_height(
        project_id=projectId,
        reader_redis_conn=reader_redis_conn
    )

    if max_block_height == 0:
        response.status_code = 400
        return {'error': 'No Block exists for this project'}

    if (block_height > max_block_height) or (block_height <= 0):
        response.status_code = 400
        return {'error': 'Invalid Block Height'}

    last_pruned_height = await helper_functions.get_last_pruned_height(
        project_id=projectId,
        reader_redis_conn=reader_redis_conn,
        writer_redis_conn=writer_redis_conn
    )
    if block_height < last_pruned_height:
        from_height = block_height
        to_height = block_height
        _data = 2  # Get data only and not the block itself
        request_id = await create_retrieval_request(
            project_id=projectId,
            from_height=from_height,
            to_height=to_height,
            data=_data,
            writer_redis_conn=writer_redis_conn,
        )

        return {'requestId': request_id}

    project_cids_key_zset = redis_keys.get_dag_cids_key(project_id=projectId)
    r = await reader_redis_conn.zrangebyscore(
        key=project_cids_key_zset,
        min=block_height,
        max=block_height,
        withscores=False
    )
    prev_dag_cid = r[0].decode('utf-8')

    payload = await retrieval_utils.retrieve_block_data(prev_dag_cid, writer_redis_conn=writer_redis_conn, data_flag=2)

    """ Return the payload data """
    return {prev_dag_cid: payload}


# Get the containerData using container_id
@app.get("/query/containerData/{container_id:str}")
@inject_reader_redis_conn
async def get_container_data(
        request: Request,
        response: Response,
        container_id: str,
        reader_redis_conn=None
):
    """
        - retrieve the containerData from containerData key
        - return containerData
    """

    rest_logger.debug("Retrieving containerData for container_id: ")
    rest_logger.debug(container_id)
    container_data_key = f"containerData:{container_id}"
    out = await reader_redis_conn.hgetall(container_data_key)
    out = {k.decode('utf-8'): v.decode('utf-8') for k, v in out.items()}
    if not out:
        return {"error": f"The container_id:{container_id} is invalid"}
    try:
        container_data = ContainerData(**out)
    except ValidationError as verr:
        rest_logger.debug("The containerData retrieved from redis is invalid")
        rest_logger.debug(out)
        rest_logger.error(verr, exc_info=True)
        return {}

    return container_data.dict()


@app.get("/query/executingContainers")
@inject_reader_redis_conn
async def get_executing_containers(
        request: Request,
        response: Response,
        maxCount: int = Query(default=10),
        data: str = Query(default="false"),
        reader_redis_conn=None
):
    """
        - Get all the container_id's from the executingContainers redis SET
        - if the data field is true, then get the containerData for each of the container as well
    """

    if isinstance(data, str):
        if data.lower() == "true":
            data = True
        else:
            data = False
    else:
        data = False

    executing_containers_key = f"executingContainers"
    all_container_ids = await reader_redis_conn.smembers(executing_containers_key)

    containers = list()
    for container_id in all_container_ids:
        container_id = container_id.decode('utf-8')
        if data is True:
            container_data_key = f"containerData:{container_id}"
            out = await reader_redis_conn.hgetall(container_data_key)
            out = {k.decode('utf-8'): v.decode('utf-8') for k, v in out.items()}
            if not out:
                _container = {
                    'containerId': container_id,
                    'containerData': dict()
                }
            else:
                try:
                    container_data = ContainerData(**out)
                except ValidationError as verr:
                    rest_logger.debug("The containerData retrieved from redis is invalid")
                    rest_logger.debug(out)
                    rest_logger.error(verr, exc_info=True)
                    _container = {
                        'containerId': container_id,
                        'containerData': dict()
                    }
                else:
                    _container = {
                        'containerId': container_id,
                        'containerData': container_data.dict()
                    }
            containers.append(_container)
        else:
            containers.append(container_id)

    return dict(count=len(containers), containers=containers)
