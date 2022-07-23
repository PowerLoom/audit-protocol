from typing import Union, Tuple, Iterable
from psutil import net_io_counters
from web3 import Web3
from functools import partial
from time import time
from httpx import AsyncClient, Timeout, Limits
from config import settings
from data_models import liquidityProcessedData, uniswapPairsSnapshotZset, uniswapPairSummary7dCidRange, uniswapPairSummary24hCidRange, uniswapPairSummaryCid24hResultant, uniswapPairSummaryCid7dResultant
from utils import helper_functions
from utils import redis_keys
from utils import retrieval_utils
from utils.ipfs_async import client as ipfs_client
from utils.redis_conn import RedisPool, provide_redis_conn
from utils.retrieval_utils import retrieve_block_data
from redis import asyncio as aioredis
import cardinality
import asyncio
import json
import logging.config
import os
import sys
from gnosis.eth import EthereumClient

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
formatter = logging.Formatter("%(levelname)-8s %(name)-4s %(asctime)s %(msecs)d %(module)s-%(funcName)s: %(message)s")
stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setFormatter(formatter)
stdout_handler.setLevel(logging.DEBUG)
logger.addHandler(stdout_handler)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setFormatter(formatter)
stderr_handler.setLevel(logging.ERROR)
logger.addHandler(stderr_handler)
logger.debug("Initialized logger")

NAMESPACE = 'UNISWAPV2'


ethereum_client = EthereumClient(settings.rpc_url)
# TODO: Use async http provider once it is considered stable by the web3.py project maintainers
# web3_async = Web3(Web3.AsyncHTTPProvider(settings.RPC.MATIC[0]))


async def get_dag_block_by_height(project_id, block_height, reader_redis_conn: aioredis.Redis):
    dag_block = {}
    try:
        dag_cid = await helper_functions.get_dag_cid(
            project_id=project_id,
            block_height=block_height,
            reader_redis_conn=reader_redis_conn
        )
        dag_block = await retrieve_block_data(block_dag_cid=dag_cid, data_flag=1)
        dag_block["dagCid"] = dag_cid
    except Exception as e:
        logger.error(f"Error: can't get dag block with msg: {str(e)} | projectId:{project_id}, block_height:{block_height}", exc_info=True)
        dag_block = {}

    return dag_block


async def get_dag_blocks_in_range(project_id, from_block, to_block, reader_redis_conn: aioredis.Redis):
    dag_chain = []
    for i in range(from_block, to_block + 1):
        t = await get_dag_block_by_height(project_id, i, reader_redis_conn)
        dag_chain.append(t)

    dag_chain.reverse()

    return dag_chain


def read_json_file(file_path: str):
    """Read given json file and return its content as a dictionary."""
    try:
        f_ = open(file_path, 'r')
    except Exception as e:
        logger.warning(f"Unable to open the {file_path} file")
        logger.error(e, exc_info=True)
        raise e
    else:
        json_data = json.loads(f_.read())
    return json_data


pair_contract_abi = read_json_file(f"abis/UniswapV2Pair.json")
erc20_abi = read_json_file('abis/IERC20.json')

def get_maker_pair_data(prop):
    prop = prop.lower()
    if prop.lower() == "name":
        return "Maker"
    elif prop.lower() == "symbol":
        return "MKR"
    else:
        return "Maker"

async def get_oldest_block_and_timestamp(pair_contract_address):
    project_id_token_reserve = f'uniswap_pairContract_pair_total_reserves_{pair_contract_address}_{NAMESPACE}'

    aioredis_pool = RedisPool()
    await aioredis_pool.populate()
    redis_conn: aioredis.Redis = aioredis_pool.writer_redis_pool

    # liquidty data
    liquidity_tail_marker_24h, liquidity_tail_marker_7d = await redis_conn.mget(
        redis_keys.get_sliding_window_cache_tail_marker(project_id_token_reserve, '24h'),
        redis_keys.get_sliding_window_cache_tail_marker(project_id_token_reserve, '7d')
    )
    liquidity_tail_marker_24h = int(liquidity_tail_marker_24h.decode('utf-8')) if liquidity_tail_marker_24h else 0
    liquidity_tail_marker_7d = int(liquidity_tail_marker_7d.decode('utf-8')) if liquidity_tail_marker_7d else 0


    liquidity_data_24h, liquidity_data_7d = await asyncio.gather(
        get_dag_block_by_height(
            project_id_token_reserve,
            liquidity_tail_marker_24h,
            redis_conn
        ),
        get_dag_block_by_height(
            project_id_token_reserve,
            liquidity_tail_marker_7d,
            redis_conn
        )
    )

    if not liquidity_data_24h or not liquidity_data_7d:
        return {
            "begin_block_height_24h": 0,
            "begin_block_timestamp_24h": 0,
            "begin_block_height_7d": 0,
            "begin_block_timestamp_7d": 0,
        }


    return {
        "begin_block_height_24h": int(liquidity_data_24h['data']['payload']['chainHeightRange']['begin']),
        "begin_block_timestamp_24h": int(liquidity_data_24h['data']['payload']['timestamp']),
        "begin_block_height_7d": int(liquidity_data_7d['data']['payload']['chainHeightRange']['begin']),
        "begin_block_timestamp_7d": int(liquidity_data_7d['data']['payload']['timestamp']),
    }

