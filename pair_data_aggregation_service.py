from functools import partial
from time import time

import aioredis
from web3 import Web3

from config import settings
from data_models import liquidityProcessedData
from utils import helper_functions
from utils import redis_keys
from utils.ipfs_async import client as ipfs_client
from utils.redis_conn import RedisPool
from utils.retrieval_utils import retrieve_block_data
import asyncio
import json
import logging.config
import os
import sys

logger = logging.getLogger('AuditProtocol|PairDataAggregation')
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

#TODO: put rpc url in settings
w3 = Web3(Web3.HTTPProvider("https://rpc-eth-arch.blockvigil.com/v1/2b3e1495cc8d5f27178974aab3f815c24ad4e201"))
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

async def get_maker_pair_data(prop):
    prop = prop.lower()
    if prop.lower() == "name":
        return "Maker"
    elif prop.lower() == "symbol":
        return "MKR"
    else:
        return "Maker"

async def get_pair_tokens_metadata(pair_contract_obj, pair_address, loop, writer_redis_conn):
    """
        returns information on the tokens contained within a pair contract - name, symbol, decimals of token0 and token1
        also returns pair symbol by concatenating {token0Symbol}-{token1Symbol}
    """
    pair_address = Web3.toChecksumAddress(pair_address)

    pairTokensAddresses = await writer_redis_conn.hgetall(redis_keys.get_uniswap_pair_contract_tokens_addresses(pair_address))
    if pairTokensAddresses:
        token0Addr = Web3.toChecksumAddress(pairTokensAddresses[b"token0Addr"].decode('utf-8'))
        token1Addr = Web3.toChecksumAddress(pairTokensAddresses[b"token1Addr"].decode('utf-8'))
    else:
        # run in loop's default executor
        pfunc_0 = partial(pair_contract_obj.functions.token0().call)
        token0Addr = await loop.run_in_executor(func=pfunc_0, executor=None)
        pfunc_1 = partial(pair_contract_obj.functions.token1().call)
        token1Addr = await loop.run_in_executor(func=pfunc_1, executor=None)
        token0Addr = Web3.toChecksumAddress(token0Addr)
        token1Addr = Web3.toChecksumAddress(token1Addr)
        await writer_redis_conn.hmset(redis_keys.get_uniswap_pair_contract_tokens_addresses(pair_address), 'token0Addr', token0Addr,
                               'token1Addr', token1Addr)
    # token0 contract
    token0 = w3.eth.contract(
        address=Web3.toChecksumAddress(token0Addr),
        abi=erc20_abi
    )
    # token1 contract
    token1 = w3.eth.contract(
        address=Web3.toChecksumAddress(token1Addr),
        abi=erc20_abi
    )
    pair_tokens_data = await writer_redis_conn.hgetall(redis_keys.get_uniswap_pair_contract_tokens_data(pair_address))
    
    logger.debug(f"Fetch token metadata, token0:{token0Addr} token1:{token1Addr}")
    
    if pair_tokens_data:
        token0_decimals = pair_tokens_data[b"token0_decimals"].decode('utf-8')
        token1_decimals = pair_tokens_data[b"token1_decimals"].decode('utf-8')
        token0_symbol = pair_tokens_data[b"token0_symbol"].decode('utf-8')
        token1_symbol = pair_tokens_data[b"token1_symbol"].decode('utf-8')
        token0_name = pair_tokens_data[b"token0_name"].decode('utf-8')
        token1_name = pair_tokens_data[b"token1_name"].decode('utf-8')
    else:
        executor_gather = list()
        if(Web3.toChecksumAddress(settings.contract_addresses.MAKER) == Web3.toChecksumAddress(token0Addr)):
            executor_gather.append(get_maker_pair_data('name'))
            executor_gather.append(get_maker_pair_data('symbol'))
        else:
            executor_gather.append(loop.run_in_executor(func=token0.functions.name().call, executor=None))
            executor_gather.append(loop.run_in_executor(func=token0.functions.symbol().call, executor=None))
        executor_gather.append(loop.run_in_executor(func=token0.functions.decimals().call, executor=None))


        if(Web3.toChecksumAddress(settings.contract_addresses.MAKER) == Web3.toChecksumAddress(token1Addr)):
            executor_gather.append(get_maker_pair_data('name'))
            executor_gather.append(get_maker_pair_data('symbol'))
        else:
            executor_gather.append(loop.run_in_executor(func=token1.functions.name().call, executor=None))
            executor_gather.append(loop.run_in_executor(func=token1.functions.symbol().call, executor=None))
        executor_gather.append(loop.run_in_executor(func=token1.functions.decimals().call, executor=None))

        [
            token0_name, token0_symbol, token0_decimals,
            token1_name, token1_symbol, token1_decimals
        ] = await asyncio.gather(*executor_gather)

        await writer_redis_conn.hmset(
            redis_keys.get_uniswap_pair_contract_tokens_data(pair_address),
            "token0_name", token0_name,
            "token0_symbol", token0_symbol,
            "token0_decimals", token0_decimals,
            "token1_name", token1_name,
            "token1_symbol", token1_symbol,
            "token1_decimals", token1_decimals,
            "pair_symbol", f"{token0_symbol}-{token1_symbol}"
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

async def get_pair_metadata_and_tokens_price(writer_redis_conn: aioredis.Redis, pair_contract_address):
    pair_contract_obj = w3.eth.contract(
        address=Web3.toChecksumAddress(pair_contract_address),
        abi=pair_contract_abi
    )

    pair_token_metadata = await get_pair_tokens_metadata(
        pair_contract_obj=pair_contract_obj,
        pair_address=Web3.toChecksumAddress(pair_contract_address),
        loop=asyncio.get_event_loop(),
        writer_redis_conn=writer_redis_conn
    )

    # get tokens prices:
    token0Price = await writer_redis_conn.get(
        redis_keys.get_uniswap_pair_cached_token_price(f"{pair_token_metadata['token0']['symbol']}-USDT"))
    if (token0Price):
        token0Price = float(token0Price.decode('utf-8'))
    else:
        token0Price = 0
        logger.error(
            f"Error: can't find {pair_token_metadata['token0']['symbol']}-USDT Price and setting it 0 | {pair_token_metadata['token0']['address']}")

    token1Price = await writer_redis_conn.get(
        redis_keys.get_uniswap_pair_cached_token_price(f"{pair_token_metadata['token1']['symbol']}-USDT"))
    if (token1Price):
        token1Price = float(token1Price.decode('utf-8'))
    else:
        token1Price = 0
        logger.error(
            f"Error: can't find {pair_token_metadata['token1']['symbol']}-USDT Price and setting it 0 | {pair_token_metadata['token1']['address']}")        
    
    return [pair_token_metadata, token0Price, token1Price]

def calculate_pair_trade_volume(dag_chain):
    # calculate / sum trade volume
    total_volume = sum(map(lambda x: x['data']['payload']['totalTrade'], dag_chain))

    # calculate / sum fee
    fees = sum(map(lambda x: x['data']['payload'].get('totalFee', 0), dag_chain))

    # get volume for individual tokens (in native token decimals):
    token0_volume = sum(map(lambda x: x['data']['payload']['token0TradeVolume'], dag_chain))
    token1_volume = sum(map(lambda x: x['data']['payload']['token1TradeVolume'], dag_chain))

    return [total_volume, fees, token0_volume, token1_volume]

async def calculate_pair_liquidity(writer_redis_conn: aioredis.Redis, pair_contract_address, token0Price, token1Price):
    project_id_token_reserve = f'uniswap_pairContract_pair_total_reserves_{pair_contract_address}_UNISWAPV2'

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
    total_liquidity = token0_liquidity * token0Price + token1_liquidity * token1Price
    block_height_total_reserve = int(liquidity_data['data']['payload']['chainHeightRange']['end'])

    return [total_liquidity, token0_liquidity, token1_liquidity, block_height_total_reserve]

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
        project_id_trade_volume = f'uniswap_pairContract_trade_volume_{pair_contract_address}_UNISWAPV2'
        project_id_token_reserve = f'uniswap_pairContract_pair_total_reserves_{pair_contract_address}_UNISWAPV2'

        # get head, tail and sliding window data from redis
        [
            cached_trade_volume_data, 
            tail_marker_24h, 
            head_marker_24h
        ] = await writer_redis_conn.mget([
            redis_keys.get_uniswap_pair_cache_sliding_window_data(f"{Web3.toChecksumAddress(pair_contract_address)}"),
            redis_keys.get_sliding_window_cache_tail_marker(project_id_trade_volume, '24h'),
            redis_keys.get_sliding_window_cache_head_marker(project_id_trade_volume, '24h')
        ])

        cached_trade_volume_data = json.loads(cached_trade_volume_data.decode('utf-8')) if cached_trade_volume_data else {}
        tail_marker_24h = int(tail_marker_24h.decode('utf-8')) if tail_marker_24h else 0
        head_marker_24h = int(head_marker_24h.decode('utf-8')) if head_marker_24h else 0        

        if not tail_marker_24h or not head_marker_24h:
            logger.error(f"Avoding pair data calculations as head or tail is empty for pair:{pair_contract_address}, tail:{tail_marker_24h}, head:{head_marker_24h}")
            return

        # initialize returning variables
        total_liquidity = 0
        token0_liquidity = 0
        token1_liquidity = 0
        total_volume_24h = 0
        token0_volume_24h = 0
        token1_volume_24h = 0
        fees_24h = 0
        cids_volume_24h = ''
        recent_logs = list()
        block_height_total_reserve = 0
        block_height_trade_volume = 0
        pair_token_metadata = {}

        [
            pair_token_metadata, 
            token0Price, 
            token1Price
        ] = await get_pair_metadata_and_tokens_price(writer_redis_conn, pair_contract_address)

        [
            total_liquidity, 
            token0_liquidity, 
            token1_liquidity, 
            block_height_total_reserve
        ] = await calculate_pair_liquidity(writer_redis_conn, pair_contract_address, token0Price, token1Price)
        if not total_liquidity:
            logger.error(f"Avoiding pair data calculation as liquidity data is not available - projectId:{project_id_token_reserve}")
            return

        if not cached_trade_volume_data:
            dag_chain_24h = await get_dag_blocks_in_range(project_id_trade_volume, tail_marker_24h, head_marker_24h, writer_redis_conn)
            if not dag_chain_24h:
                logger.error(f"dag_chain_24h array is empty for pair:{pair_contract_address}, avoiding trade volume calculations")
                return

            # calculate trade volume 24h
            [total_volume_24h, fees_24h, token0_volume_24h, token1_volume_24h] = calculate_pair_trade_volume(dag_chain_24h)

            # parse and store dag chain cids on IPFS
            volume_24h_cids = patch_cids_obj(dag_chain_24h, 'parse_cids', [])
            cids_volume_24h = await asyncio.gather(
                ipfs_client.add_json({ 
                    'resultant':{
                        "cids": volume_24h_cids,
                        'latestTimestamp_volume_24h': dag_chain_24h[0]['timestamp']
                    }
                })
            )
            # data = await ipfs_client.cat(cids_volume_24h[0])
            # print(f"cid get: {data}")
                

            # store last 24h recent logs, these will be used to show recent transaction for perticular contract
            recent_logs = await store_recent_transactions_logs(writer_redis_conn, dag_chain_24h, pair_contract_address)

            cids_volume_24h = cids_volume_24h[0]
            block_height_trade_volume = int(dag_chain_24h[0]['data']['payload']['chainHeightRange']['end'])
        else:
            
            sliding_window_front = []
            sliding_window_back = []

            # if head moved ahead
            if head_marker_24h > cached_trade_volume_data["processed_head_marker_24h"]:
                # front of the chain where head=current_head and tail=last_head
                sliding_window_front = await get_dag_blocks_in_range(
                    project_id_trade_volume, 
                    cached_trade_volume_data["processed_head_marker_24h"] + 1,
                    head_marker_24h, 
                    writer_redis_conn
                )
            
            # if tail moved ahead
            if tail_marker_24h > cached_trade_volume_data["processed_tail_marker_24h"]:
                # back of the chain where head=current_tail and tail=last_tail
                sliding_window_back = await get_dag_blocks_in_range(
                    project_id_trade_volume, 
                    cached_trade_volume_data["processed_tail_marker_24h"],
                    tail_marker_24h - 1,
                    writer_redis_conn
                )
            
            if not sliding_window_front and not sliding_window_back:
                logger.error(f"Avoiding calculation as new sliding window has not been created:{pair_contract_address} | sliding_window_front:{sliding_window_front} | sliding_window_back:{sliding_window_back}")
                return

            sliding_window_front_cids = []
            sliding_window_back_cids = []

            # parse and patch CID for 24h trade volume
            trade_volume_cids_24h = await ipfs_client.cat(cached_trade_volume_data["aggregated_volume_cid_24h"])
            if trade_volume_cids_24h:
                trade_volume_cids_24h = json.loads(trade_volume_cids_24h.decode('utf-8'))

            if sliding_window_back:
                [back_total_volume, back_fees, back_token0_volume, back_token1_volume] = calculate_pair_trade_volume(sliding_window_back)
                cached_trade_volume_data["aggregated_volume_24h"] -= back_total_volume
                cached_trade_volume_data["aggregated_fees_24h"] -= back_fees
                cached_trade_volume_data["aggregated_token0_volume_24h"] -= back_token0_volume
                cached_trade_volume_data["aggregated_token1_volume_24h"] -= back_token1_volume
                # cids for front of the chain
                sliding_window_back_cids = [{ 'dagCid': obj_24h['dagCid'], 'payloadCid': obj_24h['data']['cid']} for obj_24h in sliding_window_back]
                trade_volume_cids_24h["resultant"]["cids"] = patch_cids_obj(trade_volume_cids_24h["resultant"]["cids"], 'remove_back_patch', sliding_window_back_cids)

            if sliding_window_front:
                [front_total_volume, front_fees, front_token0_volume, front_token1_volume] = calculate_pair_trade_volume(sliding_window_front)
                cached_trade_volume_data["aggregated_volume_24h"] += front_total_volume
                cached_trade_volume_data["aggregated_fees_24h"] += front_fees
                cached_trade_volume_data["aggregated_token0_volume_24h"] += front_token0_volume
                cached_trade_volume_data["aggregated_token1_volume_24h"] += front_token1_volume
                # block height of trade volume
                block_height_trade_volume = int(sliding_window_front[0]['data']['payload']['chainHeightRange']['end'])
                # cids for front of the chain
                sliding_window_front_cids = [{ 'dagCid': obj_24h['dagCid'], 'payloadCid': obj_24h['data']['cid']} for obj_24h in sliding_window_front]
                trade_volume_cids_24h["resultant"]["cids"] = patch_cids_obj(trade_volume_cids_24h["resultant"]["cids"], 'add_front_patch', sliding_window_front_cids)
                # store recent logs
                recent_logs = await store_recent_transactions_logs(writer_redis_conn, sliding_window_front, pair_contract_address)
            else:
                block_height_trade_volume = cached_trade_volume_data["processed_block_height_trade_volume"]

            # store dag chain cids on IPFS
            latestTimestamp_volume_24h = sliding_window_front[0]['timestamp'] if sliding_window_front else trade_volume_cids_24h['resultant']['latestTimestamp_volume_24h']
            cids_volume_24h = await asyncio.gather(
            ipfs_client.add_json({ 
                    'resultant':{
                        "cids": trade_volume_cids_24h["resultant"]["cids"],
                        'latestTimestamp_volume_24h': latestTimestamp_volume_24h,
                    }
                })
            )

            # if liquidity reserve block height is not found
            if not block_height_total_reserve:
                block_height_total_reserve = cached_trade_volume_data["processed_block_height_total_reserve"]

            total_volume_24h = cached_trade_volume_data['aggregated_volume_24h']
            cids_volume_24h = cids_volume_24h[0]
            fees_24h = cached_trade_volume_data['aggregated_fees_24h']
            token0_volume_24h = cached_trade_volume_data['aggregated_token0_volume_24h']
            token1_volume_24h = cached_trade_volume_data['aggregated_token1_volume_24h']
        
        sliding_window_data = {
            "processed_tail_marker_24h": tail_marker_24h,
            "processed_head_marker_24h": head_marker_24h,
            "aggregated_volume_cid_24h": cids_volume_24h,
            "aggregated_volume_24h": total_volume_24h,
            "aggregated_fees_24h": fees_24h,
            "aggregated_token0_volume_24h": token0_volume_24h,
            "aggregated_token1_volume_24h": token1_volume_24h,
            "processed_block_height_total_reserve": block_height_total_reserve,
            "processed_block_height_trade_volume": block_height_trade_volume
        }
        prepared_snapshot = liquidityProcessedData(
            contractAddress=pair_contract_address,
            name=pair_token_metadata['pair']['symbol'],
            liquidity=f"US${round(abs(total_liquidity)):,}",
            volume_24h=f"US${round(abs(total_volume_24h)):,}",
            volume_7d=f"",
            fees_24h=f"US${round(abs(fees_24h)):,}",
            cid_volume_24h=cids_volume_24h, 
            cid_volume_7d="",
            block_height_total_reserve=block_height_total_reserve,
            block_height_trade_volume=block_height_trade_volume,
            token0Liquidity=token0_liquidity,
            token1Liquidity=token1_liquidity,
            token0TradeVolume_24h=token0_volume_24h,
            token1TradeVolume_24h=token1_volume_24h,
            token0TradeVolume_7d=0.0,
            token1TradeVolume_7d=0.0
        )
        
        await store_pair_daily_stats(writer_redis_conn, pair_contract_address, prepared_snapshot)

        logger.debug('Storing prepared trades, token reserves snapshot: %s', prepared_snapshot)
        logger.debug('Adding calculated snapshot to daily stats zset and pruning it for just 24h data')
        await writer_redis_conn.mset({
            redis_keys.get_uniswap_pair_contract_V2_pair_data(f"{Web3.toChecksumAddress(pair_contract_address)}"): prepared_snapshot.json(),
            redis_keys.get_uniswap_pair_cache_sliding_window_data(f"{Web3.toChecksumAddress(pair_contract_address)}"): json.dumps(sliding_window_data)
        })

    except Exception as e:
        logger.error('Error in process_pairs_trade_volume_and_reserves: %s', e, exc_info=True)
        raise(e)

    logger.debug(f"Calculated v2 pair data for contract: {pair_contract_address} | symbol:{pair_token_metadata['token0']['symbol']}-{pair_token_metadata['token1']['symbol']}")
    return json.loads(prepared_snapshot.json())



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

        logger.debug("Create threads to process trade volume data")
        process_data_list = []
        for pair_contract_address in pairs:
            t = process_pairs_trade_volume_and_reserves(aioredis_pool.writer_redis_pool, pair_contract_address)
            process_data_list.append(t)

        final_results = await asyncio.gather(*process_data_list, return_exceptions=True)
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
    # print("", "")
    # print(data)