from utils.redis_conn import get_reader_redis_conn, get_writer_redis_conn
from redis import asyncio as aioredis
import asyncio

async def test():
    random_key = "ASDDS"
    writer_conn: aioredis.Redis = await get_writer_redis_conn()
    reader_conn: aioredis.Redis = await get_reader_redis_conn()

    out = await writer_conn.zadd(
        key=random_key,
        member='a',
        score=10
    )
    out = await writer_conn.zadd(
        key=random_key,
        member='b',
        score=11
    )
    out = await writer_conn.zadd(
        key=random_key,
        member='c',
        score=13
    )

    out = await reader_conn.zrevrange(random_key, 0, -1)
    print(out)

if __name__ == "__main__":
    loop = asyncio.get_event_loop()
    try:
        loop.run_until_complete(test())
    except Exception as e:
        print(e)
        loop.stop()
