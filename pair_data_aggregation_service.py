from typing import Union, Iterable
from time import time
from web3 import Web3
from httpx import AsyncClient
from redis import asyncio as aioredis
from async_ipfshttpclient.main import AsyncIPFSClient
from gnosis.eth import EthereumClient
from config import settings
from data_models import liquidityProcessedData, uniswapPairsSnapshotZset, DAGBlockRange, PairLiquidity, PairTradeVolume
from exceptions import MissingIndexException
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
from config import settings

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

NAMESPACE = settings.pooler_namespace
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
                ipfs_read_client=ipfs_read_client
            )
            # handle case of null payloads in dag_blocks
            if 'data' not in dag_block:
                dag_block = {}
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
            ipfs_read_client=ipfs_read_client
        ),
        get_dag_block_by_height(
            project_id=project_id_trade_volume,
            block_height=volume_tail_marker_7d,
            reader_redis_conn=redis_conn,
            ipfs_read_client=ipfs_read_client
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
                redis_keys.get_uniswap_pair_contract_tokens_data(
                    pair_address,
                    settings.pooler_namespace
                )
            )

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
                name=redis_keys.get_uniswap_pair_contract_tokens_data(
                    pair_address,
                    settings.pooler_namespace
                    ),
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
    latest_dag_block_for_liquidity = await get_dag_block_by_height(
        project_id=project_id_token_reserve,
        block_height=liquidity_head_marker,
        reader_redis_conn=writer_redis_conn,
        ipfs_read_client=ipfs_read_client
    )
    pair_liquidity: PairLiquidity = PairLiquidity()

    if not(latest_dag_block_for_liquidity is not None and latest_dag_block_for_liquidity.get('data') and latest_dag_block_for_liquidity['data'].get('payload')):
        logger.error(
            'Got DAG block at liquidity marker %s for pair %s as : %s',
            liquidity_head_marker, pair_contract_address, latest_dag_block_for_liquidity
        )
        return pair_liquidity
    pair_liquidity.token0_liquidity = float(list(latest_dag_block_for_liquidity['data']['payload']['token0Reserves'].values())[-1])
    pair_liquidity.token1_liquidity = float(list(latest_dag_block_for_liquidity['data']['payload']['token1Reserves'].values())[-1])
    pair_liquidity.token0_liquidity_usd = float(list(latest_dag_block_for_liquidity['data']['payload']['token0ReservesUSD'].values())[-1])
    pair_liquidity.token1_liquidity_usd = float(list(latest_dag_block_for_liquidity['data']['payload']['token1ReservesUSD'].values())[-1])
    pair_liquidity.total_liquidity = pair_liquidity.token0_liquidity_usd + pair_liquidity.token1_liquidity_usd
    pair_liquidity.block_height_total_reserve = int(latest_dag_block_for_liquidity['data']['payload']['chainHeightRange']['end'])
    pair_liquidity.block_timestamp_total_reserve = int(latest_dag_block_for_liquidity['data']['payload']['timestamp'])
    if not pair_liquidity.block_timestamp_total_reserve:
        pair_liquidity.block_timestamp_total_reserve = int(latest_dag_block_for_liquidity['timestamp'])

    return pair_liquidity


async def store_recent_transactions_logs(writer_redis_conn: aioredis.Redis, dag_chain, pair_contract_address):
    recent_logs = []
    for log in dag_chain:
        event_logs = log['data']['payload']['events']['Swap']['logs'] + log['data']['payload']['events']['Mint'][
            'logs'] + log['data']['payload']['events']['Burn']['logs']
        recent_logs.extend(event_logs)

    recent_logs = sorted(recent_logs, key=lambda log: log['blockNumber'], reverse=True)

    oldLogs = await writer_redis_conn.get(
                redis_keys.get_uniswap_pair_cached_recent_logs(
                    f"{Web3.toChecksumAddress(pair_contract_address)}",
                    settings.pooler_namespace
                )
            )
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
            redis_keys.get_uniswap_pair_cached_recent_logs(
                f"{Web3.toChecksumAddress(pair_contract_address)}",
                settings.pooler_namespace
            ),
        json.dumps(recent_logs))
    return recent_logs