async def get_pair_tokens_metadata(
    pair_address, redis_conn: aioredis.Redis
):
    """
        returns information on the tokens contained within a pair contract - name, symbol, decimals of token0 and token1
        also returns pair symbol by concatenating {token0Symbol}-{token1Symbol}
    """
    try:
        pair_address = Web3.toChecksumAddress(pair_address)

        # check if cache exist
        pair_tokens_metadata_cache = await redis_conn.hgetall(redis_keys.get_uniswap_pair_contract_tokens_data(pair_address))

        # parse addresses cache or call eth rpc
        token0Addr = None
        token1Addr = None
        if pair_tokens_metadata_cache:
            token0Addr = Web3.toChecksumAddress(pair_tokens_metadata_cache[b"token0Addr"].decode('utf-8'))
            token1Addr = Web3.toChecksumAddress(pair_tokens_metadata_cache[b"token1Addr"].decode('utf-8'))
            token0_decimals = pair_tokens_metadata_cache[b"token0_decimals"].decode('utf-8')
            token1_decimals = pair_tokens_metadata_cache[b"token1_decimals"].decode('utf-8')
            token0_symbol = pair_tokens_metadata_cache[b"token0_symbol"].decode('utf-8')
            token1_symbol = pair_tokens_metadata_cache[b"token1_symbol"].decode('utf-8')
            token0_name = pair_tokens_metadata_cache[b"token0_name"].decode('utf-8')
            token1_name = pair_tokens_metadata_cache[b"token1_name"].decode('utf-8')
        else:
            # get token0 and token1 addresses
            pair_contract_obj = ethereum_client.w3.eth.contract(
                address=Web3.toChecksumAddress(pair_address),
                abi=pair_contract_abi
            )
            token0Addr, token1Addr = ethereum_client.batch_call([
                pair_contract_obj.functions.token0(),
                pair_contract_obj.functions.token1()
            ])


            # get token0 and token1 metadata:
            token0 = ethereum_client.w3.eth.contract(
                address=Web3.toChecksumAddress(token0Addr),
                abi=erc20_abi
            )
            token1 = ethereum_client.w3.eth.contract(
                address=Web3.toChecksumAddress(token1Addr),
                abi=erc20_abi
            )

            tasks = list()

            #special case to handle maker token
            maker_token0 = None
            maker_token1 = None
            if(Web3.toChecksumAddress(settings.contract_addresses.MAKER) == Web3.toChecksumAddress(token0Addr)):
                token0_name = get_maker_pair_data('name')
                token0_symbol = get_maker_pair_data('symbol')
                maker_token0 = True
            else:
                tasks.append(token0.functions.name())
                tasks.append(token0.functions.symbol())
            tasks.append(token0.functions.decimals())


            if(Web3.toChecksumAddress(settings.contract_addresses.MAKER) == Web3.toChecksumAddress(token1Addr)):
                token1_name = get_maker_pair_data('name')
                token1_symbol = get_maker_pair_data('symbol')
                maker_token1 = True
            else:
                tasks.append(token1.functions.name())
                tasks.append(token1.functions.symbol())
            tasks.append(token1.functions.decimals())

            if maker_token1:
                [token0_name, token0_symbol, token0_decimals, token1_decimals] = ethereum_client.batch_call(
                    tasks
                )
            elif maker_token0:
                [token0_decimals, token1_name, token1_symbol, token1_decimals] = ethereum_client.batch_call(
                    tasks
                )
            else:
                [
                    token0_name, token0_symbol, token0_decimals, token1_name, token1_symbol, token1_decimals
                ] = ethereum_client.batch_call(tasks)

            await redis_conn.hset(
                name=redis_keys.get_uniswap_pair_contract_tokens_data(pair_address),
                mapping={
                    "token0_name": token0_name,
                    "token0_symbol": token0_symbol,
                    "token0_decimals": token0_decimals,
                    "token1_name": token1_name,
                    "token1_symbol": token1_symbol,
                    "token1_decimals": token1_decimals,
                    "pair_symbol": f"{token0_symbol}-{token1_symbol}",
                    'token0Addr': token0Addr,
                    'token1Addr': token1Addr
                }
            )

        return {
            'token0': {
                'address': token0Addr,
                'name': token0_name,
                'symbol': token0_symbol,
                'decimals': token0_decimals
            },
            'token1': {
                'address': token1Addr,
                'name': token1_name,
                'symbol': token1_symbol,
                'decimals': token1_decimals
            },
            'pair': {
                'symbol': f'{token0_symbol}-{token1_symbol}'
            }
        }
    except Exception as err:
        # this will be retried in next cycle
        logger.error(f"RPC error while fetcing metadata for pair {pair_address}, error_msg:{err}", exc_info=True)
        raise err

def calculate_pair_trade_volume(dag_chain):
    # calculate / sum trade volume
    total_volume = sum(map(lambda x: x['data']['payload']['totalTrade'], dag_chain))

    # calculate / sum fee
    fees = sum(map(lambda x: x['data']['payload'].get('totalFee', 0), dag_chain))

    # get volume for individual tokens (in native token decimals):
    token0_volume = sum(map(lambda x: x['data']['payload']['token0TradeVolume'], dag_chain))
    token1_volume = sum(map(lambda x: x['data']['payload']['token1TradeVolume'], dag_chain))

    token0_volume_usd = sum(map(lambda x: x['data']['payload']['token0TradeVolumeUSD'], dag_chain))
    token1_volume_usd = sum(map(lambda x: x['data']['payload']['token1TradeVolumeUSD'], dag_chain))

    return [total_volume, fees, token0_volume, token1_volume, token0_volume_usd, token1_volume_usd]

