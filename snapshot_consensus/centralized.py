from .data_models import (
    SnapshotSubmission, SubmissionResponse, PeerRegistrationRequest, SubmissionAcceptanceStatus, SnapshotBase,
    EpochConsensusStatus
)
from .helpers.state import submission_delayed, register_submission, check_submissions_consensus
from .helpers.redis_keys import *
from typing import Optional
from fastapi import FastAPI, Request, Response, Query
from fastapi.middleware.cors import CORSMiddleware
from uuid import uuid4
from utils.redis_conn import RedisPool
from utils.rabbitmq_utils import get_rabbitmq_connection, get_rabbitmq_channel, get_rabbitmq_core_exchange, get_rabbitmq_routing_key
from functools import partial
from pydantic import ValidationError
from aio_pika import ExchangeType, DeliveryMode, Message
from aio_pika.pool import Pool
import logging
import sys
import json
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
service_logger = logging.getLogger(__name__)
service_logger.setLevel(logging.DEBUG)
service_logger.addHandler(stdout_handler)
service_logger.addHandler(stderr_handler)

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


@app.post('/registerProjectPeer')
async def register_peer_against_project(
        request: Request,
        response: Response
):
    req_json = await request.json()
    try:
        req_parsed: PeerRegistrationRequest = PeerRegistrationRequest.parse_obj(req_json)
    except ValidationError:
        response.status_code = 400
        return {}
    await request.app.writer_redis_pool.sadd(
        get_project_registered_peers_set_key(req_parsed.projectID),
        req_parsed.instanceID
    )


@app.post('/submitSnapshot')
async def submit_snapshot(
        request: Request,
        response: Response
):
    cur_ts = int(time.time())
    req_json = await request.json()
    try:
        req_parsed = SnapshotSubmission.parse_obj(req_json)
    except ValidationError:
        response.status_code = 400
        return {}
    # get last accepted epoch?
    if submission_delayed(
        project_id=req_parsed.projectID,
        epoch_end=req_parsed.epoch.end,
        redis_conn=request.app.writer_redis_pool
    ):
        response_obj = SubmissionResponse(status=SubmissionAcceptanceStatus.accepted, delayedSubmission=True)
    else:
        response_obj = SubmissionResponse(status=SubmissionAcceptanceStatus.accepted, delayedSubmission=False)
    consensus_status, finalized_cid = await register_submission(req_parsed, cur_ts, request.app.writer_redis_pool)
    response_obj.status = consensus_status
    response_obj.finalizedSnapshotCID = finalized_cid
    return response_obj.dict()


@app.post('/checkForSnapshotConfirmation')
async def check_submission_status(
        request: Request,
        response: Response
):
    req_json = await request.json()
    try:
        req_parsed = SnapshotSubmission.parse_obj(req_json)
    except ValidationError:
        response.status_code = 400
        return {}
    status, finalized_cid = await check_submissions_consensus(
        submission=req_parsed, redis_conn=request.app.writer_redis_pool
    )
    if status == SubmissionAcceptanceStatus.notsubmitted:
        response.status_code = 400
        return SubmissionResponse(status=status, delayedSubmission=False, finalizedSnapshotCID=None).dict()
    else:
        return SubmissionResponse(
            status=status,
            delayedSubmission=await submission_delayed(
                req_parsed.projectID, epoch_end=req_parsed.epoch.end, redis_conn=request.app.writer_redis_pool
            ),
            finalizedSnapshotCID=finalized_cid
        ).dict()


@app.get('/epochStatus')
async def epoch_status(
        request: Request,
        response: Response
):
    req_json = await request.json()
    try:
        req_parsed = SnapshotBase.parse_obj(req_json)
    except ValidationError:
        response.status_code = 400
        return {}
    status, finalized_cid = await check_submissions_consensus(
        submission=req_parsed, redis_conn=request.app.writer_redis_pool
    )
    if status != SubmissionAcceptanceStatus.finalized:
        status = EpochConsensusStatus.no_consensus
    else:
        status = EpochConsensusStatus.consensus_achieved
    return SubmissionResponse(status=status, delayedSubmission=False, finalizedSnapshotCID=finalized_cid).dict()
