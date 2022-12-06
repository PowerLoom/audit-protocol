from config import settings
from utils import redis_conn
from urllib.parse import urljoin
from snapshot_consensus.data_models import PeerRegistrationRequest
import httpx
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
    client = httpx.Client(limits=httpx.Limits(
        max_connections=20, max_keepalive_connections=20
    ))
    for each_project in project_ids.keys():
        r = client.post(
            url=urljoin(base=settings.consensus_config.service_url, url='/registerProjectPeer'),
            json=PeerRegistrationRequest(projectID=each_project, instanceID=settings.instance_id).dict()
        )
        if r.status_code == 200:
            print(
                f'Registered project {each_project} for consensus service snapshot submission | '
                f'Status code: {r.status_code} | Response: {r.text}')
        else:
            print(
                f'NOT Registered project {each_project} for consensus service snapshot submission | '
                f'Status code: {r.status_code} | Response: {r.text}')

if __name__ == '__main__':
    main()
