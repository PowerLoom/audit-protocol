import aioredis
from functools import wraps


def setup_teardown_boilerplate(fn):
    @wraps(fn)
    async def wrapped(*args, **kwargs):
        arg_conn = 'redis_conn'
        redis_conn_raw = await kwargs['request'].app.redis_pool.acquire()
        redis_conn = aioredis.Redis(redis_conn_raw)
        kwargs[arg_conn] = redis_conn
        try:
            return await fn(*args, **kwargs)
        except Exception as e:
            print(e)
            return {'error': 'Internal Server Error'}
        finally:
            kwargs['request'].app.redis_pool.release(redis_conn_raw)
    return wrapped