async def store_pair_daily_stats(writer_redis_conn: aioredis.Redis, pair_contract_address, snapshot):
    now = int(time())
    await writer_redis_conn.zadd(
        name=redis_keys.get_uniswap_pair_cache_daily_stats(
            Web3.toChecksumAddress(pair_contract_address),
            settings.pooler_namespace
            ),
        mapping={json.dumps(snapshot.json()): now}
    )
    await writer_redis_conn.zremrangebyscore(
        name=redis_keys.get_uniswap_pair_cache_daily_stats(
            Web3.toChecksumAddress(pair_contract_address),
            settings.pooler_namespace
            ),
        min=float('-inf'),
        max=float(now - 60 * 60 * 25)
    )

async def process_pairs_trade_volume_and_reserves(
        writer_redis_conn: aioredis.Redis,
        common_epoch_end,
        pair_contract_address,
        ipfs_read_client: AsyncIPFSClient,
        httpx_client: AsyncClient
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
            redis_keys.get_uniswap_pair_cache_sliding_window_data(
                f"{Web3.toChecksumAddress(pair_contract_address)}",
                settings.pooler_namespace
            ),
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
            msg = 'Values of 24h head, tail and 7h head, tail marker - some of them might be null and hence returning |' 
            f'[tail_marker_24h, head_marker_24h,  tail_marker_7d, head_marker_7d] : {tail_marker_24h, head_marker_24h, tail_marker_7d, head_marker_7d}'
            logger.info(msg)
            raise MissingIndexException(msg)

        # if head of 24h and 7d index at 1 or less then return
        if head_marker_24h <= 1 or head_marker_7d <= 1:
            raise MissingIndexException('Values of 24 head and tail marker are not sufficient to calculate aggregates')

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
            alert_msg = f"Error retrieving data in head DAG block to mark liquidity against latest source chain block in epoch | pair_contract_address: {pair_contract_address} | "
            "Resuming next cycle"
            logger.warning(
                alert_msg
            )
            raise MissingIndexException(alert_msg)

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
            dag_chain_24h_intact_payloads = [x for x in dag_chain_24h if x is not None and x.get('data') and x['data'].get('payload')]
            dag_chain_24h_bad_payloads = [x for x in dag_chain_24h if x is None or not x.get('data') or not x['data'].get('payload')]
            dag_chain_7d_intact_payloads = [x for x in dag_chain_7d if x is not None and x.get('data') and x['data'].get('payload')]
            dag_chain_7d_bad_payloads = [x for x in dag_chain_7d if x is None or not x.get('data') or not x['data'].get('payload')]
            if dag_chain_24h_bad_payloads:
                logger.warning(
                    "Found %s bad payloads in 24h dag-chain | projectId: %s",
                    dag_chain_24h_bad_payloads,
                    project_id_trade_volume
                )
                slack_alert_msg = {
                    'severity': 'MEDIUM',
                    'Service': 'TradeVolumeProcessor',
                    'errorDetails': f"\nFound following DAG blocks with bad payloads in 24h dag-chain during first aggregation building | projectId: {project_id_trade_volume}\n"
                                    f"{dag_chain_24h_bad_payloads}"
                }
                try:
                    await httpx_client.post(
                        url=settings.primary_indexer.slack_notify_URL,
                        json=slack_alert_msg
                    )
                except Exception as e:
                    logger.error(
                        "Error while sending slack alert for bad payloads in 24h dag-chain during first aggregation building | projectId: %s | error: %s",
                        project_id_trade_volume,
                        e
                    )
            if dag_chain_7d_bad_payloads:
                logger.warning(
                    "Found %s bad payloads in 7d dag-chain | projectId: %s",
                    dag_chain_7d_bad_payloads,
                    project_id_trade_volume
                )
                slack_alert_msg = {
                    'severity': 'MEDIUM',
                    'Service': 'TradeVolumeProcessor',
                    'errorDetails': f"\nFound following DAG blocks with bad payloads in 7d dag-chain during first aggregation building | projectId: {project_id_trade_volume}\n"
                                    f"{dag_chain_7d_bad_payloads}"
                }
                try:
                    await httpx_client.post(
                        url=settings.primary_indexer.slack_notify_URL,
                        json=slack_alert_msg
                    )
                except Exception as e:
                    logger.error(
                        "Error while sending slack alert for bad payloads in 7d dag-chain during first aggregation building | projectId: %s | error: %s",
                        project_id_trade_volume,
                        e
                    )   
            # calculate trade volume 24h
            pair_trade_volume_24h = calculate_pair_trade_volume(dag_chain_24h_intact_payloads)
            # calculate trade volume 7d
            pair_trade_volume_7d = calculate_pair_trade_volume(dag_chain_7d_intact_payloads)
            # store last recent logs, these will be used to show recent transaction for perticular contract
            # using only 24h dag chain as we just need 75 log at max
            await store_recent_transactions_logs(writer_redis_conn, dag_chain_24h_intact_payloads, pair_contract_address)

            cids_volume_24h: DAGBlockRange = DAGBlockRange(
                head_block_cid= dag_chain_24h[0]['dagCid'],
                tail_block_cid=dag_chain_24h[-1]['dagCid']
            )
            cids_volume_7d: DAGBlockRange = DAGBlockRange(
                head_block_cid= dag_chain_7d[0]['dagCid'],
                tail_block_cid=dag_chain_7d[-1]['dagCid']
                )
            block_height_trade_volume = common_epoch_end
        else:
            sliding_window_front_chain_24h = []
            sliding_window_back_chain_24h = []
            sliding_window_front_chain_7d = []
            sliding_window_back_chain_7d = []

            logger.debug(
                "Attempting to fetch 24h sliding window, front: %s - %s, back: %s - %s | last seen range %s - %s | projectId: %s",
                cached_trade_volume_data['processed_head_marker_24h'] + 1, head_marker_24h,
                cached_trade_volume_data['processed_tail_marker_24h']+1, tail_marker_24h,
                cached_trade_volume_data['processed_head_marker_24h'], cached_trade_volume_data['processed_tail_marker_24h'],
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
                logger.debug(
                    "Fetched 24h sliding window head: %s - %s | last seen range %s - %s | projectId: %s",
                    cached_trade_volume_data['processed_head_marker_24h'] + 1, head_marker_24h,
                    cached_trade_volume_data['processed_head_marker_24h'], cached_trade_volume_data['processed_tail_marker_24h'],
                    project_id_trade_volume
                )
            else:
                logger.debug("24h head is at %d and did not move ahead for project %s | last seen range %s - %s",
                    head_marker_24h,
                    cached_trade_volume_data["processed_head_marker_24h"],cached_trade_volume_data["processed_tail_marker_24h"],
                    project_id_trade_volume
                )
                

            # if 24h tail moved ahead
            if tail_marker_24h > cached_trade_volume_data["processed_tail_marker_24h"]:
                # back of the chain where head=current_tail and tail=last_tail
                sliding_window_back_chain_24h = await get_dag_blocks_in_range(
                    project_id_trade_volume,
                    cached_trade_volume_data["processed_tail_marker_24h"]+1,
                    tail_marker_24h,
                    writer_redis_conn,
                    ipfs_read_client
                )
                logger.debug(
                    "Fetched 24h sliding window tail: %s - %s | last seen range %s - %s | projectId: %s",
                    cached_trade_volume_data['processed_tail_marker_24h'] + 1, tail_marker_24h,
                    cached_trade_volume_data['processed_head_marker_24h'], cached_trade_volume_data['processed_tail_marker_24h'],
                    project_id_trade_volume
                )
            logger.debug(
                "Attempting to fetch 7d sliding window, front: %s - %s, back: %s - %s | last seen range %s - %s | projectId: %s",
                cached_trade_volume_data['processed_head_marker_7d'] + 1, head_marker_7d,
                cached_trade_volume_data['processed_tail_marker_7d']+1, tail_marker_7d,
                cached_trade_volume_data['processed_head_marker_7d'], cached_trade_volume_data['processed_tail_marker_7d'],
                project_id_trade_volume
            )

            # if 7d head moved ahead
            if head_marker_7d > cached_trade_volume_data["processed_head_marker_7d"]:
                sliding_window_front_chain_7d = await get_dag_blocks_in_range(
                    project_id_trade_volume,
                    cached_trade_volume_data["processed_head_marker_7d"] + 1,
                    head_marker_7d,
                    writer_redis_conn,
                    ipfs_read_client
                )
                logger.debug(
                    "Fetched 7d sliding window head: %s - %s | last seen range %s - %s | projectId: %s",
                    cached_trade_volume_data['processed_head_marker_7d'] + 1, head_marker_7d,
                    cached_trade_volume_data['processed_head_marker_7d'], cached_trade_volume_data['processed_tail_marker_7d'],
                    project_id_trade_volume
                )
            else:
                logger.debug("7d head is at %d and did not move ahead for project %s | last seen range %s - %s",
                    head_marker_7d,
                    project_id_trade_volume,
                    cached_trade_volume_data["processed_head_marker_7d"], cached_trade_volume_data["processed_tail_marker_7d"]
                    
                )

            # if 7d tail moved ahead
            if tail_marker_7d > cached_trade_volume_data["processed_tail_marker_7d"]:
                sliding_window_back_chain_7d = await get_dag_blocks_in_range(
                    project_id_trade_volume,
                    cached_trade_volume_data["processed_tail_marker_7d"]+1,
                    tail_marker_7d,
                    writer_redis_conn,
                    ipfs_read_client
                )
                logger.debug(
                    "Fetched 7d sliding window tail: %s - %s | last seen range %s - %s | projectId: %s",
                    cached_trade_volume_data['processed_tail_marker_7d'] + 1, tail_marker_7d,
                    cached_trade_volume_data['processed_head_marker_7d'], cached_trade_volume_data['processed_tail_marker_7d'],
                    project_id_trade_volume
                )
            else:
                logger.debug("7d tail is at %d and did not move ahead for project %s | last seen range %s - %s",
                    tail_marker_7d,
                    cached_trade_volume_data["processed_head_marker_7d"],cached_trade_volume_data["processed_tail_marker_7d"],
                    project_id_trade_volume
                )


            if (not sliding_window_front_chain_24h and not sliding_window_back_chain_24h) or (not sliding_window_front_chain_7d and not sliding_window_back_chain_7d):
                logger.debug("No new sliding window to process for project %s, returning last cached reserves and trade volume data", project_id_trade_volume)    
                prepared_snapshot = liquidityProcessedData(
                    contractAddress=pair_contract_address,
                    name=pair_token_metadata['pair']['symbol'],
                    liquidity=f"US${round(abs(pair_liquidity.total_liquidity)):,}",
                    volume_24h=f"US${round(abs(cached_trade_volume_data['aggregated_volume_24h'])):,}",
                    volume_7d=f"US${round(abs(cached_trade_volume_data['aggregated_volume_7d'])):,}",
                    fees_24h=f"US${round(abs(cached_trade_volume_data['aggregated_fees_24h'])):,}",
                    cid_volume_24h=DAGBlockRange(head_block_cid=cached_trade_volume_data['aggregated_volume_cid_24h']['head_block_cid'], tail_block_cid=cached_trade_volume_data['aggregated_volume_cid_24h']['tail_block_cid']),
                    cid_volume_7d=DAGBlockRange(head_block_cid=cached_trade_volume_data['aggregated_volume_cid_7d']['head_block_cid'], tail_block_cid=cached_trade_volume_data['aggregated_volume_cid_7d']['tail_block_cid']),
                    block_height=pair_liquidity.block_height_total_reserve,
                    block_timestamp=pair_liquidity.block_timestamp_total_reserve,
                    token0Liquidity=pair_liquidity.token0_liquidity,
                    token1Liquidity=pair_liquidity.token1_liquidity,
                    token0LiquidityUSD=pair_liquidity.token0_liquidity_usd,
                    token1LiquidityUSD=pair_liquidity.token1_liquidity_usd,
                    token0TradeVolume_24h=cached_trade_volume_data['aggregated_token0_volume_24h'],
                    token1TradeVolume_24h=cached_trade_volume_data['aggregated_token1_volume_24h'],
                    token0TradeVolumeUSD_24h=cached_trade_volume_data['aggregated_token0_volume_usd_24h'],
                    token1TradeVolumeUSD_24h=cached_trade_volume_data['aggregated_token1_volume_usd_24h'],
                    token0TradeVolume_7d=cached_trade_volume_data['aggregated_token0_volume_7d'],
                    token1TradeVolume_7d=cached_trade_volume_data['aggregated_token1_volume_7d'],
                    token0TradeVolumeUSD_7d=cached_trade_volume_data['aggregated_token0_volume_usd_7d'],
                    token1TradeVolumeUSD_7d=cached_trade_volume_data['aggregated_token1_volume_usd_7d']
                )
                return prepared_snapshot

            
            cids_volume_24h.head_block_cid = sliding_window_front_chain_24h[-1]['dagCid']
            
            if sliding_window_back_chain_24h:
                cids_volume_24h.tail_block_cid = sliding_window_back_chain_24h[-1]['dagCid']
            else:
                # this is to deal with the case when collected snapshots have not reached 24 hours yet
                tail_block_24h = await get_dag_block_by_height(
                    project_id=project_id_trade_volume,
                    block_height=tail_marker_24h,
                    reader_redis_conn=writer_redis_conn,
                    ipfs_read_client=ipfs_read_client
                )
                cids_volume_24h.tail_block_cid = tail_block_24h['dagCid']
            cids_volume_7d.head_block_cid = sliding_window_front_chain_7d[-1]['dagCid']
            if sliding_window_back_chain_7d:
                cids_volume_7d.tail_block_cid = sliding_window_back_chain_7d[-1]['dagCid']
            else:
                tail_block_7d = await get_dag_block_by_height(
                    project_id=project_id_trade_volume,
                    block_height=tail_marker_7d,
                    reader_redis_conn=writer_redis_conn,
                    ipfs_read_client=ipfs_read_client
                )
                cids_volume_7d.tail_block_cid = tail_block_7d['dagCid']

            if sliding_window_back_chain_24h:
                qualified_sliding_window_blocks_back_chain_24h = [x for x in sliding_window_back_chain_24h if x is not None and x.get('data') and x['data'].get('payload')]
                unqualified_sliding_window_blocks_back_chain_24h = [x for x in sliding_window_back_chain_24h if x is None or x.get('data') is None or x['data'].get('payload') is None]
                if unqualified_sliding_window_blocks_back_chain_24h:
                    slack_alert_msg = {
                        'severity': 'MEDIUM',
                        'Service': 'TradeVolumeProcessor',
                        'errorDetails': f'\nFound unqualified blocks in sliding window back chain 24h for project {project_id_trade_volume} '
                                        f'while adjusting tail from {cached_trade_volume_data["processed_tail_marker_24h"]+1} to {tail_marker_24h}.\n\n```{unqualified_sliding_window_blocks_back_chain_24h}```' 
                    }
                    try:
                        await httpx_client.post(
                            url=settings.primary_indexer.slack_notify_URL,
                            json=slack_alert_msg
                        )
                    except Exception as e:
                        logger.error(
                            'Failed to send slack alert for unqualified blocks in sliding window back chain 24h for project %s while adjusting tail from %s - %s: %s',
                            project_id_trade_volume,
                            cached_trade_volume_data['processed_tail_marker_24h'] + 1, tail_marker_24h,
                            e
                        )
                    logger.warning(slack_alert_msg['errorDetails'])
                sliding_window_back_volume_24h:PairTradeVolume = calculate_pair_trade_volume(qualified_sliding_window_blocks_back_chain_24h)
                logger.info(
                    'Calculated sliding window back volume 24h for project %s, while adjusting tail from %s - %s: %s',
                    project_id_trade_volume, cached_trade_volume_data['processed_tail_marker_24h'] + 1, tail_marker_24h, sliding_window_back_volume_24h
                    
                )
                cached_trade_volume_data["aggregated_volume_24h"] -= sliding_window_back_volume_24h.total_volume
                cached_trade_volume_data["aggregated_fees_24h"] -= sliding_window_back_volume_24h.fees
                cached_trade_volume_data["aggregated_token0_volume_24h"] -= sliding_window_back_volume_24h.token0_volume
                cached_trade_volume_data["aggregated_token1_volume_24h"] -= sliding_window_back_volume_24h.token1_volume
                cached_trade_volume_data["aggregated_token0_volume_usd_24h"] -= sliding_window_back_volume_24h.token0_volume_usd
                cached_trade_volume_data["aggregated_token1_volume_usd_24h"] -= sliding_window_back_volume_24h.token1_volume_usd

            if sliding_window_front_chain_24h:
                qualified_sliding_window_blocks_front_chain_24h = [x for x in sliding_window_back_chain_24h if x is not None and x.get('data') and x['data'].get('payload')]
                unqualified_sliding_window_blocks_front_chain_24h = [x for x in sliding_window_back_chain_24h if x is None or x.get('data') is None or x['data'].get('payload') is None]
                if unqualified_sliding_window_blocks_front_chain_24h:
                    slack_alert_msg = {
                        'severity': 'MEDIUM',
                        'Service': 'TradeVolumeProcessor',
                        'errorDetails': f'\nFound unqualified blocks in sliding window front chain 24h for project {project_id_trade_volume} '
                                        f'while adjusting tail from {cached_trade_volume_data["processed_head_marker_24h"]+1} to {head_marker_24h}\n\n```{unqualified_sliding_window_blocks_front_chain_24h}```' 
                    }
                    try:
                        await httpx_client.post(
                            url=settings.primary_indexer.slack_notify_URL,
                            json=slack_alert_msg
                        )
                    except Exception as e:
                        logger.error(
                            'Failed to send slack alert for unqualified blocks in sliding window front chain 24h for project %s while adjusting head from %s - %s: %s',
                            project_id_trade_volume,
                            cached_trade_volume_data['processed_head_marker_24h'] + 1, head_marker_24h,
                            e
                        )
                    logger.warning(slack_alert_msg['errorDetails'])
                sliding_window_front_volume_24h:PairTradeVolume =  calculate_pair_trade_volume(qualified_sliding_window_blocks_front_chain_24h)
                logger.info(
                    'Calculated sliding window front volume 24h for project %s, while adjusting head from %s - %s: %s',
                    project_id_trade_volume, cached_trade_volume_data['processed_head_marker_24h'] + 1, head_marker_24h, sliding_window_front_volume_24h
                    
                )
                cached_trade_volume_data["aggregated_volume_24h"] += sliding_window_front_volume_24h.total_volume
                cached_trade_volume_data["aggregated_fees_24h"] += sliding_window_front_volume_24h.fees
                cached_trade_volume_data["aggregated_token0_volume_24h"] += sliding_window_front_volume_24h.token0_volume
                cached_trade_volume_data["aggregated_token1_volume_24h"] += sliding_window_front_volume_24h.token1_volume
                cached_trade_volume_data["aggregated_token0_volume_usd_24h"] += sliding_window_front_volume_24h.token0_volume_usd
                cached_trade_volume_data["aggregated_token1_volume_usd_24h"] += sliding_window_front_volume_24h.token1_volume_usd
                # block height of trade volume
                block_height_trade_volume = common_epoch_end
                # store recent logs
                await store_recent_transactions_logs(writer_redis_conn, qualified_sliding_window_blocks_front_chain_24h, pair_contract_address)
            else:
                block_height_trade_volume = cached_trade_volume_data["processed_block_height_trade_volume"]

            if sliding_window_back_chain_7d:
                qualified_sliding_window_blocks_back_chain_7d = [x for x in sliding_window_back_chain_7d if x is not None and x.get('data') and x['data'].get('payload')]
                unqualified_sliding_window_blocks_back_chain_7d = [x for x in sliding_window_back_chain_7d if x is None or x.get('data') is None or x['data'].get('payload') is None]
                if unqualified_sliding_window_blocks_back_chain_7d:
                    slack_alert_msg = {
                        'severity': 'MEDIUM',
                        'Service': 'TradeVolumeProcessor',
                        'errorDetails': f'\nFound unqualified blocks in sliding window back chain 7d for project {project_id_trade_volume} '
                                        f'while adjusting tail from {cached_trade_volume_data["processed_tail_marker_7d"]+1} to {tail_marker_7d}\n\n```{unqualified_sliding_window_blocks_back_chain_7d}```' 
                    }
                    try:
                        await httpx_client.post(
                            url=settings.primary_indexer.slack_notify_URL,
                            json=slack_alert_msg
                        )
                    except Exception as e:
                        logger.error(
                            'Failed to send slack alert for unqualified blocks in sliding window back chain 7d for project %s while adjusting tail from %s - %s: %s',
                            project_id_trade_volume,
                            cached_trade_volume_data['processed_tail_marker_7d'] + 1, tail_marker_7d,
                            e
                        )
                    logger.warning(slack_alert_msg['errorDetails'])
                sliding_window_back_volume_7d:PairTradeVolume = calculate_pair_trade_volume(qualified_sliding_window_blocks_back_chain_7d)
                logger.info(
                    'Calculated sliding window back volume 7d for project %s, while adjusting tail from %s - %s: %s',
                    project_id_trade_volume, cached_trade_volume_data['processed_tail_marker_7d'] + 1, tail_marker_7d, sliding_window_back_volume_7d
                )
                cached_trade_volume_data["aggregated_volume_7d"] -= sliding_window_back_volume_7d.total_volume
                cached_trade_volume_data["aggregated_token0_volume_7d"] -= sliding_window_back_volume_7d.token0_volume
                cached_trade_volume_data["aggregated_token1_volume_7d"] -= sliding_window_back_volume_7d.token1_volume
                cached_trade_volume_data["aggregated_token0_volume_usd_7d"] -= sliding_window_back_volume_7d.token0_volume_usd
                cached_trade_volume_data["aggregated_token1_volume_usd_7d"] -= sliding_window_back_volume_7d.token1_volume_usd

            if sliding_window_front_chain_7d:
                qualified_sliding_window_blocks_front_chain_7d = [x for x in sliding_window_front_chain_7d if x is not None and x.get('data') and x['data'].get('payload')]
                unqualified_sliding_window_blocks_front_chain_7d = [x for x in sliding_window_front_chain_7d if x is None or x.get('data') is None or x['data'].get('payload') is None]
                if unqualified_sliding_window_blocks_front_chain_7d:
                    slack_alert_msg = {
                        'severity': 'MEDIUM',
                        'Service': 'TradeVolumeProcessor',
                        'errorDetails': f'\nFound unqualified blocks in sliding window front chain 7d for project {project_id_trade_volume} '
                                        f'while adjusting head from {cached_trade_volume_data["processed_head_marker_7d"]+1} to {head_marker_7d}\n\n```{unqualified_sliding_window_blocks_front_chain_7d}```' 
                    }
                    try:
                        await httpx_client.post(
                            url=settings.primary_indexer.slack_notify_URL,
                            json=slack_alert_msg
                        )
                    except Exception as e:
                        logger.error(
                            'Failed to send slack alert for unqualified blocks in sliding window front chain 7d for project %s while adjusting head from %s - %s: %s',
                            project_id_trade_volume,
                            cached_trade_volume_data['processed_head_marker_7d'] + 1, head_marker_7d,
                            e
                        )
                    logger.warning(slack_alert_msg['errorDetails'])
                sliding_window_front_volume_7d:PairTradeVolume = calculate_pair_trade_volume(qualified_sliding_window_blocks_front_chain_7d)
                logger.info(
                    'Calculated sliding window front volume 7d for project %s, while adjusting head from %s - %s: %s',
                    project_id_trade_volume, cached_trade_volume_data['processed_head_marker_7d'] + 1, head_marker_7d, sliding_window_front_volume_7d
                )
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

        # TODO: create data model for sliding window data
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
                f"{Web3.toChecksumAddress(pair_contract_address)}", settings.pooler_namespace): prepared_snapshot.json(),
            redis_keys.get_uniswap_pair_cache_sliding_window_data(
                f"{Web3.toChecksumAddress(pair_contract_address)}", settings.pooler_namespace): json.dumps(sliding_window_data)
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
        common_epoch_end,
        ipfs_write_client: AsyncIPFSClient,
        ipfs_read_client: AsyncIPFSClient
):
    aioredis_pool = RedisPool()
    await aioredis_pool.populate()
    redis_conn: aioredis.Redis = aioredis_pool.writer_redis_pool

    if len(PAIR_CONTRACTS) <= 0:
        return []

    logger.debug("Launching tasks to aggregate trade volume and liquidity for pairs")
    process_data_list = []
    for pair_contract_address in PAIR_CONTRACTS:
        trade_volume_reserves = process_pairs_trade_volume_and_reserves(
            aioredis_pool.writer_redis_pool,
            common_epoch_end,
            pair_contract_address,
            ipfs_read_client,
            async_httpx_client
        )
        process_data_list.append(trade_volume_reserves)

    final_results: Iterable[Union[liquidityProcessedData, None, BaseException]] = await asyncio.gather(
        *process_data_list, return_exceptions=True
    )

    final_results_map = dict(zip(PAIR_CONTRACTS, final_results))
    results_without_missing_index = {k: v for k, v in final_results_map.items() if not isinstance(v, MissingIndexException)}
    contracts_results_with_missing_index = {k for k, v in final_results_map.items() if isinstance(v, MissingIndexException)}
    # filter out exceptions around missing indexes
    if all(map(lambda x: isinstance(x, liquidityProcessedData), results_without_missing_index.items())):
        logger.error(
            'Aggregation on all pair contracts\' trade volume and reserves incomplete. Waiting to build v2 pairs summary in next cycle.\nErrors: %s',
            {k: v for k, v in results_without_missing_index.items() if not isinstance(v, liquidityProcessedData)}
        )
        return
    
    
    common_block_timestamp = max(
        filter(lambda x: isinstance(x, liquidityProcessedData) ,results_without_missing_index.values()), 
        key=lambda x: x.block_timestamp).block_timestamp
    
    pair_addresses = map(lambda x: x.contractAddress, filter(lambda y: isinstance(y, liquidityProcessedData), final_results))
    pair_contract_address = next(pair_addresses)
    logger.debug('Setting common blockheight for v2 pairs aggregation %s', common_epoch_end)
    
    summarized_result_l = [x.dict() for x in results_without_missing_index.values()]
    summarized_payload = {'data': summarized_result_l}
    # get last processed v2 pairs snapshot height
    l_ = await redis_conn.get(
        redis_keys.get_uniswap_pair_snapshot_last_block_height(
            settings.pooler_namespace
            )
        )
    if l_:
        last_common_block_height = int(l_)
    else:
        last_common_block_height = 0
    if common_epoch_end > last_common_block_height:
        logger.info(
            'Present common block height in V2 pairs summary snapshot is %s | Moved from %s',
            common_epoch_end, last_common_block_height
        )
        tentative_audit_project_block_height = await redis_conn.get(redis_keys.get_tentative_block_height_key(
            project_id=redis_keys.get_uniswap_pairs_summary_snapshot_project_id(settings.pooler_namespace)
        ))
        tentative_audit_project_block_height  = int(tentative_audit_project_block_height) if tentative_audit_project_block_height else 0

        logger.debug('Sending v2 pairs summary payload to audit protocol')
        # send to audit protocol for snapshot to be committed
        try:
            response = await helper_functions.commit_payload(
                project_id=redis_keys.get_uniswap_pairs_summary_snapshot_project_id(settings.pooler_namespace),
                report_payload=summarized_payload,
                session=async_httpx_client,
                skipAnchorProof=settings.txn_config.skip_summary_projects_anchor_proof
            )
        except Exception as exc:
            logger.error(
                'Error while committing pairs summary snapshot at common epoch end %s to audit protocol. '
                'Exception: %s',
                common_epoch_end, exc, exc_info=True
            )
        else:
            if 'message' in response.keys():
                logger.error(
                    'Error while committing pairs summary snapshot at common epoch end %s to audit protocol. '
                    'Response status code and other information: %s',
                    common_epoch_end, response
                )
            else:
                updated_audit_project_block_height = tentative_audit_project_block_height + 1
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
                        project_id=redis_keys.get_uniswap_pairs_summary_snapshot_project_id(settings.pooler_namespace),
                        project_block_height=0,
                        block_height=updated_audit_project_block_height,
                        reader_redis_conn=redis_conn,
                        writer_redis_conn=redis_conn,
                        ipfs_read_client=ipfs_read_client
                    )
                    if block_status is None:
                        logger.error("block_status is returned as None at height %s for project %s",
                        updated_audit_project_block_height, redis_keys.get_uniswap_pairs_summary_snapshot_project_id(settings.pooler_namespace))
                        break
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
                            name=redis_keys.get_uniswap_pair_snapshot_summary_zset(settings.pooler_namespace),
                            mapping={snapshot_zset_entry.json(): common_epoch_end}
                        ),
                        redis_conn.zadd(
                            name=redis_keys.get_uniswap_pair_snapshot_timestamp_zset(settings.pooler_namespace),
                            mapping={block_status.payload_cid: common_block_timestamp}
                        ),
                        redis_conn.set(
                            name=redis_keys.get_uniswap_pair_snapshot_payload_at_blockheight(
                                common_epoch_end,
                                settings.pooler_namespace
                                ),
                            value=json.dumps(summarized_payload),
                            ex=1800  # TTL of 30 minutes?
                        ),
                        redis_conn.set(
                            redis_keys.get_uniswap_pair_snapshot_last_block_height(settings.pooler_namespace),
                            common_epoch_end),
                    )
                    logger.debug("Updated snapshot details in redis, ops results: %s", result)

                    # prune zset
                    block_height_zset_len = await redis_conn.zcard(
                        name=redis_keys.get_uniswap_pair_snapshot_summary_zset(
                            settings.pooler_namespace
                            )
                        )

                    if block_height_zset_len > 20:
                        _ = await redis_conn.zremrangebyrank(
                            name=redis_keys.get_uniswap_pair_snapshot_summary_zset(settings.pooler_namespace),
                            min=0,
                            max=-1 * (block_height_zset_len - 20) + 1
                        )
                        logger.debug('Pruned snapshot summary CID zset by %s elements', _)

                    pruning_timestamp = int(time()) - (60 * 60 * 24 + 60 * 60 * 2)  # now - 26 hours
                    _ = await redis_conn.zremrangebyscore(
                        name=redis_keys.get_uniswap_pair_snapshot_timestamp_zset(settings.pooler_namespace),
                        min=0,
                        max=pruning_timestamp
                    )

                    logger.debug('Pruned snapshot summary Timestamp zset by %s elements', _)
                    logger.info('V2 pairs summary snapshot updated...')
                    break

                return results_without_missing_index

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
