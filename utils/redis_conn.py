import aioredis
from functools import wraps
import traceback
from config import settings as settings_conf
import asyncio
import logging
logger = logging.getLogger(__name__)
logger.setLevel(level="DEBUG")

REDIS_CONN_CONF = {
    "host": settings_conf.redis.host,
    "port": settings_conf.redis.port,
    "password": settings_conf.redis.password,
    "db": settings_conf.redis.db
}

REDIS_WRITER_CONN_CONF = {
    "host": settings_conf.redis.host,
    "port": settings_conf.redis.port,
    "password": settings_conf.redis.password,
    "db": settings_conf.redis.db
}

REDIS_READER_CONN_CONF = {
    "host": settings_conf.redis_reader.host,
    "port": settings_conf.redis_reader.port,
    "password": settings_conf.redis_reader.password,
    "db": settings_conf.redis_reader.db
}


async def get_writer_redis_pool():
    out = await aioredis.create_pool(
        address=(REDIS_WRITER_CONN_CONF['host'], REDIS_WRITER_CONN_CONF['port']),
        db=REDIS_WRITER_CONN_CONF['db'],
        password=REDIS_WRITER_CONN_CONF['password'],
        maxsize=50
    )
    return out


async def get_reader_redis_pool():
    out = await aioredis.create_pool(
        address=(REDIS_READER_CONN_CONF['host'], REDIS_READER_CONN_CONF['port']),
        db=REDIS_READER_CONN_CONF['db'],
        password=REDIS_READER_CONN_CONF['password'],
        maxsize=50
    )
    return out


async def get_writer_redis_conn():
    out = await aioredis.create_redis(
        address=(REDIS_WRITER_CONN_CONF['host'], REDIS_WRITER_CONN_CONF['port']),
        db=REDIS_WRITER_CONN_CONF['db'],
        password=REDIS_WRITER_CONN_CONF['password'],
    )
    return out


async def get_reader_redis_conn():
    out = await aioredis.create_redis(
        address=(REDIS_READER_CONN_CONF['host'], REDIS_READER_CONN_CONF['port']),
        db=REDIS_READER_CONN_CONF['db'],
        password=REDIS_READER_CONN_CONF['password'],
    )
    return out


class RedisPool:

    def __init__(self):
        self.reader_redis_pool = asyncio.run(get_reader_redis_pool())
        self.writer_redis_pool = asyncio.run(get_writer_redis_pool())


def setup_teardown_boilerplate(fn):
    @wraps(fn)
    async def wrapped(*args, **kwargs):
        arg_conn = 'redis_conn'
        # func_params = fn.__code__.co_varnames
        redis_conn_raw = await kwargs['request'].app.redis_pool.acquire()
        redis_conn = aioredis.Redis(redis_conn_raw)
        # conn_in_args = arg_conn in func_params and func_params.index(arg_conn) < len(args)
        # conn_in_kwargs = arg_conn in kwargs
        kwargs[arg_conn] = redis_conn
        try:
            return await fn(*args, **kwargs)
        except Exception as e:
            traceback.print_exception(type(e),e,e.__traceback__)
            return {'error': 'Internal Server Error'}
        finally:
            kwargs['request'].app.redis_pool.release(redis_conn_raw)
    return wrapped


def inject_redis_conn(fn):
    @wraps(fn)
    async def redis_conn_wrapper(*args, **kwargs):
        arg_name = 'redis_conn'
        redis_conn = await aioredis.create_redis_pool((REDIS_CONN_CONF['host'],REDIS_CONN_CONF['port']))
        kwargs[arg_name] = redis_conn
        return await fn(*args, **kwargs)
    return redis_conn_wrapper


def inject_reader_redis_conn(fn):
    @wraps(fn)
    async def reader_redis_wrapper(*args, **kwargs):
        arg_name = 'reader_redis_conn'
        reader_redis_conn_raw = await kwargs['request'].app.reader_redis_pool.acquire()
        reader_redis_conn = aioredis.Redis(reader_redis_conn_raw)
        kwargs[arg_name] = reader_redis_conn
        try:
            return await fn(*args, **kwargs)
        except Exception as e:
            logger.error(e, exc_info=True)
            return {'error': 'Internal Server Error'}
        finally:
            kwargs['request'].app.reader_redis_pool.release(reader_redis_conn_raw)
    return reader_redis_wrapper


def inject_writer_redis_conn(fn):
    @wraps(fn)
    async def writer_redis_wrapper(*args, **kwargs):
        arg_name = 'writer_redis_conn'
        writer_redis_conn_raw = await kwargs['request'].app.writer_redis_pool.acquire()
        writer_redis_conn = aioredis.Redis(writer_redis_conn_raw)
        kwargs[arg_name] = writer_redis_conn

        try:
            return await fn(*args, **kwargs)
        except Exception as e:
            logger.error(e, exc_info=True)
            return {'error': 'Internal Server Error'}
        finally:
            kwargs['request'].app.writer_redis_pool.release(writer_redis_conn_raw)
    return writer_redis_wrapper


def provide_async_reader_conn_inst(fn):
    @wraps(fn)
    async def async_redis_reader_wrapper(*args, **kwargs):
        arg_name = 'reader_redis_conn'
        reader_redis_conn = await get_reader_redis_conn()
        kwargs[arg_name] = reader_redis_conn
        try:
            return await fn(*args, **kwargs)
        except Exception as e:
            logger.error(e, exc_info=True)
            return {'error': 'Internal Server Error'}
        finally:
            reader_redis_conn.close()
            await reader_redis_conn.wait_closed()
    return async_redis_reader_wrapper


def provide_async_writer_conn_inst(fn):
    @wraps(fn)
    async def async_redis_writer_wrapper(*args, **kwargs):
        arg_name = 'writer_redis_conn'
        writer_redis_conn = await get_writer_redis_conn()
        kwargs[arg_name] = writer_redis_conn
        try:
            return await fn(*args, **kwargs)
        except Exception as e:
            logger.error(e, exc_info=True)
            return {'error': 'Internal Server Error'}
        finally:
            writer_redis_conn.close()
            await writer_redis_conn.wait_closed()
    return async_redis_writer_wrapper
