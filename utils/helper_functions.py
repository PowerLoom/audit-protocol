from functools import partial, wraps
from utils import redis_keys
from httpx import AsyncClient, Timeout, Limits
from config import settings
from tenacity import AsyncRetrying, stop_after_attempt, wait_random_exponential
import logging
import sys
from redis import asyncio as aioredis

logger = logging.getLogger("Utils|helper_functions")
logger.setLevel(logging.DEBUG)
formatter = logging.Formatter("%(levelname)-8s %(name)-4s %(asctime)s %(msecs)d %(module)s-%(funcName)s: %(message)s")
stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setFormatter(formatter)
stdout_handler.setLevel(logging.DEBUG)
logger.addHandler(stdout_handler)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setFormatter(formatter)
stderr_handler.setLevel(logging.ERROR)
logger.addHandler(stderr_handler)

async def get_tentative_block_height(
        project_id: str,
        reader_redis_conn: aioredis.Redis
) -> int:
    tentative_block_height_key = redis_keys.get_tentative_block_height_key(project_id=project_id)
    out: bytes = await reader_redis_conn.get(tentative_block_height_key)
    if out:
        tentative_block_height = int(out)
    else:
        tentative_block_height = 0
    return tentative_block_height


async def get_last_dag_cid(
        project_id: str,
        reader_redis_conn: aioredis.Redis
) -> str:
    last_dag_cid_key = redis_keys.get_last_dag_cid_key(project_id=project_id)
    out: bytes = await reader_redis_conn.get(last_dag_cid_key)
    if out:
        last_dag_cid = out.decode('utf-8')
    else:
        last_dag_cid = ""

    return last_dag_cid


async def get_dag_cid(
        project_id: str,
        block_height: int,
        reader_redis_conn: aioredis.Redis
):

    dag_cids_key = redis_keys.get_dag_cids_key(project_id)
    out = await reader_redis_conn.zrangebyscore(
        name=dag_cids_key,
        max=block_height,
        min=block_height,
        withscores=False
    )

    if out:
        if isinstance(out, list):
            out = out.pop()
        dag_cid = out.decode('utf-8')
    else:
        dag_cid = None

    return dag_cid


async def get_last_payload_cid(
        project_id: str,
        reader_redis_conn: aioredis.Redis
):
    last_payload_cid_key = redis_keys.get_last_snapshot_cid_key(project_id=project_id)
    out: bytes = await reader_redis_conn.get(last_payload_cid_key)
    if out:
        last_payload_cid = out.decode('utf-8')
    else:
        last_payload_cid = ""

    return last_payload_cid


def cleanup_children_procs(fn):
    @wraps(fn)
    def wrapper(self, *args, **kwargs):
        try:
            fn(self, *args, **kwargs)
            logger.info('Finished running process core...')
        except Exception as e:
            logger.error('Received an exception on process core run(): %s', e, exc_info=True)
            logger.error('Waiting on spawned callback workers to join...\n%s', self._spawned_processes_map)
            for k, v in self._spawned_processes_map.items():
                logger.error('spawned Process Pid to wait on %s', v.pid)
                # internal state reporter might set proc_id_map[k] = -1
                if v != -1:
                    logger.error('Waiting on spawned core worker %s | PID %s  to join...', k, v.pid)
                    v.join()
            logger.error('Finished waiting for all children...now can exit.')
        finally:
            sys.exit(0)
    return wrapper


async def commit_payload(project_id, report_payload, session: AsyncClient, web3_storage_flag=True, skipAnchorProof=False):
    # setting web3Storage flag to true by default
    audit_protocol_url = f'http://{settings.ap_backend.host}:{settings.ap_backend.port}/commit_payload'
    async for attempt in AsyncRetrying(reraise=True, stop=stop_after_attempt(3), wait=wait_random_exponential(multiplier=1, max=30)):
        with attempt:
            response_obj = await session.post(
                    url=audit_protocol_url,
                    json={'payload': report_payload, 'projectId': project_id, 'web3Storage': web3_storage_flag, 'skipAnchorProof': skipAnchorProof}
            )
            logger.debug('Got audit protocol response: %s', response_obj.text)
            response_status_code = response_obj.status_code
            response = response_obj.json()
            if response_status_code in range(200, 300):
                return response
            elif response_status_code == 500 or response_status_code == 502:
                return {
                    "message": f"failed with status code: {response_status_code}", "response": response
                }  # ignore 500 and 502 errors
            elif 'error' in response.keys():
                return{
                    "message": f"failed with error message: {response}"
                }
            else:
                raise Exception(
                    'Failed audit protocol engine call with status code: {} and response: {}'.format(
                        response_status_code, response))


async def get_block_height(
        project_id: str,
        reader_redis_conn,
) -> int:
    block_height_key = redis_keys.get_block_height_key(project_id=project_id)
    out: bytes = await reader_redis_conn.get(block_height_key)
    if out:
        block_height = int(out)
    else:
        block_height = 0

    return block_height


async def get_last_pruned_height(
        project_id: str,
        reader_redis_conn
):
    last_pruned_key = redis_keys.get_pruning_status_key()
    out: bytes = await reader_redis_conn.hget(last_pruned_key, project_id)
    if out:
        last_pruned_height: int = int(out.decode('utf-8'))
    else:
        last_pruned_height: int = 0
    return last_pruned_height


async def check_project_exists(
        project_id: str,
        reader_redis_conn: aioredis.Redis
):
    stored_projects_key = redis_keys.get_stored_project_ids_key()
    out = await reader_redis_conn.sismember(
        name=stored_projects_key,
        value=project_id
    )

    return out

