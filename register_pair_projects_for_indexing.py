from config import settings
from utils import redis_conn
import redis
import os
import json
from redis import Redis


REDIS_CONN_CONF = redis_conn.REDIS_CONN_CONF


####### CHNAGE THIS ##########
NAMESPACE = 'UNISWAPV2'
##############################



## CHANGE INDEX KEY FOR EACH INSTANCE: 
if NAMESPACE == 'UNISWAPV2':
    REDIS_INDEXES_KEY = 'cache:indexesRequested'
else:
    REDIS_INDEXES_KEY = f'cache:indexesRequested:{NAMESPACE}'


# update hashMap
def main():
    if not os.path.exists('static/cached_pair_addresses.json'):
        return
    f = open('static/cached_pair_addresses.json', 'r')
    pairs = json.loads(f.read())

    if len(pairs) <= 0:
        return
    r = Redis(**REDIS_CONN_CONF)
    project_ids = dict()
    for each_pair in pairs:
        addr = each_pair.lower()
        
        project_ids.update({f'uniswap_pairContract_trade_volume_{addr}_{NAMESPACE}': json.dumps({'series': ['24h', '7d']})})
        project_ids.update({f'uniswap_pairContract_pair_total_reserves_{addr}_{NAMESPACE}': json.dumps({'series': ['0']})})
    
    r.hset(REDIS_INDEXES_KEY, mapping=project_ids)


if __name__ == '__main__':
    main()