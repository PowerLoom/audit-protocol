import asyncio
import logging
from utils.redis_conn import RedisPool
from .helpers.redis_keys import get_epoch_generator_last_epoch, get_epoch_generator_epoch_history
from .conf import settings

# Set up logging
logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)
handler = logging.StreamHandler()
formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
handler.setFormatter(formatter)
logger.addHandler(handler)

async def cleanup_redis_state():
    # Create an instance of the RedisPool class and connect to the Redis instance
    redis_pool = RedisPool(writer_redis_conf=settings.redis)
    await redis_pool.populate()

    # Get the writer Redis pool
    writer_redis_pool = redis_pool.writer_redis_pool

    # Delete the keys used by the code
    await writer_redis_pool.delete(get_epoch_generator_last_epoch())
    logger.info("Deleted key: %s", get_epoch_generator_last_epoch())
    await writer_redis_pool.delete(get_epoch_generator_epoch_history())
    logger.info("Deleted key: %s", get_epoch_generator_epoch_history())

if __name__ == '__main__':
    # Run the cleanup function in an asyncio event loop
    loop = asyncio.get_event_loop()
    loop.run_until_complete(cleanup_redis_state())
