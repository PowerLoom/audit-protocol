import aioredis
from functools import wraps


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
        except:
            return {'error': 'Internal Server Error'}
        finally:
            kwargs['request'].app.redis_pool.release(redis_conn_raw)
    return wrapped
