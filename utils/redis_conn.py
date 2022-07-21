from functools import wraps
from config import settings as settings_conf
# import aioredis
from redis import asyncio as aioredis
import traceback
import redis
import contextlib
import redis.exceptions as redis_exc
import logging
import sys


logger = logging.getLogger(__name__)
logger.setLevel(level=logging.DEBUG)
formatter = logging.Formatter(u"%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s")

stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setLevel(logging.DEBUG)
stdout_handler.setFormatter(formatter)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setLevel(logging.ERROR)
stderr_handler.setFormatter(formatter)

logger.addHandler(stdout_handler)
logger.addHandler(stderr_handler)

REDIS_CONN_CONF = {
    "host": settings_conf.redis.host,
    "port": settings_conf.redis.port,
    "password": settings_conf.redis.password,
    "db": settings_conf.redis.db,
    "retry_on_error": redis.exceptions.ReadOnlyError
}

REDIS_WRITER_CONN_CONF = {
    "host": settings_conf.redis.host,
    "port": settings_conf.redis.port,
    "password": settings_conf.redis.password,
    "db": settings_conf.redis.db,
    "retry_on_error": redis.exceptions.ReadOnlyError
}

REDIS_READER_CONN_CONF = {
    "host": settings_conf.redis_reader.host,
    "port": settings_conf.redis_reader.port,
    "password": settings_conf.redis_reader.password,
    "db": settings_conf.redis_reader.db,
    "retry_on_error": redis.exceptions.ReadOnlyError
}


def construct_writer_redis_url():
    if REDIS_WRITER_CONN_CONF["password"]:
        return f'redis://{REDIS_WRITER_CONN_CONF["password"]}@{REDIS_WRITER_CONN_CONF["host"]}:{REDIS_WRITER_CONN_CONF["port"]}'\
               f'/{REDIS_WRITER_CONN_CONF["db"]}'
    else:
        return f'redis://{REDIS_WRITER_CONN_CONF["host"]}:{REDIS_WRITER_CONN_CONF["port"]}/{REDIS_WRITER_CONN_CONF["db"]}'


def construct_reader_redis_url():
    if REDIS_READER_CONN_CONF["password"]:
        return f'redis://{REDIS_READER_CONN_CONF["password"]}@{REDIS_READER_CONN_CONF["host"]}:{REDIS_READER_CONN_CONF["port"]}'\
               f'/{REDIS_READER_CONN_CONF["db"]}'
    else:
        return f'redis://{REDIS_READER_CONN_CONF["host"]}:{REDIS_READER_CONN_CONF["port"]}/{REDIS_READER_CONN_CONF["db"]}'


@contextlib.contextmanager
def get_redis_conn_from_pool(connection_pool: redis.BlockingConnectionPool) -> redis.Redis:
    """
    Contextmanager that will create and teardown a session.
    """
    try:
        redis_conn = redis.Redis(connection_pool=connection_pool)
        yield redis_conn
    except redis_exc.RedisError:
        raise
    except KeyboardInterrupt:
        pass


def provide_redis_conn(fn):
    @wraps(fn)
    def wrapper(*args, **kwargs):
        arg_conn = 'redis_conn'
        func_params = fn.__code__.co_varnames
        conn_in_args = arg_conn in func_params and func_params.index(arg_conn) < len(args)
        conn_in_kwargs = arg_conn in kwargs
        if conn_in_args or conn_in_kwargs:
            # logging.debug('Found redis_conn populated already in %s', fn.__name__)
            return fn(*args, **kwargs)
        else:
            # logging.debug('Found redis_conn not populated in %s', fn.__name__)
            connection_pool = redis.BlockingConnectionPool(**REDIS_CONN_CONF, max_connections=20)
            # logging.debug('Created Redis connection Pool')
            with get_redis_conn_from_pool(connection_pool) as redis_obj:
                kwargs[arg_conn] = redis_obj
                logging.debug('Returning after populating redis connection object')
                return fn(*args, **kwargs)
    return wrapper


async def get_writer_redis_pool(pool_size=200):
    return await aioredis.from_url(
        url=construct_writer_redis_url(),
        max_connections=pool_size,
        retry_on_error=redis.exceptions.ReadOnlyError
    )


async def get_reader_redis_pool(pool_size=200):
    return await aioredis.from_url(
        url=construct_reader_redis_url(),
        max_connections=pool_size,
        retry_on_error=redis.exceptions.ReadOnlyError
    )


# TODO: find references to usage and replace with pool interface
async def get_writer_redis_conn():
    out = await aioredis.Redis(
        host=REDIS_WRITER_CONN_CONF['host'],
        port=REDIS_WRITER_CONN_CONF['port'],
        db=REDIS_WRITER_CONN_CONF['db'],
        password=REDIS_WRITER_CONN_CONF['password'],
        retry_on_error=redis.exceptions.ReadOnlyError,
        single_connection_client=True
    )
    return out

# TODO: find references to usage and replace with pool interface
async def get_reader_redis_conn():
    out = await aioredis.Redis(
        host=REDIS_READER_CONN_CONF['host'],
        port=REDIS_READER_CONN_CONF['port'],
        db=REDIS_READER_CONN_CONF['db'],
        password=REDIS_READER_CONN_CONF['password'],
        retry_on_error=redis.exceptions.ReadOnlyError,
        single_connection_client=True
    )
    return out


class RedisPool:
    def __init__(self, pool_size=200, replication_mode=False):
        self.reader_redis_pool = None
        self.writer_redis_pool = None
        self._pool_size = pool_size
        self._replication_mode = replication_mode

    async def populate(self):
        if not self.writer_redis_pool:
            self.writer_redis_pool: aioredis.Redis = await get_writer_redis_pool(self._pool_size)
            if not self._replication_mode:
                self.reader_redis_pool = self.writer_redis_pool
            else:
                if not self.reader_redis_pool:
                    self.reader_redis_pool: aioredis.Redis = await get_reader_redis_pool(self._pool_size)
