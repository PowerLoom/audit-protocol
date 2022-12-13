from typing import Union, Iterable
from time import time
from web3 import Web3
from httpx import AsyncClient, Timeout, Limits
from redis import asyncio as aioredis
from async_ipfshttpclient.main import AsyncIPFSClient
from gnosis.eth import EthereumClient
from config import settings
from data_models import liquidityProcessedData, uniswapPairsSnapshotZset, DAGBlockRange, PairLiquidity, PairTradeVolume
from utils import helper_functions
from utils import redis_keys
from utils import retrieval_utils
from utils.redis_conn import RedisPool
from utils.retrieval_utils import get_dag_block_by_height
import asyncio
import json
import logging.config
import sys
import os
import cardinality

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

# read all pair-contracts
PAIR_CONTRACTS = []
if os.path.exists('static/cached_pair_addresses.json'):
    f = open('static/cached_pair_addresses.json', 'r')
    PAIR_CONTRACTS = json.loads(f.read())


async def get_dag_blocks_in_range(
        project_id,
        from_block,
        to_block,
        reader_redis_conn: aioredis.Redis,
        ipfs_read_client: AsyncIPFSClient
):
    dag_chain = []
    for i in range(from_block, to_block + 1):
        try:
            dag_block = await get_dag_block_by_height(
                project_id=project_id,
                block_height=i,
                reader_redis_conn=reader_redis_conn,
                ipfs_read_client=ipfs_read_client,
                cache_size_unit=len(PAIR_CONTRACTS)
            )
        except Exception as e:
            logger.error("Error: can't get dag block with msg: %s |"
                         "projectId:%s, block_height:%s",
                         str(e),
                         project_id,
                         i,
                         exc_info=True)
            dag_block = {}
        if dag_block:
            dag_chain.append(dag_block)

    dag_chain.reverse()
    return dag_chain


def read_json_file(file_path: str):
    """Read given json file and return its content as a dictionary."""
    try:
        file = open(file_path, 'r')
    except Exception as exc:
        logger.warning("Unable to open the %s file",file_path)
        logger.error(exc, exc_info=True)
        raise exc
    else:
        json_data = json.loads(file.read())
    return json_data


pair_contract_abi = read_json_file("abis/UniswapV2Pair.json")
erc20_abi = read_json_file('abis/IERC20.json')


def get_maker_pair_data(prop):
    prop = prop.lower()
    if prop.lower() == "name":
        return "Maker"
    elif prop.lower() == "symbol":
        return "MKR"
    else:
        return "Maker"


