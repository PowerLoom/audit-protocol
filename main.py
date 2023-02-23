from typing import Optional
from fastapi import FastAPI, Request, Response, Query
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from eth_utils import keccak
from async_ipfshttpclient.main import AsyncIPFSClientSingleton
from uuid import uuid4
from utils.redis_conn import RedisPool
from utils.rabbitmq_utils import get_rabbitmq_connection, get_rabbitmq_channel, get_rabbitmq_core_exchange, get_rabbitmq_routing_key
from utils import helper_functions
from utils import redis_keys
from functools import partial
from utils import retrieval_utils
from data_models import PayloadCommit, PayloadCommitAPIRequest, PeerRegistrationRequest, ProjectRegistrationRequest, ProjectRegistrationRequestForIndexing
from config import settings

from pydantic import ValidationError
from aio_pika import DeliveryMode, Message
from aio_pika.pool import Pool
import logging
import sys
import json
from redis import asyncio as aioredis
import redis
import time
import asyncio
import httpx

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

    app.ipfs_singleton = AsyncIPFSClientSingleton()
    await app.ipfs_singleton.init_sessions()
    app.ipfs_read_client = app.ipfs_singleton._ipfs_read_client

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


# Health check endpoint that returns 200 OK
@app.get('/health')
async def health_check():
    return {'status': 'OK'}

@app.post('/commit_payload')
async def commit_payload(
        request: Request,
        response: Response
):
    req_json: dict = await request.json()
    try:
        req_parsed: PayloadCommitAPIRequest = PayloadCommitAPIRequest.parse_obj(req_json)
    except ValidationError as e:
        response.status_code = 400
        rest_logger.error('Got bad request: %s | Parsing error: %s', req_json, e, exc_info=True)
        return {'error': 'Invalid request'}

    project_id = req_parsed.projectId
    out = await helper_functions.check_project_exists(
        project_id=project_id, reader_redis_conn=request.app.reader_redis_pool
    )
    if out == 0:
        return {'error': 'The projectId provided does not exist'}

    skip_anchor_proof_tx = req_parsed.skipAnchorProof
    """ Create a unique identifier for this payload """

    payload_data = {
        'payload': req_parsed.payload,
        'projectId': project_id,
    }
    # salt with commit time
    payload_commit_id = '0x' + keccak(text=json.dumps(payload_data)+str(time.time())).hex()
    rest_logger.debug("Created the unique payload commit id:%s", payload_commit_id)

    payload_for_commit = PayloadCommit(**{
        'projectId': project_id,
        'commitId': payload_commit_id,
        'payload': req_parsed.payload,
        'web3Storage': req_parsed.web3Storage,
        'skipAnchorProof': skip_anchor_proof_tx,
        'sourceChainDetails': req_parsed.sourceChainDetails,
        'requestID': req_parsed.requestID
    })

    # push payload for commit to rabbitmq queue
    async with request.app.rmq_channel_pool.acquire() as channel:
        exchange = await channel.get_exchange(
            name=get_rabbitmq_core_exchange(),
            # always ensure exchanges and queues are initialized as part of launch sequence, not to be checked here
            ensure=False
        )
        message = Message(
            payload_for_commit.json().encode('utf-8'),
            delivery_mode=DeliveryMode.PERSISTENT,
        )

        await exchange.publish(
            message=message,
            routing_key=get_rabbitmq_routing_key('commit-payloads')
        )
        rest_logger.debug(
            'Published payload against commit ID %s to RabbitMQ payload commit service queue', payload_commit_id
        )

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
        'commitId': payload_commit_id
    }


@app.post('/{projectId:str}/confirmations/callback')
async def register_confirmation_callback(
        request: Request,
        response: Response,
        projectId: str
):
    writer_redis_conn: aioredis.Redis = request.app.writer_redis_pool
    req_json = await request.json()
    await writer_redis_conn.set(f'powerloom:project:{projectId}:callbackURL', req_json['callbackURL'])
    response.status_code = 200
    return {'success': True}


