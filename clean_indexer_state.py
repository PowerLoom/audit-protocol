from config import settings
from utils.redis_conn import provide_redis_conn
import redis

NAMESPACE = settings.pooler_namespace


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
    cached_each_contract_v2_pair_data = f'*{NAMESPACE}*contractV2PairCachedData*'
    recent_logs = f'*{NAMESPACE}*recentLogs*'
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
    cached_v2_pair_data_keys = set()
    try:
        for k in redis_conn.scan_iter(match=cached_each_contract_v2_pair_data, count=1000):
            cached_v2_pair_data_keys.add(k)
        _ = redis_conn.delete(*cached_v2_pair_data_keys)
    except:
        pass
    else:
        print(f'Deleted {_} cached v2 pair data')
    recent_logs_keys = set()
    try:
        for k in redis_conn.scan_iter(match=recent_logs, count=1000):
            recent_logs_keys.add(k)
        _ = redis_conn.delete(*recent_logs_keys)
    except:
        pass
    else:
        print(f'Deleted {_} recent logs')

if __name__ == '__main__':
    main()