async def get_oldest_block_and_timestamp(
        pair_contract_address,
        redis_conn,
        ipfs_read_client
):
    project_id_trade_volume = f'uniswap_pairContract_trade_volume_{pair_contract_address}_{NAMESPACE}'

    # trade volume data
    volume_tail_marker_24h, volume_tail_marker_7d = await redis_conn.mget(
        redis_keys.get_sliding_window_cache_tail_marker(project_id_trade_volume, '24h'),
        redis_keys.get_sliding_window_cache_tail_marker(project_id_trade_volume, '7d')
    )
    volume_tail_marker_24h = int(volume_tail_marker_24h.decode('utf-8')) if volume_tail_marker_24h else 0
    volume_tail_marker_7d = int(volume_tail_marker_7d.decode('utf-8')) if volume_tail_marker_7d else 0

    volume_data_24h, volume_data_7d = await asyncio.gather(
        get_dag_block_by_height(
            project_id=project_id_trade_volume,
            block_height=volume_tail_marker_24h,
            reader_redis_conn=redis_conn,
            ipfs_read_client=ipfs_read_client,
            cache_size_unit=len(PAIR_CONTRACTS)
        ),
        get_dag_block_by_height(
            project_id=project_id_trade_volume,
            block_height=volume_tail_marker_7d,
            reader_redis_conn=redis_conn,
            ipfs_read_client=ipfs_read_client,
            cache_size_unit=len(PAIR_CONTRACTS)
        )
    )

    if not volume_data_24h or not volume_data_7d:
        return {
            "begin_block_height_24h": 0,
            "begin_block_timestamp_24h": 0,
            "begin_block_height_7d": 0,
            "begin_block_timestamp_7d": 0,
        }

    return {
        "begin_block_height_24h": int(volume_data_24h['data']['payload']['chainHeightRange']['begin']),
        "begin_block_timestamp_24h": int(volume_data_24h['data']['payload']['timestamp']),
        "begin_block_height_7d": int(volume_data_7d['data']['payload']['chainHeightRange']['begin']),
        "begin_block_timestamp_7d": int(volume_data_7d['data']['payload']['timestamp']),
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
        pair_tokens_metadata_cache = await redis_conn.hgetall(
            redis_keys.get_uniswap_pair_contract_tokens_data(pair_address))

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

            # special case to handle maker token
            maker_token0 = None
            maker_token1 = None
            if Web3.toChecksumAddress(settings.contract_addresses.MAKER) == Web3.toChecksumAddress(token0Addr):
                token0_name = get_maker_pair_data('name')
                token0_symbol = get_maker_pair_data('symbol')
                maker_token0 = True
            else:
                tasks.append(token0.functions.name())
                tasks.append(token0.functions.symbol())
            tasks.append(token0.functions.decimals())

            if Web3.toChecksumAddress(settings.contract_addresses.MAKER) == Web3.toChecksumAddress(token1Addr):
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
        logger.error(
            "RPC error while fetching metadata for pair %s, error_msg:%s",
            pair_address,
            err,
            exc_info=True
        )
        raise err


def calculate_pair_trade_volume(dag_chain):
    pair_trade_volume: PairTradeVolume = PairTradeVolume()
    # calculate / sum trade volume
    pair_trade_volume.total_volume = sum(map(lambda x: x['data']['payload']['totalTrade'], dag_chain))

    # calculate / sum fee
    pair_trade_volume.fees = sum(map(lambda x: x['data']['payload'].get('totalFee', 0), dag_chain))

    # get volume for individual tokens (in native token decimals):
    pair_trade_volume.token0_volume = sum(map(lambda x: x['data']['payload']['token0TradeVolume'], dag_chain))
    pair_trade_volume.token1_volume = sum(map(lambda x: x['data']['payload']['token1TradeVolume'], dag_chain))

    pair_trade_volume.token0_volume_usd = sum(map(lambda x: x['data']['payload']['token0TradeVolumeUSD'], dag_chain))
    pair_trade_volume.token1_volume_usd = sum(map(lambda x: x['data']['payload']['token1TradeVolumeUSD'], dag_chain))

    return pair_trade_volume


async def calculate_pair_liquidity(
        writer_redis_conn: aioredis.Redis,
        pair_contract_address,
        ipfs_read_client: AsyncIPFSClient
):
    project_id_token_reserve = f'uniswap_pairContract_pair_total_reserves_{pair_contract_address}_{NAMESPACE}'

    # liquidity data
    liquidity_head_marker = await writer_redis_conn.get(
        redis_keys.get_sliding_window_cache_head_marker(project_id_token_reserve, '0'))
    liquidity_head_marker = int(liquidity_head_marker.decode('utf-8')) if liquidity_head_marker else 0
    liquidity_data = await get_dag_block_by_height(
        project_id=project_id_token_reserve,
        block_height=liquidity_head_marker,
        reader_redis_conn=writer_redis_conn,
        ipfs_read_client=ipfs_read_client,
        cache_size_unit=len(PAIR_CONTRACTS)
    )
    pair_liquidity: PairLiquidity = PairLiquidity()

    if not liquidity_data:
        return pair_liquidity
    pair_liquidity.token0_liquidity = float(list(liquidity_data['data']['payload']['token0Reserves'].values())[-1])
    pair_liquidity.token1_liquidity = float(list(liquidity_data['data']['payload']['token1Reserves'].values())[-1])
    pair_liquidity.token0_liquidity_usd = float(list(liquidity_data['data']['payload']['token0ReservesUSD'].values())[-1])
    pair_liquidity.token1_liquidity_usd = float(list(liquidity_data['data']['payload']['token1ReservesUSD'].values())[-1])
    pair_liquidity.total_liquidity = pair_liquidity.token0_liquidity_usd + pair_liquidity.token1_liquidity_usd
    pair_liquidity.block_height_total_reserve = int(liquidity_data['data']['payload']['chainHeightRange']['end'])
    pair_liquidity.block_timestamp_total_reserve = int(liquidity_data['data']['payload']['timestamp'])
    if not pair_liquidity.block_timestamp_total_reserve:
        pair_liquidity.block_timestamp_total_reserve = int(liquidity_data['timestamp'])

    return pair_liquidity


async def store_recent_transactions_logs(writer_redis_conn: aioredis.Redis, dag_chain, pair_contract_address):
    recent_logs = []
    for log in dag_chain:
        event_logs = log['data']['payload']['events']['Swap']['logs'] + log['data']['payload']['events']['Mint'][
            'logs'] + log['data']['payload']['events']['Burn']['logs']
        recent_logs.extend(event_logs)

    recent_logs = sorted(recent_logs, key=lambda log: log['blockNumber'], reverse=True)

    oldLogs = await writer_redis_conn.get(
        redis_keys.get_uniswap_pair_cached_recent_logs(f"{Web3.toChecksumAddress(pair_contract_address)}"))
    if oldLogs:
        oldLogs = json.loads(oldLogs.decode('utf-8'))
        recent_logs = recent_logs + oldLogs

    # trime logs to be max 75 each pair
    if len(recent_logs) > 75:
        recent_logs = recent_logs[:75]

    logger.debug(
        "Storing recent logs for pair: %s, of len:%s",
        pair_contract_address,
        len(recent_logs)
        )
    await writer_redis_conn.set(
        redis_keys.get_uniswap_pair_cached_recent_logs(f"{Web3.toChecksumAddress(pair_contract_address)}"),
        json.dumps(recent_logs))
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


async def process_pairs_trade_volume_and_reserves(
        writer_redis_conn: aioredis.Redis,
        pair_contract_address,
        ipfs_read_client: AsyncIPFSClient
):
    try:
        project_id_trade_volume = f'uniswap_pairContract_trade_volume_{pair_contract_address}_{NAMESPACE}'
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

        cached_trade_volume_data = json.loads(
            cached_trade_volume_data.decode('utf-8')) if cached_trade_volume_data else {}
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
            logger.info(
                'Values of 24h head, tail and 7h head, tail marker - some of them might be null and hence returning |'
                '[tail_marker_24h, head_marker_24h,  tail_marker_7d, head_marker_7d] : %s',
                [tail_marker_24h, head_marker_24h, tail_marker_7d, head_marker_7d]
            )
            return

        # if head of 24h and 7d index at 1 or less then return
        if head_marker_24h <= 1 or head_marker_7d <= 1:
            return

        # initialize returning variables
        block_height_trade_volume = 0
        pair_token_metadata = {}

        # 24h variables
        pair_trade_volume_24h:PairTradeVolume = PairTradeVolume()
        cids_volume_24h : DAGBlockRange = DAGBlockRange(head_block_cid="",tail_block_cid="")

        # 7d variables
        pair_trade_volume_7d:PairTradeVolume = PairTradeVolume()
        cids_volume_7d : DAGBlockRange = DAGBlockRange(head_block_cid="",tail_block_cid="")

        pair_token_metadata = await get_pair_tokens_metadata(
            pair_address=Web3.toChecksumAddress(pair_contract_address),
            redis_conn=writer_redis_conn
        )

        pair_liquidity: PairLiquidity = await calculate_pair_liquidity(writer_redis_conn, pair_contract_address, ipfs_read_client)
        if not pair_liquidity.total_liquidity:
            return

        if not cached_trade_volume_data:
            logger.debug(
                "Starting to fetch 24h dag-chain for the first time, range: %s - %s | projectId: %s",
                tail_marker_24h,
                head_marker_24h,
                project_id_trade_volume
            )
            dag_chain_24h = await get_dag_blocks_in_range(
                project_id_trade_volume, tail_marker_24h, head_marker_24h, writer_redis_conn, ipfs_read_client
            )
            if not dag_chain_24h:
                return

            logger.debug(
                "Starting to fetch 7d dag-chain for the first time, range: %s - %s | projectId: %s",
                tail_marker_7d,
                head_marker_7d,
                project_id_trade_volume
            )
            dag_chain_7d = await get_dag_blocks_in_range(
                project_id_trade_volume, tail_marker_7d, head_marker_7d, writer_redis_conn, ipfs_read_client
            )
            if not dag_chain_7d:
                return

            # calculate trade volume 24h
            pair_trade_volume_24h = calculate_pair_trade_volume(dag_chain_24h)

            # calculate trade volume 7d
            pair_trade_volume_7d = calculate_pair_trade_volume(dag_chain_7d)

            # store last recent logs, these will be used to show recent transaction for perticular contract
            # using only 24h dag chain as we just need 75 log at max
            await store_recent_transactions_logs(writer_redis_conn, dag_chain_24h, pair_contract_address)

            cids_volume_24h: DAGBlockRange = DAGBlockRange(
                head_block_cid= dag_chain_24h[0]['dagCid'],
                tail_block_cid=dag_chain_24h[-1]['dagCid']
            )
            cids_volume_7d: DAGBlockRange = DAGBlockRange(
                head_block_cid= dag_chain_7d[0]['dagCid'],
                tail_block_cid=dag_chain_7d[-1]['dagCid']
                )
            block_height_trade_volume = int(dag_chain_24h[0]['data']['payload']['chainHeightRange']['end'])
        else:
            sliding_window_front_chain_24h = []
            sliding_window_back_chain_24h = []
            sliding_window_front_chain_7d = []
            sliding_window_back_chain_7d = []

            logger.debug(
                "Starting to fetch 24h sliding window, front: %s - %s, back: %s - %s | projectId: %s",
                    cached_trade_volume_data['processed_head_marker_24h'] + 1, head_marker_24h,
                    cached_trade_volume_data['processed_tail_marker_24h'], tail_marker_24h - 1,
                    project_id_trade_volume
                )

            # if 24h head moved ahead
            if head_marker_24h > cached_trade_volume_data["processed_head_marker_24h"]:
                # front of the chain where head=current_head and tail=last_head
                sliding_window_front_chain_24h = await get_dag_blocks_in_range(
                    project_id_trade_volume,
                    cached_trade_volume_data["processed_head_marker_24h"] + 1,
                    head_marker_24h,
                    writer_redis_conn,
                    ipfs_read_client
                )
            else:
                logger.debug("24h head is at %d and did not move ahead for project %s",
                head_marker_24h,
                project_id_trade_volume)
                return

            # if 24h tail moved ahead
            if tail_marker_24h > cached_trade_volume_data["processed_tail_marker_24h"]:
                # back of the chain where head=current_tail and tail=last_tail
                sliding_window_back_chain_24h = await get_dag_blocks_in_range(
                    project_id_trade_volume,
                    cached_trade_volume_data["processed_tail_marker_24h"],
                    tail_marker_24h-1,
                    writer_redis_conn,
                    ipfs_read_client
                )

            logger.debug(
                "Starting to fetch 7d sliding window, front: %s - %s, back: %s - %s | projectId: %s",
                    cached_trade_volume_data['processed_head_marker_7d'] + 1, head_marker_7d,
                    cached_trade_volume_data['processed_tail_marker_7d'], tail_marker_7d - 1,
                    project_id_trade_volume
            )

            # if 7d head moved ahead
            if head_marker_7d > cached_trade_volume_data["processed_head_marker_24h"]:
                # front of the chain where head=current_head and tail=last_head
                sliding_window_front_chain_7d = await get_dag_blocks_in_range(
                    project_id_trade_volume,
                    cached_trade_volume_data["processed_head_marker_7d"] + 1,
                    head_marker_7d,
                    writer_redis_conn,
                    ipfs_read_client
                )

            # if 7d tail moved ahead
            if tail_marker_7d > cached_trade_volume_data["processed_tail_marker_7d"]:
                # back of the chain where head=current_tail and tail=last_tail
                sliding_window_back_chain_7d = await get_dag_blocks_in_range(
                    project_id_trade_volume,
                    cached_trade_volume_data["processed_tail_marker_7d"],
                    tail_marker_7d-1,
                    writer_redis_conn,
                    ipfs_read_client
                )

            if not sliding_window_front_chain_24h and not sliding_window_back_chain_24h:
                return

            if not sliding_window_front_chain_7d and not sliding_window_back_chain_7d:
                return
            #TODO: Need to fix this index based access to data model based access.
            cids_volume_24h.head_block_cid = sliding_window_front_chain_24h[0]['dagCid']
            tail_dag_block = await get_dag_blocks_in_range(
                project_id_trade_volume,
                tail_marker_24h,
                tail_marker_24h,
                writer_redis_conn,
                ipfs_read_client
            )
            cids_volume_24h.tail_block_cid = tail_dag_block[0]['dagCid']
            cids_volume_7d.head_block_cid = sliding_window_front_chain_7d[0]['dagCid']
            tail_dag_block = await get_dag_blocks_in_range(
                project_id_trade_volume,
                tail_marker_7d,
                tail_marker_7d,
                writer_redis_conn,
                ipfs_read_client
            )
            cids_volume_7d.tail_block_cid = tail_dag_block[0]['dagCid']

            if sliding_window_back_chain_24h:
                sliding_window_back_volume_24h:PairTradeVolume = calculate_pair_trade_volume(sliding_window_back_chain_24h)
                cached_trade_volume_data["aggregated_volume_24h"] -= sliding_window_back_volume_24h.total_volume
                cached_trade_volume_data["aggregated_fees_24h"] -= sliding_window_back_volume_24h.fees
                cached_trade_volume_data["aggregated_token0_volume_24h"] -= sliding_window_back_volume_24h.token0_volume
                cached_trade_volume_data["aggregated_token1_volume_24h"] -= sliding_window_back_volume_24h.token1_volume
                cached_trade_volume_data["aggregated_token0_volume_usd_24h"] -= sliding_window_back_volume_24h.token0_volume_usd
                cached_trade_volume_data["aggregated_token1_volume_usd_24h"] -= sliding_window_back_volume_24h.token1_volume_usd

            if sliding_window_front_chain_24h:
                sliding_window_front_volume_24h:PairTradeVolume =  calculate_pair_trade_volume(sliding_window_front_chain_24h)
                cached_trade_volume_data["aggregated_volume_24h"] += sliding_window_front_volume_24h.total_volume
                cached_trade_volume_data["aggregated_fees_24h"] += sliding_window_front_volume_24h.fees
                cached_trade_volume_data["aggregated_token0_volume_24h"] += sliding_window_front_volume_24h.token0_volume
                cached_trade_volume_data["aggregated_token1_volume_24h"] += sliding_window_front_volume_24h.token1_volume
                cached_trade_volume_data["aggregated_token0_volume_usd_24h"] += sliding_window_front_volume_24h.token0_volume_usd
                cached_trade_volume_data["aggregated_token1_volume_usd_24h"] += sliding_window_front_volume_24h.token1_volume_usd
                # block height of trade volume
                block_height_trade_volume = int(
                    sliding_window_front_chain_24h[0]['data']['payload']['chainHeightRange']['end'])

                # store recent logs
                await store_recent_transactions_logs(writer_redis_conn, sliding_window_front_chain_24h, pair_contract_address)
            else:
                block_height_trade_volume = cached_trade_volume_data["processed_block_height_trade_volume"]

            if sliding_window_back_chain_7d:
                sliding_window_back_volume_7d:PairTradeVolume = calculate_pair_trade_volume(sliding_window_back_chain_7d)
                cached_trade_volume_data["aggregated_volume_7d"] -= sliding_window_back_volume_7d.total_volume
                cached_trade_volume_data["aggregated_token0_volume_7d"] -= sliding_window_back_volume_7d.token0_volume
                cached_trade_volume_data["aggregated_token1_volume_7d"] -= sliding_window_back_volume_7d.token1_volume
                cached_trade_volume_data["aggregated_token0_volume_usd_7d"] -= sliding_window_back_volume_7d.token0_volume_usd
                cached_trade_volume_data["aggregated_token1_volume_usd_7d"] -= sliding_window_back_volume_7d.token1_volume_usd

            if sliding_window_front_chain_7d:
                sliding_window_front_volume_7d:PairTradeVolume = calculate_pair_trade_volume(sliding_window_front_chain_7d)
                cached_trade_volume_data["aggregated_volume_7d"] += sliding_window_front_volume_7d.total_volume
                cached_trade_volume_data["aggregated_token0_volume_7d"] += sliding_window_front_volume_7d.token0_volume
                cached_trade_volume_data["aggregated_token1_volume_7d"] += sliding_window_front_volume_7d.token1_volume
                cached_trade_volume_data["aggregated_token0_volume_usd_7d"] += sliding_window_front_volume_7d.token0_volume_usd
                cached_trade_volume_data["aggregated_token1_volume_usd_7d"] += sliding_window_front_volume_7d.token1_volume_usd

            # if liquidity reserve block height is not found
            if pair_liquidity.block_height_total_reserve == 0:
                pair_liquidity.block_height_total_reserve = cached_trade_volume_data["processed_block_height_total_reserve"]
            if pair_liquidity.block_timestamp_total_reserve == 0:
                pair_liquidity.block_timestamp_total_reserve = cached_trade_volume_data["processed_block_timestamp_total_reserve"]

            pair_trade_volume_24h.total_volume = cached_trade_volume_data['aggregated_volume_24h']
            pair_trade_volume_7d.total_volume = cached_trade_volume_data['aggregated_volume_7d']
            pair_trade_volume_24h.fees = cached_trade_volume_data['aggregated_fees_24h']
            pair_trade_volume_24h.token0_volume = cached_trade_volume_data['aggregated_token0_volume_24h']
            pair_trade_volume_24h.token1_volume = cached_trade_volume_data['aggregated_token1_volume_24h']
            pair_trade_volume_7d.token0_volume = cached_trade_volume_data['aggregated_token0_volume_7d']
            pair_trade_volume_7d.token1_volume = cached_trade_volume_data['aggregated_token1_volume_7d']
            pair_trade_volume_24h.token0_volume_usd = cached_trade_volume_data['aggregated_token0_volume_usd_24h']
            pair_trade_volume_24h.token1_volume_usd = cached_trade_volume_data['aggregated_token1_volume_usd_24h']
            pair_trade_volume_7d.token0_volume_usd = cached_trade_volume_data['aggregated_token0_volume_usd_7d']
            pair_trade_volume_7d.token1_volume_usd = cached_trade_volume_data['aggregated_token1_volume_usd_7d']

        sliding_window_data = {
            "processed_tail_marker_24h": tail_marker_24h,
            "processed_head_marker_24h": head_marker_24h,
            "processed_tail_marker_7d": tail_marker_7d,
            "processed_head_marker_7d": head_marker_7d,
            "aggregated_volume_cid_24h": cids_volume_24h.dict(),
            "aggregated_volume_cid_7d": cids_volume_7d.dict(),
            "aggregated_volume_24h": pair_trade_volume_24h.total_volume,
            "aggregated_fees_24h": pair_trade_volume_24h.fees,
            "aggregated_token0_volume_24h": pair_trade_volume_24h.token0_volume,
            "aggregated_token1_volume_24h": pair_trade_volume_24h.token1_volume,
            "aggregated_token0_volume_usd_24h": pair_trade_volume_24h.token0_volume_usd,
            "aggregated_token1_volume_usd_24h": pair_trade_volume_24h.token1_volume_usd,
            "aggregated_volume_7d": pair_trade_volume_7d.total_volume,
            "aggregated_token0_volume_7d": pair_trade_volume_7d.token0_volume,
            "aggregated_token1_volume_7d": pair_trade_volume_7d.token1_volume,
            "aggregated_token0_volume_usd_7d": pair_trade_volume_7d.token0_volume_usd,
            "aggregated_token1_volume_usd_7d": pair_trade_volume_7d.token1_volume_usd,
            "processed_block_height_total_reserve": pair_liquidity.block_height_total_reserve,
            "processed_block_height_trade_volume": block_height_trade_volume,
            "processed_block_timestamp_total_reserve": pair_liquidity.block_timestamp_total_reserve,
        }
        prepared_snapshot = liquidityProcessedData(
            contractAddress=pair_contract_address,
            name=pair_token_metadata['pair']['symbol'],
            liquidity=f"US${round(abs(pair_liquidity.total_liquidity)):,}",
            volume_24h=f"US${round(abs(pair_trade_volume_24h.total_volume)):,}",
            volume_7d=f"US${round(abs(pair_trade_volume_7d.total_volume)):,}",
            fees_24h=f"US${round(abs(pair_trade_volume_24h.fees)):,}",
            cid_volume_24h=cids_volume_24h,
            cid_volume_7d=cids_volume_7d,
            block_height=pair_liquidity.block_height_total_reserve,
            block_timestamp=pair_liquidity.block_timestamp_total_reserve,
            token0Liquidity=pair_liquidity.token0_liquidity,
            token1Liquidity=pair_liquidity.token1_liquidity,
            token0LiquidityUSD=pair_liquidity.token0_liquidity_usd,
            token1LiquidityUSD=pair_liquidity.token1_liquidity_usd,
            token0TradeVolume_24h=pair_trade_volume_24h.token0_volume,
            token1TradeVolume_24h=pair_trade_volume_24h.token1_volume,
            token0TradeVolumeUSD_24h=pair_trade_volume_24h.token0_volume_usd,
            token1TradeVolumeUSD_24h=pair_trade_volume_24h.token1_volume_usd,
            token0TradeVolume_7d=pair_trade_volume_7d.token0_volume,
            token1TradeVolume_7d=pair_trade_volume_7d.token1_volume,
            token0TradeVolumeUSD_7d=pair_trade_volume_7d.token0_volume_usd,
            token1TradeVolumeUSD_7d=pair_trade_volume_7d.token1_volume_usd
        )

        await store_pair_daily_stats(writer_redis_conn, pair_contract_address, prepared_snapshot)

        logger.debug('Storing prepared trades, token reserves snapshot: %s', prepared_snapshot)
        logger.debug('Adding calculated snapshot to daily stats zset and pruning it for just 24h data')
        await writer_redis_conn.mset({
            redis_keys.get_uniswap_pair_contract_V2_pair_data(
                f"{Web3.toChecksumAddress(pair_contract_address)}"): prepared_snapshot.json(),
            redis_keys.get_uniswap_pair_cache_sliding_window_data(
                f"{Web3.toChecksumAddress(pair_contract_address)}"): json.dumps(sliding_window_data)
        })
        logger.debug(
            "Calculated v2 pair data for contract: %s | symbol:%s-%s",
             pair_contract_address,
             pair_token_metadata['token0']['symbol'],
             pair_token_metadata['token1']['symbol'])
        return prepared_snapshot
    except Exception as exc:
        logger.error('Error in process_pairs_trade_volume_and_reserves: %s', exc, exc_info=True)
        raise

async def v2_pairs_data(
        async_httpx_client: AsyncClient,
        ipfs_write_client: AsyncIPFSClient,
        ipfs_read_client: AsyncIPFSClient
):
    try:

        aioredis_pool = RedisPool()
        await aioredis_pool.populate()
        redis_conn: aioredis.Redis = aioredis_pool.writer_redis_pool

        if len(PAIR_CONTRACTS) <= 0:
            return []

        logger.debug("Create threads to aggregate trade volume and liquidity for pairs")
        process_data_list = []
        for pair_contract_address in PAIR_CONTRACTS:
            trade_volume_reserves = process_pairs_trade_volume_and_reserves(
                aioredis_pool.writer_redis_pool,
                pair_contract_address,
                ipfs_read_client
            )
            process_data_list.append(trade_volume_reserves)

        final_results: Iterable[Union[liquidityProcessedData, BaseException]] = await asyncio.gather(*process_data_list,
                                                                                                     return_exceptions=True)
        common_blockheight_reached = False
        common_block_timestamp = False
        pair_contract_address = None
        collected_heights = list(
            map(lambda x: x.block_height, filter(lambda y: isinstance(y, liquidityProcessedData), final_results)))
        collected_block_timestamp = list(
            map(lambda x: x.block_timestamp, filter(lambda y: isinstance(y, liquidityProcessedData), final_results)))
        pair_addresses = list(
            map(lambda x: x.contractAddress, filter(lambda y: isinstance(y, liquidityProcessedData), final_results)))
        none_results = list(
            map(lambda x: x, filter(lambda y: isinstance(y, type(None)), final_results)))
        # logger.debug('Final results from running v2 pairs coroutines: %s', final_results)
        # logger.debug('Filtered heights from running v2 pairs coroutines: %s', collected_heights)
        if cardinality.count(none_results) == cardinality.count(final_results):
            logger.debug("pair-contract chains didn't move ahead, sleeping till next cycle...")
        elif cardinality.count(final_results) != cardinality.count(collected_heights):
            if len(collected_heights) == 0:
                logger.error(
                    'Got empty result for all pairs, either pair chains didn\'t move ahead or there is an error while fetching dag-blocks data')
            else:
                logger.error(
                    'In V2 pairs overall summary snapshot creation: '
                    'No common block height found among all summarized contracts. Some results returned exception.'
                    'Block heights found: %s',collected_heights
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
        tentative_audit_project_block_height = 0
        if common_blockheight_reached:
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
                tentative_audit_project_block_height = await redis_conn.get(redis_keys.get_tentative_block_height_key(
                    project_id=redis_keys.get_uniswap_pairs_summary_snapshot_project_id()
                ))
                tentative_audit_project_block_height  = int(tentative_audit_project_block_height) if tentative_audit_project_block_height else 0

                logger.debug('Sending v2 pairs summary payload to audit protocol')
                # send to audit protocol for snapshot to be committed
                try:
                    response = await helper_functions.commit_payload(
                        project_id=redis_keys.get_uniswap_pairs_summary_snapshot_project_id(),
                        report_payload=summarized_payload,
                        session=async_httpx_client,
                        skipAnchorProof=False
                    )
                except Exception as exc:
                    logger.error(
                        'Error while committing pairs summary snapshot at block height %s to audit protocol. '
                        'Exception: %s',
                        common_blockheight_reached, exc, exc_info=True
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
                        updated_audit_project_block_height = tentative_audit_project_block_height + 1
        if wait_for_snapshot_project_new_commit:
            wait_cycles = 0
            while True:
                # introduce a break condition if something goes wrong and snapshot summary does not move ahead
                wait_cycles += 1
                if wait_cycles > 3:  # Wait for 60 seconds after which move ahead as something must have has gone wrong with snapshot summary submission
                    logger.info(
                        "Waited for %s cycles, snapshot summary project has not moved ahead. Stopped waiting to retry in next cycle.",
                        wait_cycles)
                    break
                logger.debug('Waiting for 10 seconds to check if latest v2 pairs summary snapshot was committed...')
                await asyncio.sleep(10)

                block_status = await retrieval_utils.retrieve_block_status(
                    project_id=redis_keys.get_uniswap_pairs_summary_snapshot_project_id(),
                    project_block_height=0,
                    block_height=updated_audit_project_block_height,
                    reader_redis_conn=redis_conn,
                    writer_redis_conn=redis_conn,
                    ipfs_read_client=ipfs_read_client
                )
                if block_status.status < 3:
                    continue
                logger.info(
                    'Audit project height against V2 pairs summary snapshot is %s | Moved from %s',
                    updated_audit_project_block_height, tentative_audit_project_block_height
                )
                begin_block_data = await get_oldest_block_and_timestamp(
                    pair_contract_address, redis_conn, ipfs_read_client
                )

                snapshot_zset_entry = uniswapPairsSnapshotZset(
                    cid=block_status.payload_cid,
                    txHash=block_status.tx_hash,
                    begin_block_height_24h=begin_block_data["begin_block_height_24h"],
                    begin_block_timestamp_24h=begin_block_data["begin_block_timestamp_24h"],
                    begin_block_height_7d=begin_block_data["begin_block_height_7d"],
                    begin_block_timestamp_7d=begin_block_data["begin_block_timestamp_7d"],
                    txStatus=block_status.status,
                    dagHeight=updated_audit_project_block_height
                )

                logger.debug("Prepared snapshot redis-zset entry: %s ", snapshot_zset_entry.json())

                # store in snapshots zset
                result = await asyncio.gather(
                    redis_conn.zadd(
                        name=redis_keys.get_uniswap_pair_snapshot_summary_zset(),
                        mapping={snapshot_zset_entry.json(): common_blockheight_reached}),
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
                        common_blockheight_reached),

                )
                logger.debug("Updated snapshot details in redis, ops results: %s", result)

                # prune zset
                block_height_zset_len = await redis_conn.zcard(name=redis_keys.get_uniswap_pair_snapshot_summary_zset())

                if block_height_zset_len > 20:
                    _ = await redis_conn.zremrangebyrank(
                        name=redis_keys.get_uniswap_pair_snapshot_summary_zset(),
                        min=0,
                        max=-1 * (block_height_zset_len - 20) + 1
                    )
                    logger.debug('Pruned snapshot summary CID zset by %s elements', _)

                pruning_timestamp = int(time()) - (60 * 60 * 24 + 60 * 60 * 2)  # now - 26 hours
                _ = await redis_conn.zremrangebyscore(
                    name=redis_keys.get_uniswap_pair_snapshot_timestamp_zset(),
                    min=0,
                    max=pruning_timestamp
                )

                logger.debug('Pruned snapshot summary Timestamp zset by %s elements', _)
                logger.info('V2 pairs summary snapshot updated...')
                break

        return final_results

    except Exception as exc:
        logger.error("Error at V2 pair data: %s", str(exc), exc_info=True)


if __name__ == '__main__':
    print("", "")
    # loop = asyncio.get_event_loop()
    # data = loop.run_until_complete(fetch_and_update_status_of_older_snapshots(
    #     summary_snapshots_zset_key=redis_keys.get_uniswap_pair_daily_stats_snapshot_zset(),
    #     summary_snapshots_project_id=redis_keys.get_uniswap_pairs_v2_daily_snapshot_project_id(),
    #     data_model=uniswapDailyStatsSnapshotZset,
    # ))
    # print("## ##", "")
    # print(data)
