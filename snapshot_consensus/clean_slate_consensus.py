import asyncio
import logging
from utils.redis_conn import RedisPool
from .conf import settings

# Set up logging
logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)
handler = logging.StreamHandler()
formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
handler.setFormatter(formatter)
logger.addHandler(handler)

async def cleanup_redis_state():
    """Delete all redis keys except the ones ending with ':peers'"""
    # Create an instance of the RedisPool class and connect to the Redis instance
    redis_pool = RedisPool(writer_redis_conf=settings.redis)
    await redis_pool.populate()

    # Get the writer Redis pool
    writer_redis_pool = redis_pool.writer_redis_pool

    # Set the pattern for matching keys
    pattern = "*:centralizedConsensus:*"
    except_pattern = b":peers"
    cursor = "0"
    cursor, keys = await writer_redis_pool.scan(cursor=cursor, match=pattern, count=1000)
    for key in keys:
        if not key.endswith(except_pattern):
            await writer_redis_pool.delete(key)
            logger.info("Deleted key: %s", key)

if __name__ == '__main__':
    # Run the cleanup function in an asyncio event loop
    loop = asyncio.get_event_loop()
    loop.run_until_complete(cleanup_redis_state())
