from typing import Optional
from fastapi import FastAPI, Request, Response, Query
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from eth_utils import keccak

import utils.diffmap_utils
from config import settings
from uuid import uuid4
from utils.redis_conn import RedisPool
from utils.rabbitmq_utils import get_rabbitmq_connection, get_rabbitmq_channel
from utils import helper_functions
from utils import redis_keys
from functools import partial
from utils import retrieval_utils
from utils.diffmap_utils import process_payloads_for_diff
from data_models import ContainerData, PayloadCommit
from pydantic import ValidationError
from aio_pika import ExchangeType, DeliveryMode, Message
from aio_pika.pool import Pool
import logging
import sys
import json
import aioredis
import redis
import time
import asyncio


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
    app.rmq_connection_pool = Pool(get_rabbitmq_connection, max_size=5, loop=asyncio.get_running_loop())
    app.rmq_channel_pool = Pool(
        partial(get_rabbitmq_channel, app.rmq_connection_pool), max_size=20, loop=asyncio.get_running_loop()
    )
    app.aioredis_pool = RedisPool()
    await app.aioredis_pool.populate()

    app.reader_redis_pool = app.aioredis_pool.reader_redis_pool
    app.writer_redis_pool = app.aioredis_pool.writer_redis_pool

    # app.evc = EVCore(verbose=True)
    # app.contract = app.evc.generate_contract_sdk(
    #     contract_address=settings.audit_contract,
    #     app_name='auditrecords'
    # )


async def get_max_block_height(project_id: str, reader_redis_conn: aioredis.Redis):
    """
        - Given the projectId and redis_conn, get the prev_dag_cid, block height and
        tetative block height of that projectId from redis
    """
    prev_dag_cid = await helper_functions.get_last_dag_cid(project_id=project_id, reader_redis_conn=reader_redis_conn)
    block_height = await helper_functions.get_block_height(project_id=project_id, reader_redis_conn=reader_redis_conn)
    last_tentative_block_height = await helper_functions.get_tentative_block_height(
        project_id=project_id, reader_redis_conn=reader_redis_conn
    )
    last_payload_cid = await helper_functions.get_last_payload_cid(project_id, reader_redis_conn)
    return prev_dag_cid, block_height, last_tentative_block_height, last_payload_cid


async def create_retrieval_request(project_id: str, from_height: int, to_height: int, data: int, writer_redis_conn: aioredis.Redis):
    request_id = str(uuid4())

    """ Setup the retrievalRequestInfo HashTable """
    retrieval_request_info_key = redis_keys.get_retrieval_request_info_key(request_id=request_id)
    fields = {
        'projectId': project_id,
        'to_height': to_height,
        'from_height': from_height,
        'data': data
    }

    _ = await writer_redis_conn.hset(
        name=retrieval_request_info_key,
        mapping=fields
    )
    requests_list_key = f"pendingRetrievalRequests"
    _ = await writer_redis_conn.sadd(requests_list_key, request_id)

    return request_id


