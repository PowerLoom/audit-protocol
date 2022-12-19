from .epoch_generator import main
import asyncio
import logging
from utils.redis_conn import RedisPool
from .helpers.redis_keys import get_epoch_generator_last_epoch, get_epoch_generator_epoch_history
from .conf import settings
from .epoch_generator import EpochGenerator

# Set up logging
logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)
handler = logging.StreamHandler()
formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
handler.setFormatter(formatter)
logger.addHandler(handler)

async def cleanup_redis(redis_conn: RedisPool):
    # Delete the keys used by the code
    await redis_conn.delete(get_epoch_generator_last_epoch())
    logger.info("Deleted key: %s", get_epoch_generator_last_epoch())
    await redis_conn.delete(get_epoch_generator_epoch_history())
    logger.info("Deleted key: %s", get_epoch_generator_epoch_history())

async def test_epoch_generation():
    # Create an instance of the RedisPool class and connect to the Redis instance
    redis_pool = RedisPool(writer_redis_conf=settings.test_redis)
    await redis_pool.populate()
    writer_redis_pool = redis_pool.writer_redis_pool
    reader_redis_pool = redis_pool.reader_redis_pool
    
    await cleanup_redis(writer_redis_pool)
    ticker_process = EpochGenerator(name="PowerLoom|EpochTracker|LinearTest1", simulation_mode=True)
    kwargs = dict()
    await ticker_process.setup(**kwargs)
    await ticker_process.init(begin_block_epoch=16216611)

    # Get the current epoch end block height from Redis
    epoch_end_block_height = await reader_redis_pool.get(get_epoch_generator_last_epoch())
    if not(epoch_end_block_height):
        assert False, "Epoch end block height not found in Redis"
    epoch_end_block_height = int(epoch_end_block_height.decode("utf-8"))
    logger.info("Current epoch end block height: %s", epoch_end_block_height)
    assert epoch_end_block_height == 16216700, "Epoch end block height is not correct"

async def test_state_resume():
    # Create an instance of the RedisPool class and connect to the Redis instance
    redis_pool = RedisPool(writer_redis_conf=settings.test_redis)
    await redis_pool.populate()
    writer_redis_pool = redis_pool.writer_redis_pool
    reader_redis_pool = redis_pool.reader_redis_pool
    
    ticker_process = EpochGenerator(name="PowerLoom|EpochTracker|LinearTest2", simulation_mode=True)
    kwargs = dict()
    await ticker_process.setup(**kwargs)
    await ticker_process.init()

    # Get the current epoch end block height from Redis
    epoch_end_block_height = await reader_redis_pool.get(get_epoch_generator_last_epoch())
    if not(epoch_end_block_height):
        assert False, "Epoch end block height not found in Redis"
    epoch_end_block_height = int(epoch_end_block_height.decode("utf-8"))
    logger.info("Current epoch end block height: %s", epoch_end_block_height)
    assert epoch_end_block_height == 16216790, "Epoch end block height is not correct"


async def test_force_restart_with_redis_state_present():
    # Create an instance of the RedisPool class and connect to the Redis instance
    redis_pool = RedisPool(writer_redis_conf=settings.test_redis)
    await redis_pool.populate()
    writer_redis_pool = redis_pool.writer_redis_pool
    reader_redis_pool = redis_pool.reader_redis_pool
    
    ticker_process = EpochGenerator(name="PowerLoom|EpochTracker|LinearTest3", simulation_mode=True)
    kwargs = dict()
    await ticker_process.setup(**kwargs)
    # Should not do anything
    await ticker_process.init(begin_block_epoch=16216611)

    # Get the current epoch end block height from Redis
    epoch_end_block_height = await reader_redis_pool.get(get_epoch_generator_last_epoch())
    if not(epoch_end_block_height):
        assert False, "Epoch end block height not found in Redis"
    epoch_end_block_height = int(epoch_end_block_height.decode("utf-8"))
    logger.info("Current epoch end block height: %s", epoch_end_block_height)
    assert epoch_end_block_height == 16216790, "Epoch end block height is not correct"
    ## Cleanup redis
    # await cleanup_redis(writer_redis_pool)


if __name__ == '__main__':
    # Run the cleanup function in an asyncio event loop
    loop = asyncio.get_event_loop()
    loop.run_until_complete(test_epoch_generation())
    loop.run_until_complete(test_state_resume())
    loop.run_until_complete(test_force_restart_with_redis_state_present())
