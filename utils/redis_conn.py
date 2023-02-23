from functools import wraps
from config import settings as settings_conf
from typing import Optional
# from redis import asyncio as aioredis
from redis import asyncio as aioredis
from settings_model import RedisConfig
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


def inject_retry_exception_conf(redis_conf: dict):
    redis_conf.update({"retry_on_error": [redis.exceptions.ReadOnlyError, ]})


REDIS_CONN_CONF = settings_conf.redis.dict()

REDIS_WRITER_CONN_CONF = settings_conf.redis.dict()

REDIS_READER_CONN_CONF = settings_conf.redis_reader.dict()


def construct_redis_url(redis_settings: RedisConfig):
    if redis_settings.password:
        return f'redis://{redis_settings.password}@{redis_settings.host}:{redis_settings.port}/{redis_settings.db}'
    else:
        return f'redis://{redis_settings.host}:{redis_settings.port}/{redis_settings.db}'


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
            inject_retry_exception_conf(redis_conf=REDIS_CONN_CONF)
            connection_pool = redis.BlockingConnectionPool(**REDIS_CONN_CONF, max_connections=20)
            # logging.debug('Created Redis connection Pool')
            with get_redis_conn_from_pool(connection_pool) as redis_obj:
                kwargs[arg_conn] = redis_obj
                logging.debug('Returning after populating redis connection object')
                return fn(*args, **kwargs)
    return wrapper


async def get_redis_pool(redis_settings: RedisConfig = settings_conf.redis, pool_size=200):
    return await aioredis.from_url(
        url=construct_redis_url(redis_settings),
        max_connections=pool_size,
        retry_on_error=[redis.exceptions.ReadOnlyError, ]
    )


# TODO: find references to usage and replace with pool interface
async def get_writer_redis_conn():
    out = await aioredis.Redis(
        host=REDIS_WRITER_CONN_CONF['host'],
        port=REDIS_WRITER_CONN_CONF['port'],
        db=REDIS_WRITER_CONN_CONF['db'],
        password=REDIS_WRITER_CONN_CONF['password'],
        retry_on_error=[redis.exceptions.ReadOnlyError, ],
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
        retry_on_error=[redis.exceptions.ReadOnlyError, ],
        single_connection_client=True
    )
    return out


class RedisPool:
    def __init__(
            self,
            writer_redis_conf: RedisConfig = settings_conf.redis,
            reader_redis_conf: Optional[RedisConfig] = settings_conf.redis_reader,
            pool_size=200,
            replication_mode=True
    ):
        self.reader_redis_pool = None
        self.writer_redis_pool = None
        self._writer_redis_conf = writer_redis_conf
        self._reader_redis_conf = reader_redis_conf
        self._pool_size = pool_size
        self._replication_mode = replication_mode

    async def populate(self):
        if not self.writer_redis_pool:
            self.writer_redis_pool: aioredis.Redis = await get_redis_pool(self._writer_redis_conf, self._pool_size)
            if self._replication_mode:
                self.reader_redis_pool = self.writer_redis_pool
            else:
                if not self.reader_redis_pool:
                    self.reader_redis_pool: aioredis.Redis = await get_redis_pool(self._reader_redis_conf, self._pool_size)
