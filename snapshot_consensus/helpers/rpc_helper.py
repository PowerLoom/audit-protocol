from itertools import repeat
from .message_models import RPCNodesObject
from ..conf import settings
from functools import wraps
import requests
import time
import logging
import logging.handlers
import multiprocessing
import json
import sys

formatter = logging.Formatter(u"%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s")
stdout_handler = logging.StreamHandler(sys.stdout)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setLevel(logging.ERROR)
# stderr_handler.setFormatter(formatter)
service_logger = logging.getLogger(__name__)
service_logger.setLevel(logging.DEBUG)
service_logger.addHandler(stdout_handler)
service_logger.addHandler(stderr_handler)

REQUEST_TIMEOUT = settings.chain.rpc.request_timeout
RETRY_LIMIT = settings.chain.rpc.retry

def auto_retry(tries=3, exc=Exception, delay=5):
    def deco(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            for _ in range(tries):
                try:
                    return func(*args, **kwargs)
                except exc:
                    time.sleep(delay)
                    continue
            raise exc

        return wrapper

    return deco


class BailException(RuntimeError):
    pass


class ConstructRPC:
    def __init__(self, network_id):
        self._network_id = network_id
        self._querystring = {"id": network_id, "jsonrpc": "2.0"}

    def sync_post_json_rpc(self, procedure, rpc_nodes: RPCNodesObject, params=None):
        q_s = self.construct_one_timeRPC(procedure=procedure, params=params)
        rpc_urls = rpc_nodes.NODES
        retry = dict(zip(rpc_urls, repeat(0)))
        success = False
        while True:
            if all(val == rpc_nodes.RETRY_LIMIT for val in retry.values()):
                rpc_logger.error("Retry limit reached for all RPC endpoints. Following request")
                rpc_logger.error("%s", q_s)
                rpc_logger.error("%s", retry)
                raise BailException
            for _url in rpc_urls:
                try:
                    retry[_url] += 1
                    r = requests.post(_url, json=q_s, timeout=REQUEST_TIMEOUT)
                    json_response = r.json()
                except (requests.exceptions.Timeout,
                        requests.exceptions.ConnectionError,
                        requests.exceptions.HTTPError):
                    success = False
                except Exception as e:
                    success = False
                else:
                    if procedure == 'eth_getBlockByNumber' and not json_response['result']:
                        continue
                    success = True
                    return json_response

    def rpc_eth_blocknumber(self, rpc_nodes: RPCNodesObject):
        rpc_response = self.sync_post_json_rpc(procedure="eth_blockNumber", rpc_nodes=rpc_nodes)
        try:
            new_blocknumber = int(rpc_response["result"], 16)
        except Exception as e:
            raise BailException
        else:
            return new_blocknumber

    @auto_retry(tries=RETRY_LIMIT, exc=BailException)
    def rpc_eth_getblock_by_number(self, blocknum, rpc_nodes):
        rpc_response = self.sync_post_json_rpc(procedure="eth_getBlockByNumber", rpc_nodes=rpc_nodes,
                                               params=[hex(blocknum), False])
        try:
            blockdetails = rpc_response["result"]
        except Exception as e:
            raise
        else:
            return blockdetails

    def construct_one_timeRPC(self, procedure, params, defaultBlock=None):
        self._querystring["method"] = procedure
        self._querystring["params"] = []
        if type(params) is list:
            self._querystring["params"].extend(params)
        elif params is not None:
            self._querystring["params"].append(params)
        if defaultBlock is not None:
            self._querystring["params"].append(defaultBlock)
        return self._querystring
