from config import settings
from utils.redis_conn import provide_redis_conn
import redis

NAMESPACE = 'UNISWAPV2-prod-1'


@provide_redis_conn
def main(redis_conn: redis.Redis=None):
    cached_sliding_window_data = 'uniswap:pairContract:'+NAMESPACE+':*:slidingWindowData'
    cached_keys_list = set()
    for k in redis_conn.scan_iter(match=cached_sliding_window_data, count=1000):
        cached_keys_list.add(k)
    try:
        _ = redis_conn.delete(*cached_keys_list)
    except:
        pass
    else:
        print(f'Deleted {_} cached sliding window states')
    primary_indexes_head = f'*uniswap_pairContract_trade_volume*{NAMESPACE}*slidingCache*head'
    primary_indexes_tail = f'*uniswap_pairContract_trade_volume*{NAMESPACE}*slidingCache*tail'
    cached_indexes_tail_list = set()
    cached_indexes_head_list = set()
    try:
        for k in redis_conn.scan_iter(match=primary_indexes_tail, count=1000):
            cached_indexes_tail_list.add(k)
        _ = redis_conn.delete(*cached_indexes_tail_list)
    except:
        pass
    else:
        print(f'Deleted {_} cached tail indexes')
    for k in redis_conn.scan_iter(match=primary_indexes_head, count=1000):
        cached_indexes_head_list.add(k)
    try:
        _ = redis_conn.delete(*cached_indexes_head_list)
    except:
        pass
    else:
        print(f'Deleted {_} cached head indexes')

if __name__ == '__main__':
    main()