@app.post('/commit_payload')
async def commit_payload(
        request: Request,
        response: Response
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
    reader_redis_conn: aioredis.Redis = request.app.reader_redis_pool
    writer_redis_conn: aioredis.Redis = request.app.writer_redis_pool
    req_args = await request.json()
    try:
        payload = req_args['payload']
        project_id = req_args['projectId']
        rest_logger.debug(f"Extracted payload and projectId from the request: {payload} , Payload data type: {type(payload)} ProjectId: {project_id}")
    except Exception as e:
        return {'error': "Either payload or projectId"}

    last_tentative_block_height_key = redis_keys.get_tentative_block_height_key(project_id)

    """ 
    Retrieve the block height, last dag cid and tentative block height 
    for the project_id
    """
    prev_dag_cid, block_height, last_tentative_block_height, last_snapshot_cid = \
        await get_max_block_height(project_id, reader_redis_conn=reader_redis_conn)

    rest_logger.debug(f"Got last tentative block height: {last_tentative_block_height}. Previous IPLD CID in the DAG: {prev_dag_cid}")

    last_tentative_block_height = last_tentative_block_height + 1
    skip_anchor_proof_tx = req_args.get('skipAnchorProof', True)  # skip anchor tx by default, unless passed
    """ Create a unique identifier for this payload """
    payload_data = {
        'tentativeBlockHeight': last_tentative_block_height,
        'payload': payload,
        'projectId': project_id,
    }
    # salt with commit time
    payload_commit_id = '0x' + keccak(text=json.dumps(payload_data)+str(time.time())).hex()
    rest_logger.debug(f"Created the unique payload commit id: {payload_commit_id}")

    web3_storage_flag = req_args.get('web3Storage', False)
    payload_for_commit = PayloadCommit(**{
        'projectId': project_id,
        'commitId': payload_commit_id,
        'payload': payload,
        'tentativeBlockHeight': last_tentative_block_height,
        'web3Storage': web3_storage_flag,
        'skipAnchorProof': skip_anchor_proof_tx
    })

    # push payload for commit to rabbitmq queue
    async with request.app.rmq_channel_pool.acquire() as channel:
        exchange = await channel.get_exchange(
            name=settings.rabbitmq.setup['core']['exchange'],
            # always ensure exchanges and queues are initialized as part of launch sequence, not to be checked here
            ensure=False
        )
        message = Message(
            payload_for_commit.json().encode('utf-8'),
            delivery_mode=DeliveryMode.PERSISTENT,
        )

        await exchange.publish(
            message=message,
            routing_key='commit-payloads'
        )
        rest_logger.debug(
            'Published payload against commit ID %s to RabbitMQ payload commit service queue', payload_commit_id
        )

    _ = await writer_redis_conn.set(last_tentative_block_height_key, last_tentative_block_height)
    # await writer_redis_conn.zadd(
    #     key=redis_keys.get_payload_commit_id_process_logs_zset_key(
    #         project_id=project_id, payload_commit_id=payload_commit_id
    #     ),
    #     member=json.dumps(
    #         {
    #             'worker': 'api_entry',
    #             'update': {
    #                 'action': 'RabbitMQ.Publish',
    #                 'info': {
    #                     'msg': payload_for_commit.dict(),
    #                     'status': 'Success'
    #                 }
    #             }
    #         }
    #     ),
    #     score=int(time.time())
    # )

    return {
        'tentativeHeight': last_tentative_block_height,
        'commitId': payload_commit_id
    }


@app.post('/{projectId:str}/diffRules')
async def configure_project(
        request: Request,
        response: Response,
        projectId: str
):
    writer_redis_conn: aioredis.Redis = request.app.writer_redis_pool
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
    rest_logger.debug(f'Set diff rules {rules} for project ID {projectId}')


@app.post('/{projectId:str}/confirmations/callback')
async def configure_project(
        request: Request,
        response: Response,
        projectId: str
):
    writer_redis_conn: aioredis.Redis = request.app.writer_redis_pool
    req_json = await request.json()
    await writer_redis_conn.set(f'powerloom:project:{projectId}:callbackURL', req_json['callbackURL'])
    response.status_code = 200
    return {'success': True}


@app.get('/{projectId:str}/getDiffRules')
async def get_diff_rules(
        request: Request,
        response: Response,
        projectId: str
):
    """ This endpoint returs the diffRules set against a projectId """
    out = await helper_functions.check_project_exists(
        project_id=projectId, reader_redis_conn=request.app.reader_redis_pool
    )
    if out == 0:
        return {'error': 'The projectId provided does not exist'}

    diff_rules_key = redis_keys.get_diff_rules_key(projectId)
    reader_redis_conn: aioredis.Redis = request.app.reader_redis_pool
    out = await reader_redis_conn.get(diff_rules_key)
    if out is None:
        """ For projectId's who dont have any diffRules, return empty dict"""
        return dict()
    rest_logger.debug(out)
    rules = json.loads(out.decode('utf-8'))
    return rules


@app.get('/requests/{request_id:str}')
async def request_status(
        request: Request,
        response: Response,
        request_id: str
):
    """
        Given a request_id, return either the status of that request or retrieve all the payloads for that
    """

    reader_redis_conn: aioredis.Redis = request.app.reader_redis_pool
    # Check if the request is already in the pending list
    requests_list_key = f"pendingRetrievalRequests"
    out = await reader_redis_conn.sismember(requests_list_key, request_id)
    if out == 1:
        return {'requestId': request_id, 'requestStatus': 'Pending'}

    # Get all the retrieved files
    retrieval_files_key = redis_keys.get_retrieval_request_files_key(request_id)
    retrieved_files = await reader_redis_conn.zrange(
        name=retrieval_files_key,
        start=0,
        end=-1,
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
async def get_latest_project_updates(
        request: Request,
        response: Response,
        namespace: str = Query(default=None),
        maxCount: int = Query(default=20)
):
    project_diffs_snapshots = list()
    reader_redis_conn: aioredis.Redis = request.app.reader_redis_pool
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
async def get_payloads_diff_counts(
        request: Request,
        response: Response,
        projectId: str,
        maxCount: int = Query(default=10)
):
    reader_redis_conn: aioredis.Redis = request.app.reader_redis_pool
    out = await helper_functions.check_project_exists(project_id=projectId, reader_redis_conn=reader_redis_conn)
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
async def get_payloads_diffs(
        request: Request,
        response: Response,
        projectId: str,
        from_height: int = Query(default=1),
        to_height: int = Query(default=-1),
        maxCount: int = Query(default=10)
):
    reader_redis_conn: aioredis.Redis = request.app.reader_redis_pool
    out = await helper_functions.check_project_exists(project_id=projectId, reader_redis_conn=reader_redis_conn)
    if out == 0:
        return {'error': 'The projectId provided does not exist'}

    max_block_height = 0
    max_block_height = await helper_functions.get_block_height(
        project_id=projectId,
        reader_redis_conn=reader_redis_conn
    )

    if to_height == -1:
        to_height = max_block_height
        rest_logger.debug("Max Block Height: %d",max_block_height)
    if (from_height <= 0) or (to_height > max_block_height) or (from_height > to_height):
        return {'error': 'Invalid Height'}
    extracted_count = 0
    diff_response = list()
    diff_snapshots_cache_zset = redis_keys.get_diff_snapshots_key(projectId)
    r = await reader_redis_conn.zrevrangebyscore(
        name=diff_snapshots_cache_zset,
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
async def get_payloads(
        request: Request,
        response: Response,
        projectId: str,
        from_height: int = Query(default=1),
        diffs: Optional[str] = Query(default='true'),  # FIXME: default flag behavior unexpected
        to_height: int = Query(default=-1),
        data: Optional[str] = Query(None)
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
    reader_redis_conn: aioredis.Redis = request.app.reader_redis_pool
    writer_redis_conn: aioredis.Redis = request.app.writer_redis_pool
    out = await helper_functions.check_project_exists(project_id=projectId, reader_redis_conn=reader_redis_conn)

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

    last_pruned_height = await helper_functions.get_last_pruned_height(
        project_id=projectId, reader_redis_conn=reader_redis_conn
    )
    rest_logger.debug('Last pruned height: %s. Checking max overlap...', last_pruned_height)

    # TODO: review logic around span, overlap, cached blocks etc. It's a complete shitpile at the moment.
    # for eg: check_overlap() is called once more from fetch_blocks(). Why?
    # ( a span is supposed to be a LRU cache of sorts adjusted within a moving window as DAG blocks keep piling up)
    max_overlap, max_span_id, each_height_spans = await retrieval_utils.check_overlap(
        from_height=from_height,
        to_height=to_height,
        project_id=projectId,
        reader_redis_conn=reader_redis_conn
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
        project_id=projectId,
        reader_redis_conn=reader_redis_conn
    )

    rest_logger.debug(f"Containers Required: {containers}")

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
        data_flag=data,
        reader_redis_conn=reader_redis_conn
    )

    current_height = to_height
    cur_dag_cid = None
    idx = 0
    blocks = list()
    while current_height >= from_height:
        # rest_logger.debug("Fetching block at height: %s", current_height)
        if not cur_dag_cid:
            project_cids_key_zset = redis_keys.get_dag_cids_key(projectId)
            r = await reader_redis_conn.zrangebyscore(
                name=project_cids_key_zset,
                min=current_height,
                max=current_height,
                withscores=False
            )
            if r:
                cur_dag_cid = r[0].decode('utf-8')
            else:
                return {'error': 'NoRecordsFound'}
        data_flag = 1 if data else 0
        # NOTE: not yet clear why the earlier call to retrieval_utils.fetch_blocks() would not populate `dag_blocks` map
        if dag_blocks.get(cur_dag_cid) is None:
            # rest_logger.debug("Fetching block from IPFS")
            block = await retrieval_utils.retrieve_block_data(cur_dag_cid, writer_redis_conn=writer_redis_conn, data_flag=data_flag)
        else:
            # rest_logger.debug("Block already fetched")
            block = dag_blocks.get(cur_dag_cid)
        # rest_logger.debug("Block Retrieved: ")
        # rest_logger.debug(block)
        formatted_block = dict()
        formatted_block['dagCid'] = cur_dag_cid
        formatted_block.update({k: v for k, v in block.items()})
        formatted_block['prevDagCid'] = formatted_block.pop('prevCid')

        # NOTE: removed a duplicate diff generation logic.
        # Get the diff_map between the current and previous snapshot
        # rest_logger.debug('Diff flag set as: %s', diffs)
        if diffs:
            # FIXME: find a better way to get the entire chunk of diffs within the height range. insert each accordingly
            diff_at_height_r = await reader_redis_conn.zrangebyscore(
                name=redis_keys.get_diff_snapshots_key(projectId),
                min=current_height,
                max=current_height,
                withscores=False
            )
            if diff_at_height_r:
                diff_map = json.loads(diff_at_height_r[0])['diff']
            else:
                diff_map = {}
            formatted_block['diff'] = diff_map
        blocks.append(formatted_block)
        if formatted_block['prevDagCid']:
            cur_dag_cid = formatted_block['prevDagCid']['/']
            # the decrement in current_height will ensure the loop ends here so we dont need to set cur_dag_cid
        current_height = current_height - 1
        idx += 1
    return blocks


@app.get('/{projectId}/payloads/height')
async def payload_height(
        request: Request,
        response: Response,
        projectId: str
):
    reader_redis_conn: aioredis.Redis = request.app.reader_redis_pool
    max_block_height = await helper_functions.get_block_height(
        project_id=projectId,
        reader_redis_conn=reader_redis_conn
    )
    rest_logger.debug(max_block_height)

    return {"height": max_block_height}


@app.get('/{projectId}/payload/{block_height}')
async def get_block(
        request: Request,
        response: Response,
        projectId: str,
        block_height: int

):
    reader_redis_conn: aioredis.Redis = request.app.reader_redis_pool
    writer_redis_conn: aioredis.Redis = request.app.writer_redis_pool
    out = await helper_functions.check_project_exists(project_id=projectId, reader_redis_conn=reader_redis_conn)
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
        reader_redis_conn=reader_redis_conn
    )

    if block_height < last_pruned_height:
        rest_logger.debug("Block being fetched at height: %d", block_height)

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
        name=project_cids_key_zset,
        min=block_height,
        max=block_height,
        withscores=False
    )

    prev_dag_cid = r[0].decode('utf-8')

    block = await retrieval_utils.retrieve_block_data(prev_dag_cid, writer_redis_conn=writer_redis_conn, data_flag=0)

    return {prev_dag_cid: block}


@app.get('/{projectId:str}/payload/{block_height:int}/data')
async def get_block_data(
        request: Request,
        response: Response,
        projectId: str,
        block_height: int
):
    reader_redis_conn: aioredis.Redis = request.app.reader_redis_pool
    writer_redis_conn: aioredis.Redis = request.app.writer_redis_pool
    out = await helper_functions.check_project_exists(project_id=projectId, reader_redis_conn=reader_redis_conn)
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
        reader_redis_conn=reader_redis_conn
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
        name=project_cids_key_zset,
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
async def get_container_data(
        request: Request,
        response: Response,
        container_id: str
):
    """
        - retrieve the containerData from containerData key
        - return containerData
    """

    rest_logger.debug("Retrieving containerData for container_id: %s",container_id)
    container_data_key = f"containerData:{container_id}"
    reader_redis_conn: aioredis.Redis = request.app.reader_redis_pool
    out = await reader_redis_conn.hgetall(container_data_key)
    out = {k.decode('utf-8'): v.decode('utf-8') for k, v in out.items()}
    if not out:
        return {"error": f"The container_id:{container_id} is invalid"}
    try:
        container_data = ContainerData(**out)
    except ValidationError as verr:
        rest_logger.debug(f"The containerData {out} retrieved from redis is invalid with error {verr}", exc_info=True)
        return {}

    return container_data.dict()


@app.get("/query/executingContainers")
async def get_executing_containers(
        request: Request,
        response: Response,
        maxCount: int = Query(default=10),
        data: str = Query(default="false")
):
    """
        - Get all the container_id's from the executingContainers redis SET
        - if the data field is true, then get the containerData for each of the container as well
    """
    reader_redis_conn: aioredis.Redis = request.app.reader_redis_pool
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
                    rest_logger.debug(f"The containerData {out} retrieved from redis is invalid with error {verr}", exc_info=True)
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
