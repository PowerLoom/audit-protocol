from fastapi import FastAPI, Request, Response
import aioredis
import ipfshttpclient
from config import settings
import logging
from skydb import SkydbTable
import sys
import json
import io

""" Powergate Imports """
from pygate_grpc.client import PowerGateClient

ipfs_client = ipfshttpclient.connect()
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
    app.redis_pool = await aioredis.create_pool(
        address=(REDIS_CONN_CONF['host'], REDIS_CONN_CONF['port']),
        db=REDIS_CONN_CONF['db'],
        password=REDIS_CONN_CONF['password'],
        maxsize=5
    )

@app.post('/')
async def create_dag(
        request: Request,
        response: Response,
    ):
    data = await request.json()
    if 'event_name' in data.keys():
        if data['event_name'] == 'RecordAppended':
            rest_logger.debug(data)
            """ Create a Redis Connection """
            if settings.METADATA_CACHE == "redis":
                rest_logger.debug("Creating a redis connection....")
                redis_conn_raw = await request.app.redis_pool.acquire()
                redis_conn = aioredis.Redis(redis_conn_raw)

            """ Get data from the event """
            txHash = data['txHash']
            timestamp = data['event_data']['timestamp']
            payload_cid = data['event_data']['ipfsCid']

            """ Get required data from redis """
            key = f"TRANSACTION:{txHash}"
            data = await redis_conn.hgetall(key)
            decoded_data = {k.decode('utf-8'):v.decode('utf-8') for k,v in data.items()}
            project_id = int(decoded_data['project_id'])

            """ Get the filecoin token for the project Id """
            KEY = f"filecoinToken:{project_id}"
            token = await redis_conn.get(KEY)
            token = token.decode('utf-8')

            """ Fill up the dag """
            dag = settings.dag_structure
            dag['Height'] = decoded_data['tentative_block_height']
            dag['prevCid'] = decoded_data['prev_dag_cid']
            dag['Data'] = {
                        'Cid': payload_cid,
						'Type': 'HOT_IPFS',
                    }
            dag['TxHash'] = txHash
            dag['Timestamp'] = timestamp
            rest_logger.debug(dag)

            """ Convert dag structure to json and put it on ipfs dag """
            json_string = json.dumps(dag).encode('utf-8')
            data = ipfs_client.dag.put(io.BytesIO(json_string))
            rest_logger.debug(data)
            rest_logger.debug(data['Cid']['/'])

            """ Put the dag json on to filecoin """
            powgate_client = PowerGateClient(settings.POWERGATE_CLIENT_ADDR, False)
            staged_res = powgate_client.data.stage_bytes(json_string, token=token)
            job = powgate_client.config.apply(staged_res.cid, override=False, token=token)

            KEY = f"blockFilecoinStorage:{project_id}:{decoded_data['tentative_block_height']}"
            _ = await redis_conn.hset(
                key=KEY,
                field="blockStageCid",
                value=staged_res.cid,
            )

            _ = await redis_conn.hset(
                key=KEY,
                field="blockDagCid",
                value=data['Cid']['/'],
            )

            _ = await redis_conn.hset(
                key=KEY,
                field="jobId",
                value=job.jobId,
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
                ipfs_table.add_row({'cid': data['Cid']['/']})
            elif settings.METADATA_CACHE == 'redis':
                await redis_conn.set(f'projectID:{project_id}:lastDagCid', data['Cid']['/'])
                await redis_conn.zadd(
                    key=f'projectID:{project_id}:Cids',
                    score=int(decoded_data['tentative_block_height']),
                    member=data['Cid']['/']
                )
                await redis_conn.set(f'projectID:{project_id}:blockHeight', int(decoded_data['tentative_block_height']) + 1)
                request.app.redis_pool.release(redis_conn_raw)

        rest_logger.debug("Latest block added succesfully onto the DAG")

    response.status_code = 200
    return {}