async def calculate_pair_liquidity(writer_redis_conn: aioredis.Redis, pair_contract_address):
    project_id_token_reserve = f'uniswap_pairContract_pair_total_reserves_{pair_contract_address}_{NAMESPACE}'

    # liquidty data
    liquidity_head_marker = await writer_redis_conn.get(redis_keys.get_sliding_window_cache_head_marker(project_id_token_reserve, '24h'))
    liquidity_head_marker = int(liquidity_head_marker.decode('utf-8')) if liquidity_head_marker else 0
    liquidity_data = await get_dag_block_by_height(
        project_id_token_reserve,
        liquidity_head_marker,
        writer_redis_conn
    )

    if not liquidity_data:
        return [None, None, None, None]

    token0_liquidity = float(list(liquidity_data['data']['payload']['token0Reserves'].values())[-1])
    token1_liquidity = float(list(liquidity_data['data']['payload']['token1Reserves'].values())[-1])
    token0_liquidity_usd = float(list(liquidity_data['data']['payload']['token0ReservesUSD'].values())[-1])
    token1_liquidity_usd = float(list(liquidity_data['data']['payload']['token1ReservesUSD'].values())[-1])
    total_liquidity = token0_liquidity_usd + token1_liquidity_usd
    block_height_total_reserve = int(liquidity_data['data']['payload']['chainHeightRange']['end'])
    block_timestamp_total_reserve = int(liquidity_data['data']['payload']['timestamp'])
    if not block_timestamp_total_reserve:
        block_timestamp_total_reserve = int(liquidity_data['timestamp'])

    return [
        total_liquidity,
        token0_liquidity,
        token1_liquidity,
        token0_liquidity_usd,
        token1_liquidity_usd,
        block_height_total_reserve,
        block_timestamp_total_reserve
    ]


def patch_cids_obj(dag_chain, patch_type, patch):
    cid_list = []
    if patch_type == "parse_cids":
        # parse dagCid and payloadCid from dag chain
        cid_list = [{ 'dagCid': obj['dagCid'], 'payloadCid': obj['data']['cid']} for obj in dag_chain]
    elif patch_type == "add_front_patch":
        # add sliding window front cids to existing cid obj list
        cid_list = patch + dag_chain
    elif patch_type == "remove_back_patch":
        # remove sliding window back cids from existing cid obj list
        cid_list = dag_chain[:]
        for dag_obj in cid_list:
            for patch_obj in patch:
                if dag_obj["dagCid"] == patch_obj["dagCid"]:
                    dag_chain.remove(dag_obj)
        cid_list = dag_chain[:]

    return cid_list


async def store_recent_transactions_logs(writer_redis_conn: aioredis.Redis, dag_chain, pair_contract_address):
    recent_logs = []
    for log in dag_chain:
        recent_logs.extend(log['data']['payload']['recent_logs'])
    recent_logs = sorted(recent_logs, key=lambda log: log['timestamp'], reverse=True)

    oldLogs = await writer_redis_conn.get(redis_keys.get_uniswap_pair_cached_recent_logs(f"{Web3.toChecksumAddress(pair_contract_address)}"))
    if oldLogs:
        oldLogs = json.loads(oldLogs.decode('utf-8'))
        recent_logs = recent_logs + oldLogs

    # trime logs to be max 75 each pair
    if len(recent_logs) > 75:
        recent_logs = recent_logs[:75]

    logger.debug(f"Storing recent logs for pair: {redis_keys.get_uniswap_pair_cached_recent_logs(f'{Web3.toChecksumAddress(pair_contract_address)}')}, of len:{len(recent_logs)}")
    await writer_redis_conn.set(redis_keys.get_uniswap_pair_cached_recent_logs(f"{Web3.toChecksumAddress(pair_contract_address)}"), json.dumps(recent_logs))
    return recent_logs


async def store_pair_daily_stats(writer_redis_conn: aioredis.Redis, pair_contract_address, snapshot):
    now = int(time())
    await writer_redis_conn.zadd(
        name=redis_keys.get_uniswap_pair_cache_daily_stats(Web3.toChecksumAddress(pair_contract_address)),
        mapping={json.dumps(snapshot.json()): now}
    )
    await writer_redis_conn.zremrangebyscore(
        name=redis_keys.get_uniswap_pair_cache_daily_stats(Web3.toChecksumAddress(pair_contract_address)),
        min=float('-inf'),
        max=float(now - 60 * 60 * 25)
    )


