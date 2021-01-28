from utils.redis_conn import provide_async_reader_conn_inst, provide_async_writer_conn_inst
from utils import redis_keys
import aioredis
import asyncio


@provide_async_reader_conn_inst
async def test_(project_id, reader_redis_conn: aioredis.Redis):
    out = await reader_redis_conn.sismember(
        key=redis_keys.get_stored_project_ids_key(),
        member=project_id
    )
    print(out)


@provide_async_writer_conn_inst
async def add_project_test(project_id, writer_redis_conn: aioredis.Redis):
    _ = await writer_redis_conn.sadd(
        key=redis_keys.get_stored_project_ids_key(),
        member=project_id
    )


@provide_async_writer_conn_inst
async def rem_project_test(project_id, writer_redis_conn: aioredis.Redis):
    _ = await writer_redis_conn.srem(
        key=redis_keys.get_stored_project_ids_key(),
        member=project_id
    )


loop = asyncio.get_event_loop()
try:
    loop.run_until_complete(asyncio.gather(add_project_test("dummy_init")))
    loop.run_until_complete(asyncio.gather(test_("dummy_init")))
    loop.run_until_complete(asyncio.gather(rem_project_test("dummy_init")))
    loop.run_until_complete(asyncio.gather(test_("dummy_init")))
except Exception as e:
    pass
finally:
    loop.stop()