import redis_keys
import aioredis


async def get_tentative_block_height(
        project_id: str,
        reader_redis_conn
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
        reader_redis_conn
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
        key=dag_cids_key,
        max=block_height,
        min=block_height,
        withscores=False
    )

    if out:
        if isinstance(out, list):
            out = out.pop()
        dag_cid = out.decode('utf-8')
    else:
        dag_cid = ""

    return dag_cid

async def get_last_payload_cid(
        project_id: str,
        reader_redis_conn
):
    last_payload_cid_key = redis_keys.get_last_snapshot_cid_key(project_id=project_id)
    out: bytes = await reader_redis_conn.get(last_payload_cid_key)
    if out:
        last_payload_cid = out.decode('utf-8')
    else:
        last_payload_cid = ""

    return last_payload_cid


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
        reader_redis_conn,
        writer_redis_conn
):
    last_pruned_key = redis_keys.get_last_pruned_key(project_id=project_id)
    out: bytes = await reader_redis_conn.get(last_pruned_key)
    if out:
        last_pruned_height: int = int(out.decode('utf-8'))
    else:
        _ = await writer_redis_conn.set(last_pruned_key, 0)
        last_pruned_height: int = 0
    return last_pruned_height