async def process_pairs_trade_volume_and_reserves(writer_redis_conn: aioredis.Redis, pair_contract_address):
    try:
        project_id_trade_volume = f'uniswap_pairContract_trade_volume_{pair_contract_address}_{NAMESPACE}'
        project_id_token_reserve = f'uniswap_pairContract_pair_total_reserves_{pair_contract_address}_{NAMESPACE}'

        # get head, tail and sliding window data from redis
        [
            cached_trade_volume_data,
            tail_marker_24h,
            head_marker_24h,
            tail_marker_7d,
            head_marker_7d
        ] = await writer_redis_conn.mget([
            redis_keys.get_uniswap_pair_cache_sliding_window_data(f"{Web3.toChecksumAddress(pair_contract_address)}"),
            redis_keys.get_sliding_window_cache_tail_marker(project_id_trade_volume, '24h'),
            redis_keys.get_sliding_window_cache_head_marker(project_id_trade_volume, '24h'),
            redis_keys.get_sliding_window_cache_tail_marker(project_id_trade_volume, '7d'),
            redis_keys.get_sliding_window_cache_head_marker(project_id_trade_volume, '7d')
        ])

        cached_trade_volume_data = json.loads(cached_trade_volume_data.decode('utf-8')) if cached_trade_volume_data else {}
        tail_marker_24h = int(tail_marker_24h.decode('utf-8')) if tail_marker_24h else 0
        head_marker_24h = int(head_marker_24h.decode('utf-8')) if head_marker_24h else 0
        tail_marker_7d = int(tail_marker_7d.decode('utf-8')) if tail_marker_7d else 0
        head_marker_7d = int(head_marker_7d.decode('utf-8')) if head_marker_7d else 0

        if any([
                not tail_marker_24h,
                not head_marker_24h,
                not tail_marker_7d,
                not head_marker_7d
            ]):
            logger.error(f"Avoding pair data calculations as head or tail is empty for pair:{pair_contract_address}, tail:{tail_marker_24h}, head:{head_marker_24h}")
            return

        # if head of 24h and 7d index at 1 or less then return
        if head_marker_24h <= 1 or head_marker_7d <=1:
            logger.error(f"Avoding pair data calculations as head and tail are set to 1 or 0 block for pair:{pair_contract_address}, tail:{tail_marker_24h}, head:{head_marker_24h}")
            return

        # initialize returning variables
        total_liquidity = 0
        token0_liquidity = 0
        token1_liquidity = 0
        recent_logs = list()
        block_height_total_reserve = 0
        block_height_trade_volume = 0
        block_timestamp_total_reserve = 0
        pair_token_metadata = {}

        # 24h variables
        total_volume_24h = 0
        token0_volume_24h = 0
        token1_volume_24h = 0
        token0_volume_usd_24h = 0
        token1_volume_usd_24h = 0
        fees_24h = 0
        cids_volume_24h = ''

        # 7d variables
        total_volume_7d = 0
        token0_volume_7d = 0
        token1_volume_7d = 0
        token0_volume_usd_7d = 0
        token1_volume_usd_7d = 0
        fees_7d = 0
        cids_volume_7d = ''


        pair_token_metadata = await get_pair_tokens_metadata(
            pair_address=Web3.toChecksumAddress(pair_contract_address),
            redis_conn=writer_redis_conn
        ) 

        [
            total_liquidity,
            token0_liquidity,
            token1_liquidity,
            token0_liquidity_usd,
            token1_liquidity_usd,
            block_height_total_reserve,
            block_timestamp_total_reserve
        ] = await calculate_pair_liquidity(writer_redis_conn, pair_contract_address)
        if not total_liquidity:
            logger.error(f"Error, exit process as liquidity data is not available - projectId:{project_id_token_reserve}")
            return

        if not cached_trade_volume_data:
            dag_chain_24h = await get_dag_blocks_in_range(project_id_trade_volume, tail_marker_24h, head_marker_24h, writer_redis_conn)
            if not dag_chain_24h:
                logger.error(f"dag_chain_24h array is empty for pair:{pair_contract_address}, avoiding trade volume calculations")
                return

            dag_chain_7d = await get_dag_blocks_in_range(project_id_trade_volume, tail_marker_7d, head_marker_7d, writer_redis_conn)
            if not dag_chain_7d:
                logger.error(f"dag_chain_7d array is empty for pair:{pair_contract_address}, avoiding trade volume calculations")
                return

            # calculate trade volume 24h
            [total_volume_24h, fees_24h, token0_volume_24h, token1_volume_24h, token0_volume_usd_24h, token1_volume_usd_24h] = calculate_pair_trade_volume(dag_chain_24h)

            # calculate trade volume 7d
            [total_volume_7d, fees_7d, token0_volume_7d, token1_volume_7d, token0_volume_usd_7d, token1_volume_usd_7d] = calculate_pair_trade_volume(dag_chain_7d)

            # parse and store dag chain cids on IPFS
            volume_cids = await asyncio.gather(
                ipfs_client.add_json(uniswapPairSummary24hCidRange(
                    resultant=uniswapPairSummaryCid24hResultant(
                        trade_volume_24h_cids= {
                            "latest_dag_cid": dag_chain_24h[0]['dagCid'],
                            "oldest_dag_cid": dag_chain_24h[-1]['dagCid'],
                        },
                        latestTimestamp_volume_24h = str(dag_chain_24h[0]['timestamp'])
                    )
                ).dict()),
                ipfs_client.add_json(uniswapPairSummary7dCidRange(
                    resultant=uniswapPairSummaryCid7dResultant(
                        trade_volume_7d_cids= {
                            "latest_dag_cid": dag_chain_7d[0]['dagCid'],
                            "oldest_dag_cid": dag_chain_7d[-1]['dagCid'],
                        },
                        latestTimestamp_volume_7d = str(dag_chain_7d[0]['timestamp'])
                    )
                ).dict())
            )
            # data = await ipfs_client.cat(volume_cids[0])
            # print(f"cid get: {data}")


            # store last recent logs, these will be used to show recent transaction for perticular contract
            # using only 24h dag chain as we just need 75 log at max
            await store_recent_transactions_logs(writer_redis_conn, dag_chain_24h, pair_contract_address)

            cids_volume_24h = volume_cids[0]
            cids_volume_7d = volume_cids[1]
            block_height_trade_volume = int(dag_chain_24h[0]['data']['payload']['chainHeightRange']['end'])
        else:

            sliding_window_front_24h = []
            sliding_window_back_24h = []
            sliding_window_front_7d = []
            sliding_window_back_7d = []

            # if 24h head moved ahead
            if head_marker_24h > cached_trade_volume_data["processed_head_marker_24h"]:
                # front of the chain where head=current_head and tail=last_head
                sliding_window_front_24h = await get_dag_blocks_in_range(
                    project_id_trade_volume,
                    cached_trade_volume_data["processed_head_marker_24h"] + 1,
                    head_marker_24h,
                    writer_redis_conn
                )

            # if 24h tail moved ahead
            if tail_marker_24h > cached_trade_volume_data["processed_tail_marker_24h"]:
                # back of the chain where head=current_tail and tail=last_tail
                sliding_window_back_24h = await get_dag_blocks_in_range(
                    project_id_trade_volume,
                    cached_trade_volume_data["processed_tail_marker_24h"],
                    tail_marker_24h - 1,
                    writer_redis_conn
                )

            # if 7d head moved ahead
            if head_marker_7d > cached_trade_volume_data["processed_head_marker_24h"]:
                # front of the chain where head=current_head and tail=last_head
                sliding_window_front_7d = await get_dag_blocks_in_range(
                    project_id_trade_volume,
                    cached_trade_volume_data["processed_head_marker_7d"] + 1,
                    head_marker_7d,
                    writer_redis_conn
                )

            # if 7d tail moved ahead
            if tail_marker_7d > cached_trade_volume_data["processed_tail_marker_7d"]:
                # back of the chain where head=current_tail and tail=last_tail
                sliding_window_back_7d = await get_dag_blocks_in_range(
                    project_id_trade_volume,
                    cached_trade_volume_data["processed_tail_marker_7d"],
                    tail_marker_7d - 1,
                    writer_redis_conn
                )

            if not sliding_window_front_24h and not sliding_window_back_24h:
                return

            if not sliding_window_front_7d and not sliding_window_back_7d:
                return

            sliding_window_front_24h_cids = []
            sliding_window_back_24h_cids = []
            sliding_window_front_7d_cids = []
            sliding_window_back_7d_cids = []

            # fetch stored cids from IPFS
            [trade_volume_cids_24h, trade_volume_cids_7d] = await asyncio.gather(
                ipfs_client.cat(cached_trade_volume_data["aggregated_volume_cid_24h"]),
                ipfs_client.cat(cached_trade_volume_data["aggregated_volume_cid_7d"])
            )

            # parse and patch CID for 24h trade volume
            trade_volume_cids_24h = json.loads(trade_volume_cids_24h.decode('utf-8')) if trade_volume_cids_24h else {}
            # parse and patch CID for 7d trade volume
            trade_volume_cids_7d = json.loads(trade_volume_cids_7d.decode('utf-8')) if trade_volume_cids_7d else {}


            if sliding_window_back_24h:
                [back_total_volume, back_fees, back_token0_volume, back_token1_volume, back_token0_volume_usd, back_token1_volume_usd] = calculate_pair_trade_volume(sliding_window_back_24h)
                cached_trade_volume_data["aggregated_volume_24h"] -= back_total_volume
                cached_trade_volume_data["aggregated_fees_24h"] -= back_fees
                cached_trade_volume_data["aggregated_token0_volume_24h"] -= back_token0_volume
                cached_trade_volume_data["aggregated_token1_volume_24h"] -= back_token1_volume
                cached_trade_volume_data["aggregated_token0_volume_usd_24h"] -= back_token0_volume_usd
                cached_trade_volume_data["aggregated_token1_volume_usd_24h"] -= back_token1_volume_usd

                if trade_volume_cids_24h:
                    # set last element of back sliding window as oldest dag cid
                    trade_volume_cids_24h = json.loads(trade_volume_cids_24h) if isinstance(trade_volume_cids_24h, str) else trade_volume_cids_24h
                    if isinstance(trade_volume_cids_24h["resultant"]["trade_volume_24h_cids"], list):
                        trade_volume_cids_24h["resultant"]["trade_volume_24h_cids"] = {}
                    trade_volume_cids_24h["resultant"]["trade_volume_24h_cids"]["oldest_dag_cid"] = sliding_window_back_24h[-1]['dagCid']
                else:
                    trade_volume_cids_24h["resultant"]["trade_volume_24h_cids"]["oldest_dag_cid"] = sliding_window_back_24h[-1]['dagCid']
                    trade_volume_cids_24h["resultant"]["trade_volume_24h_cids"]["latest_dag_cid"] = sliding_window_back_24h[0]['dagCid']

            if sliding_window_front_24h:
                [front_total_volume, front_fees, front_token0_volume, front_token1_volume, front_token0_volume_usd, front_token1_volume_usd] = calculate_pair_trade_volume(sliding_window_front_24h)
                cached_trade_volume_data["aggregated_volume_24h"] += front_total_volume
                cached_trade_volume_data["aggregated_fees_24h"] += front_fees
                cached_trade_volume_data["aggregated_token0_volume_24h"] += front_token0_volume
                cached_trade_volume_data["aggregated_token1_volume_24h"] += front_token1_volume
                cached_trade_volume_data["aggregated_token0_volume_usd_24h"] += front_token0_volume_usd
                cached_trade_volume_data["aggregated_token1_volume_usd_24h"] += front_token1_volume_usd
                # block height of trade volume
                block_height_trade_volume = int(sliding_window_front_24h[0]['data']['payload']['chainHeightRange']['end'])

                if trade_volume_cids_24h:
                    trade_volume_cids_24h = json.loads(trade_volume_cids_24h) if isinstance(trade_volume_cids_24h, str) else trade_volume_cids_24h
                    # set first element of front sliding window as latest dag cid
                    if isinstance(trade_volume_cids_24h["resultant"]["trade_volume_24h_cids"], list):
                        trade_volume_cids_24h["resultant"]["trade_volume_24h_cids"] = {}
                    trade_volume_cids_24h["resultant"]["trade_volume_24h_cids"]["latest_dag_cid"] = sliding_window_front_24h[0]['dagCid']
                else:
                    trade_volume_cids_24h["resultant"]["trade_volume_24h_cids"]["latest_dag_cid"] = sliding_window_front_24h[0]['dagCid']
                    trade_volume_cids_24h["resultant"]["trade_volume_24h_cids"]["oldest_dag_cid"] = sliding_window_front_24h[-1]['dagCid']

                # store recent logs
                await store_recent_transactions_logs(writer_redis_conn, sliding_window_front_24h, pair_contract_address)
            else:
                block_height_trade_volume = cached_trade_volume_data["processed_block_height_trade_volume"]

            if sliding_window_back_7d:
                [back_total_volume, back_fees, back_token0_volume, back_token1_volume, back_token0_volume_usd, back_token1_volume_usd] = calculate_pair_trade_volume(sliding_window_back_7d)
                cached_trade_volume_data["aggregated_volume_7d"] -= back_total_volume
                cached_trade_volume_data["aggregated_token0_volume_7d"] -= back_token0_volume
                cached_trade_volume_data["aggregated_token1_volume_7d"] -= back_token1_volume
                cached_trade_volume_data["aggregated_token0_volume_usd_7d"] -= back_token0_volume_usd
                cached_trade_volume_data["aggregated_token1_volume_usd_7d"] -= back_token1_volume_usd

                if trade_volume_cids_7d:
                    trade_volume_cids_7d = json.loads(trade_volume_cids_7d) if isinstance(trade_volume_cids_7d, str) else trade_volume_cids_7d
                    # set last element of back sliding window as oldest dag cid
                    if isinstance(trade_volume_cids_7d["resultant"]["trade_volume_7d_cids"], list):
                        trade_volume_cids_7d["resultant"]["trade_volume_7d_cids"] = {}
                    trade_volume_cids_7d["resultant"]["trade_volume_7d_cids"]["oldest_dag_cid"] = sliding_window_back_7d[-1]['dagCid']
                else:
                    trade_volume_cids_7d["resultant"]["trade_volume_7d_cids"]["oldest_dag_cid"] = sliding_window_back_7d[-1]['dagCid']
                    trade_volume_cids_7d["resultant"]["trade_volume_7d_cids"]["latest_dag_cid"] = sliding_window_back_7d[0]['dagCid']


            if sliding_window_front_7d:
                [front_total_volume, front_fees, front_token0_volume, front_token1_volume, front_token0_volume_usd, front_token1_volume_usd] = calculate_pair_trade_volume(sliding_window_front_7d)
                cached_trade_volume_data["aggregated_volume_7d"] += front_total_volume
                cached_trade_volume_data["aggregated_token0_volume_7d"] += front_token0_volume
                cached_trade_volume_data["aggregated_token1_volume_7d"] += front_token1_volume
                cached_trade_volume_data["aggregated_token0_volume_usd_7d"] += front_token0_volume_usd
                cached_trade_volume_data["aggregated_token1_volume_usd_7d"] += front_token1_volume_usd

                if trade_volume_cids_7d:
                    trade_volume_cids_7d = json.loads(trade_volume_cids_7d) if isinstance(trade_volume_cids_7d, str) else trade_volume_cids_7d
                    if isinstance(trade_volume_cids_7d["resultant"]["trade_volume_7d_cids"], list):
                        trade_volume_cids_7d["resultant"]["trade_volume_7d_cids"] = {}
                    trade_volume_cids_7d["resultant"]["trade_volume_7d_cids"]["latest_dag_cid"] = sliding_window_front_7d[0]['dagCid']
                else:
                    trade_volume_cids_7d["resultant"]["trade_volume_7d_cids"]["latest_dag_cid"] = sliding_window_front_7d[0]['dagCid']
                    trade_volume_cids_7d["resultant"]["trade_volume_7d_cids"]["oldest_dag_cid"] = sliding_window_front_7d[-1]['dagCid']



            # store dag chain cids on IPFS
            latestTimestamp_volume_24h = sliding_window_front_24h[0]['timestamp'] if sliding_window_front_24h else trade_volume_cids_24h['resultant']['latestTimestamp_volume_24h']
            latestTimestamp_volume_7d = sliding_window_front_7d[0]['timestamp'] if sliding_window_front_7d else trade_volume_cids_7d['resultant']['latestTimestamp_volume_7d']
            volume_cids = await asyncio.gather(
                ipfs_client.add_json(uniswapPairSummary24hCidRange(
                    resultant=uniswapPairSummaryCid24hResultant(
                        trade_volume_24h_cids= trade_volume_cids_24h["resultant"]["trade_volume_24h_cids"],
                        latestTimestamp_volume_24h = str(latestTimestamp_volume_24h)
                    )
                ).dict()),
                ipfs_client.add_json(uniswapPairSummary7dCidRange(
                    resultant=uniswapPairSummaryCid7dResultant(
                        trade_volume_7d_cids= trade_volume_cids_7d["resultant"]["trade_volume_7d_cids"],
                        latestTimestamp_volume_7d = str(latestTimestamp_volume_7d)
                    )
                ).dict())
            )

            # if liquidity reserve block height is not found
            if not block_height_total_reserve:
                block_height_total_reserve = cached_trade_volume_data["processed_block_height_total_reserve"]
            if not block_timestamp_total_reserve:
                block_timestamp_total_reserve = cached_trade_volume_data["processed_block_timestamp_total_reserve"]

            total_volume_24h = cached_trade_volume_data['aggregated_volume_24h']
            total_volume_7d = cached_trade_volume_data['aggregated_volume_7d']
            cids_volume_24h = volume_cids[0]
            cids_volume_7d = volume_cids[1]
            fees_24h = cached_trade_volume_data['aggregated_fees_24h']
            token0_volume_24h = cached_trade_volume_data['aggregated_token0_volume_24h']
            token1_volume_24h = cached_trade_volume_data['aggregated_token1_volume_24h']
            token0_volume_7d = cached_trade_volume_data['aggregated_token0_volume_7d']
            token1_volume_7d = cached_trade_volume_data['aggregated_token1_volume_7d']
            token0_volume_usd_24h = cached_trade_volume_data['aggregated_token0_volume_usd_24h']
            token1_volume_usd_24h = cached_trade_volume_data['aggregated_token1_volume_usd_24h']
            token0_volume_usd_7d = cached_trade_volume_data['aggregated_token0_volume_usd_7d']
            token1_volume_usd_7d = cached_trade_volume_data['aggregated_token1_volume_usd_7d']

        sliding_window_data = {
            "processed_tail_marker_24h": tail_marker_24h,
            "processed_head_marker_24h": head_marker_24h,
            "processed_tail_marker_7d": tail_marker_7d,
            "processed_head_marker_7d": head_marker_7d,
            "aggregated_volume_cid_24h": cids_volume_24h,
            "aggregated_volume_cid_7d": cids_volume_7d,
            "aggregated_volume_24h": total_volume_24h,
            "aggregated_fees_24h": fees_24h,
            "aggregated_token0_volume_24h": token0_volume_24h,
            "aggregated_token1_volume_24h": token1_volume_24h,
            "aggregated_token0_volume_usd_24h": token0_volume_usd_24h,
            "aggregated_token1_volume_usd_24h": token1_volume_usd_24h,
            "aggregated_volume_7d": total_volume_7d,
            "aggregated_token0_volume_7d": token0_volume_7d,
            "aggregated_token1_volume_7d": token1_volume_7d,
            "aggregated_token0_volume_usd_7d": token0_volume_usd_7d,
            "aggregated_token1_volume_usd_7d": token1_volume_usd_7d,
            "processed_block_height_total_reserve": block_height_total_reserve,
            "processed_block_height_trade_volume": block_height_trade_volume,
            "processed_block_timestamp_total_reserve": block_timestamp_total_reserve,
        }
        prepared_snapshot = liquidityProcessedData(
            contractAddress=pair_contract_address,
            name=pair_token_metadata['pair']['symbol'],
            liquidity=f"US${round(abs(total_liquidity)):,}",
            volume_24h=f"US${round(abs(total_volume_24h)):,}",
            volume_7d=f"US${round(abs(total_volume_7d)):,}",
            fees_24h=f"US${round(abs(fees_24h)):,}",
            cid_volume_24h=cids_volume_24h,
            cid_volume_7d=cids_volume_7d,
            block_height=block_height_total_reserve,
            block_timestamp=block_timestamp_total_reserve,
            token0Liquidity=token0_liquidity,
            token1Liquidity=token1_liquidity,
            token0LiquidityUSD=float(token0_liquidity_usd),
            token1LiquidityUSD=float(token1_liquidity_usd),
            token0TradeVolume_24h=token0_volume_24h,
            token1TradeVolume_24h=token1_volume_24h,
            token0TradeVolumeUSD_24h=token0_volume_usd_24h,
            token1TradeVolumeUSD_24h=token1_volume_usd_24h,
            token0TradeVolume_7d=token0_volume_7d,
            token1TradeVolume_7d=token1_volume_7d,
            token0TradeVolumeUSD_7d=token0_volume_usd_7d,
            token1TradeVolumeUSD_7d=token1_volume_usd_7d
        )

        await store_pair_daily_stats(writer_redis_conn, pair_contract_address, prepared_snapshot)

        logger.debug('Storing prepared trades, token reserves snapshot: %s', prepared_snapshot)
        logger.debug('Adding calculated snapshot to daily stats zset and pruning it for just 24h data')
        await writer_redis_conn.mset({
            redis_keys.get_uniswap_pair_contract_V2_pair_data(f"{Web3.toChecksumAddress(pair_contract_address)}"): prepared_snapshot.json(),
            redis_keys.get_uniswap_pair_cache_sliding_window_data(f"{Web3.toChecksumAddress(pair_contract_address)}"): json.dumps(sliding_window_data)
        })
        logger.debug(f"Calculated v2 pair data for contract: {pair_contract_address} | symbol:{pair_token_metadata['token0']['symbol']}-{pair_token_metadata['token1']['symbol']}")
        return prepared_snapshot
    except Exception as e:
        logger.error('Error in process_pairs_trade_volume_and_reserves: %s', e, exc_info=True)
        raise