@app.get('/{projectId:str}/payloads')
async def get_payloads(
        request: Request,
        response: Response,
        projectId: str,
        from_height: int = Query(default=1),
        to_height: int = Query(default=-1),
        data: Optional[str] = Query(None)
):

    reader_redis_conn: aioredis.Redis = request.app.reader_redis_pool
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

    if to_height == -1:
        to_height = max_block_height

    if (from_height <= 0) or (to_height > max_block_height) or (from_height > to_height):
        response.status_code = 400
        return {'error': 'Invalid Height'}

    last_pruned_height = await helper_functions.get_last_pruned_height(
        project_id=projectId, reader_redis_conn=reader_redis_conn
    )
    rest_logger.debug('Last pruned height: %s.', last_pruned_height)

    #TODO: Add support to fetch from archived data using dagSegments and traversal logic.
    if (from_height <= last_pruned_height):
        rest_logger.debug("Querying for archived data not yet supported.")
        return {'error': 'The data being queried has been archived. Querying for archived data is not supported.'}

    dag_blocks = await retrieval_utils.fetch_blocks(
        from_height=from_height,
        to_height=to_height,
        project_id=projectId,
        data_flag=data,
        reader_redis_conn=reader_redis_conn,
        ipfs_read_client=request.app.ipfs_read_client
    )
    return dag_blocks


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

    block = await retrieval_utils.retrieve_block_data(
        block_dag_cid=prev_dag_cid,
        project_id=projectId,
        ipfs_read_client=request.app.ipfs_read_client,
        writer_redis_conn=writer_redis_conn,
        data_flag=0
    )

    return {prev_dag_cid: block}


@app.get('/{projectId}/payload/{block_height}/status')
async def get_block_status(
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
    if (block_height <= 0):
        response.status_code = 400
        return {'error': 'Invalid Block Height'}
    """ This endpoint is responsible for retrieving the blockHeight status along with required data """
    max_block_height = await helper_functions.get_block_height(
        project_id=projectId,
        reader_redis_conn=reader_redis_conn
    )
    rest_logger.debug(max_block_height)

    block_status = await retrieval_utils.retrieve_block_status(
        project_id=projectId,
        project_block_height=max_block_height,
        block_height=block_height,
        reader_redis_conn=reader_redis_conn,
        writer_redis_conn=writer_redis_conn,
        ipfs_read_client=request.app.ipfs_read_client
    )

    if block_status is None:
        response.status_code = 404
        return {'error': 'Could not retrieve block status'}
    return block_status.dict()


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

    payload = await retrieval_utils.retrieve_block_data(
        block_dag_cid=prev_dag_cid,
        project_id=projectId,
        ipfs_read_client=request.app.ipfs_read_client,
        writer_redis_conn=writer_redis_conn,
        data_flag=2
    )

    """ Return the payload data """
    return {prev_dag_cid: payload}


@app.post('/registerProjects')
async def register_projects(
        request: Request,
        response: Response,
):
    req_json = await request.json()
    try:
        project_registration_request = ProjectRegistrationRequest.parse_obj(req_json)
    except ValidationError:
        response.status_code = 400
        return {'error': 'Bad request'}


    writer_redis_conn: aioredis.Redis = request.app.writer_redis_pool

    await writer_redis_conn.sadd(
        redis_keys.get_stored_project_ids_key(),
        *project_registration_request.projectIDs
    )

    client = httpx.AsyncClient(limits=httpx.Limits(
        max_connections=20, max_keepalive_connections=20
    ))

    failed_tasks = []
    # Skip summary and stats projectIDs
    projects_for_consensus = filter(lambda project_id: "Summary" not in project_id and "Stats" not in project_id, project_registration_request.projectIDs)

    tasks = []
    for project_id in projects_for_consensus:
        tasks.append(client.post(
            url=settings.consensus_config.service_url + "/registerProjectPeer",
            json=PeerRegistrationRequest(projectID = project_id, instanceID = settings.instance_id).dict()
        ))

    results = await asyncio.gather(*tasks)
    for project_id, r in zip(projects_for_consensus, results):
        if r.status_code != 200:
            failed_tasks.append(project_id)

    if len(failed_tasks) > 0:
        response.status_code = 500
        return {'error': f'Could not register all project peers, failed tasks: {failed_tasks}'}

    return {'success': True}

@app.post('/registerProjectsForIndexing')
async def register_projects_for_indexing(
        request: Request,
        response: Response,
):
    req_json = await request.json()
    try:
        indexing_data = ProjectRegistrationRequestForIndexing.parse_obj(req_json)
    except ValidationError:
        response.status_code = 400
        return {'error': 'Bad request'}


    writer_redis_conn: aioredis.Redis = request.app.writer_redis_pool

    project_ids = dict()
    for project_indexer_data in indexing_data.projects:

        project_ids.update({project_indexer_data.projectID: json.dumps(project_indexer_data.indexerConfig)})

    await writer_redis_conn.hset(redis_keys.get_projects_registered_for_cache_indexing_key_with_namespace(indexing_data.namespace), mapping=project_ids)

    return {'success': True}