async def v2_pairs_data():
    f = None
    try:
        if not os.path.exists('static/cached_pair_addresses.json'):
            return []
        f = open('static/cached_pair_addresses.json', 'r')
        pairs = json.loads(f.read())

        if len(pairs) <= 0:
            return []

        aioredis_pool = RedisPool()
        await aioredis_pool.populate()
        redis_conn: aioredis.Redis = aioredis_pool.writer_redis_pool

        logger.debug("Create threads to process trade volume data")
        process_data_list = []
        for pair_contract_address in pairs:
            t = process_pairs_trade_volume_and_reserves(aioredis_pool.writer_redis_pool, pair_contract_address)
            process_data_list.append(t)

        final_results: Iterable[Union[liquidityProcessedData, BaseException]] = await asyncio.gather(*process_data_list, return_exceptions=True)
        common_blockheight_reached = False
        common_block_timestamp = False
        pair_contract_address = None
        collected_heights = list(map(lambda x: x.block_height, filter(lambda y: isinstance(y, liquidityProcessedData), final_results)))
        collected_block_timestamp = list(map(lambda x: x.block_timestamp, filter(lambda y: isinstance(y, liquidityProcessedData), final_results)))
        pair_addresses = list(map(lambda x: x.contractAddress, filter(lambda y: isinstance(y, liquidityProcessedData), final_results)))
        # logger.debug('Final results from running v2 pairs coroutines: %s', final_results)
        # logger.debug('Filtered heights from running v2 pairs coroutines: %s', collected_heights)
        if cardinality.count(final_results) != cardinality.count(collected_heights):
            logger.error(
                         'In V2 pairs overall summary snapshot creation: '
                         'No common block height found among all summarized contracts. Some results returned exception.'
                         'Block heights found: %s', list(collected_heights)
            )
            common_blockheight_reached = False
        else:
            if all(collected_heights[0] == y for y in collected_heights):
                common_blockheight_reached = collected_heights[0]
                common_block_timestamp = collected_block_timestamp[0]
                pair_contract_address = pair_addresses[0]
                logger.debug('Setting common blockheight reached to %s', collected_heights[0])
            else:
                logger.error(
                    'In V2 pairs overall summary snapshot creation: '
                    'No common block height found among all summarized contracts.'
                    'Block heights found: %s', list(collected_heights)
                )
        wait_for_snapshot_project_new_commit = False
        current_audit_project_block_height = 0
        if common_blockheight_reached:
            # TODO: make these configurable
            async_httpx_client = AsyncClient(
                timeout=Timeout(timeout=5.0),
                follow_redirects=False,
                limits=Limits(max_connections=100, max_keepalive_connections=20, keepalive_expiry=5.0)
            )
            summarized_result_l = [x.dict() for x in final_results]
            summarized_payload = {'data': summarized_result_l}
            # get last processed v2 pairs snapshot height
            l_ = await redis_conn.get(redis_keys.get_uniswap_pair_snapshot_last_block_height())
            if l_:
                last_common_block_height = int(l_)
            else:
                last_common_block_height = 0
            if common_blockheight_reached > last_common_block_height:
                logger.info(
                    'Present common block height in V2 pairs summary snapshot is %s | Moved from %s',
                    common_blockheight_reached, last_common_block_height
                )
                current_audit_project_block_height = await helper_functions.get_block_height(
                    project_id=redis_keys.get_uniswap_pairs_summary_snapshot_project_id(),
                    reader_redis_conn=redis_conn
                )
                logger.debug('Sending v2 pairs summary payload to audit protocol')
                # send to audit protocol for snapshot to be committed
                try:
                    response = await helper_functions.commit_payload(
                        project_id=redis_keys.get_uniswap_pairs_summary_snapshot_project_id(),
                        report_payload=summarized_payload,
                        session=async_httpx_client,
                        skipAnchorProof=False
                    )
                except Exception as e:
                    logger.error(
                        'Error while committing pairs summary snapshot at block height %s to audit protocol. '
                        'Exception: %s',
                        common_blockheight_reached, e, exc_info=True
                    )
                else:
                    if 'message' in response.keys():
                        logger.error(
                            'Error while committing pairs summary snapshot at block height %s to audit protocol. '
                            'Response status code and other information: %s',
                            common_blockheight_reached, response
                        )
                    else:
                        wait_for_snapshot_project_new_commit = True
                        updated_audit_project_block_height = current_audit_project_block_height+1
        if wait_for_snapshot_project_new_commit:
            waitCycles = 0
            while True:
                # introduce a break condition if something goes wrong and snapshot summary does not move ahead
                waitCycles+=1
                if waitCycles > 18: # Wait for 60 seconds after which move ahead as something must have has gone wrong with snapshot summary submission
                    logger.info(f"Waited for {waitCycles} cycles, snapshot summary project has not moved ahead. Stopped waiting to retry in next cycle.")
                    break
                logger.debug('Waiting for 10 seconds to check if latest v2 pairs summary snapshot was committed...')
                await asyncio.sleep(10)

                block_status = await retrieval_utils.retrieve_block_status(
                                    redis_keys.get_uniswap_pairs_summary_snapshot_project_id(),
                                    0,updated_audit_project_block_height,redis_conn,redis_conn)
                if block_status.status < 3:
                    continue
                logger.info(
                    'Audit project height against V2 pairs summary snapshot is %s | Moved from %s',
                    updated_audit_project_block_height, current_audit_project_block_height
                )
                begin_block_data = await get_oldest_block_and_timestamp(pair_contract_address)

                snapshotZsetEntry = uniswapPairsSnapshotZset(
                    cid=block_status.payload_cid,
                    txHash=block_status.tx_hash,
                    begin_block_height_24h=begin_block_data["begin_block_height_24h"],
                    begin_block_timestamp_24h=begin_block_data["begin_block_timestamp_24h"],
                    begin_block_height_7d=begin_block_data["begin_block_height_7d"],
                    begin_block_timestamp_7d=begin_block_data["begin_block_timestamp_7d"]
                )

                # store in snapshots zset
                await asyncio.gather(
                    redis_conn.zadd(
                        name=redis_keys.get_uniswap_pair_snapshot_summary_zset(),
                        mapping={snapshotZsetEntry.json(): common_blockheight_reached}),
                    redis_conn.zadd(
                        name=redis_keys.get_uniswap_pair_snapshot_timestamp_zset(),
                        mapping={block_status.payload_cid: common_block_timestamp}),
                    redis_conn.set(
                        name=redis_keys.get_uniswap_pair_snapshot_payload_at_blockheight(common_blockheight_reached),
                        value=json.dumps(summarized_payload),
                        ex=1800  # TTL of 30 minutes?
                    ),
                    redis_conn.set(
                        redis_keys.get_uniswap_pair_snapshot_last_block_height(),
                        common_blockheight_reached)
                )

                #prune zset
                block_height_zset_len = await redis_conn.zcard(name=redis_keys.get_uniswap_pair_snapshot_summary_zset())

                # TODO: snapshot history zset size limit configurable
                if block_height_zset_len > 20:
                    _ = await redis_conn.zremrangebyrank(
                        name=redis_keys.get_uniswap_pair_snapshot_summary_zset(),
                        min=0,
                        max=-1 * (block_height_zset_len - 20) + 1
                    )
                    logger.debug('Pruned snapshot summary CID zset by %s elements', _)

                pruning_timestamp = int(time()) - (60 * 60 * 24 + 60 * 60 * 2) # now - 26 hours
                _ = await redis_conn.zremrangebyscore(
                    name=redis_keys.get_uniswap_pair_snapshot_timestamp_zset(),
                    min=0,
                    max=pruning_timestamp
                )
                logger.debug('Pruned snapshot summary Timestamp zset by %s elements', _)
                logger.info('V2 pairs summary snapshot updated...')

                break

        return final_results

    except Exception as e:
        logger.error(f"Error at V2 pair data: {str(e)}", exc_info=True)
    finally:
        if f is not None:
            f.close()

if __name__ == '__main__':
    print("", "")
    # loop = asyncio.get_event_loop()
    # data = loop.run_until_complete(v2_pairs_data())
    # print("## ##", "")
    # print(data)
