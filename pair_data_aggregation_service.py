from typing import List
from web3 import Web3
from utils import redis_keys
import asyncio
import aiohttp
import math
from functools import wraps, partial
import json
import os
import sys
import logging.config
from config import settings
from utils.ipfs_async import client as ipfs_client
from utils.retrieval_utils import retrieve_block_data
from utils.redis_conn import provide_async_reader_conn_inst, provide_async_writer_conn_inst, RedisPool
import aioredis
from utils import helper_functions, dag_utils
from data_models import liquidityProcessedData

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

async def get_dag_block_by_height(project_id, block_height):
    dag_block = {}
    try:
        dag_cid = await helper_functions.get_dag_cid(project_id=project_id, block_height=block_height)
        dag_block = await retrieve_block_data(block_dag_cid=dag_cid, data_flag=1)
        dag_block["dagCid"] = dag_cid
    except Exception as e:
        logger.error(f"Error: can't get dag block with msg: {str(e)} | projectId:{project_id}, block_height:{block_height}", exc_info=True)
        dag_block = {}

    return dag_block


async def get_dag_blocks_in_range(project_id, from_block, to_block):
    tasks = []
    for i in range(from_block, to_block + 1):
        t = get_dag_block_by_height(project_id, i)
        tasks.append(t)
    
    dag_chain = await asyncio.gather(*tasks)

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
        logger.debug(f"Reading {file_path} file")
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
    logger.debug(f"Fetch pair data, pair address:{pair_address}")

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
    
    logger.debug(f"Fetch token data, token0:{token0Addr} token1:{token1Addr}")
    
    if pair_tokens_data:
        logger.debug(f"got pair token data from cache | pair:{pair_address}")

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

        logger.debug(f"Storing pair tokens data in redis | pair:{pair_address}")

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
        

async def process_pairs_trade_volume_and_reserves(writer_redis_conn, pair_contract_address):
    try:
        project_id_trade_volume = f'uniswap_pairContract_trade_volume_{pair_contract_address}_UNISWAPV2'
        project_id_token_reserve = f'uniswap_pairContract_pair_total_reserves_{pair_contract_address}_UNISWAPV2'

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

        # trade volume dag chain
        tail_marker_24h = await writer_redis_conn.get(redis_keys.get_sliding_window_cache_tail_marker(project_id_trade_volume, '24h'))
        head_marker_24h = await writer_redis_conn.get(redis_keys.get_sliding_window_cache_head_marker(project_id_trade_volume, '24h'))
        tail_marker_24h = int(tail_marker_24h.decode('utf-8')) if tail_marker_24h else 0
        head_marker_24h = int(head_marker_24h.decode('utf-8')) if head_marker_24h else 0
        # tail_marker_7d = await writer_redis_conn.get(redis_keys.get_sliding_window_cache_tail_marker(project_id_trade_volume, '7d'))
        # head_marker_7d = await writer_redis_conn.set(redis_keys.get_sliding_window_cache_head_marker(project_id_trade_volume, '7d'))
        
        dag_chain_24h = await get_dag_blocks_in_range(project_id_trade_volume, tail_marker_24h, head_marker_24h)

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
        
        # liquidty data
        liquidity_head_marker = await writer_redis_conn.get(redis_keys.get_sliding_window_cache_head_marker(project_id_token_reserve, '24h'))
        liquidity_head_marker = int(liquidity_head_marker.decode('utf-8')) if liquidity_head_marker else 0
        liquidity_data = await get_dag_block_by_height(project_id_token_reserve, liquidity_head_marker)
        
        if not liquidity_data:
            logger.error(f"No liquidity data avaiable for - projectId:{project_id_token_reserve} and liquidity head marker:{liquidity_head_marker}")
            return
        
        token0_liquidity = float(list(liquidity_data['data']['payload']['token0Reserves'].values())[-1])
        token1_liquidity = float(list(liquidity_data['data']['payload']['token1Reserves'].values())[-1])
        total_liquidity = token0_liquidity * token0Price + token1_liquidity * token1Price

        # get data for last 24hour
        total_volume_24h = sum(map(lambda x: x['data']['payload']['totalTrade'], dag_chain_24h))

        # calculate / sum last 24h fee
        fees_24h = sum(map(lambda x: x['data']['payload'].get('totalFee', 0), dag_chain_24h))

        # get volume for individual tokens (in native token decimals):
        token0_volume_24h = sum(map(lambda x: x['data']['payload']['token0TradeVolume'], dag_chain_24h))
        token1_volume_24h = sum(map(lambda x: x['data']['payload']['token1TradeVolume'], dag_chain_24h))


        # get data for last 7d
        # volume_7d = sum(map(lambda x: x['data']['payload']['totalTrade'], volume_7d_data))

        
        
        volume_24h_cids = [{ 'dagCid': obj_24h['dagCid'], 'payloadCid': obj_24h['data']['cid']} for obj_24h in dag_chain_24h]
        # volume_7d_cids = [{ 'dagCid': obj_7d['dagCid'], 'payloadCid': obj_7d['data']['cid']} for obj_7d in volume_7d_data]

        cids_volume_24h = await asyncio.gather(
            ipfs_client.add_json({ 'resultant':{
                "cids": volume_24h_cids,
                'latestTimestamp_volume_24h': dag_chain_24h[0]['timestamp'],
                'earliestTimestamp_volume_24h': dag_chain_24h[-1]['timestamp']
            }})
        )

        # store last 24h recent logs, these will be used to show recent transaction for perticular contract
        recent_logs = list()
        for log in dag_chain_24h:
            recent_logs.extend(log['data']['payload']['recent_logs'])

        recent_logs = sorted(recent_logs, key=lambda log: log['timestamp'], reverse=True)

        if len(recent_logs) > 70:
            recent_logs = recent_logs[:70]

        await writer_redis_conn.set(
            redis_keys.get_uniswap_pair_cached_recent_logs(f"{Web3.toChecksumAddress(pair_contract_address)}"), 
            json.dumps(recent_logs)
        )
        logger.debug(f"Stored recent logs for pair: {redis_keys.get_uniswap_pair_cached_recent_logs(f'{Web3.toChecksumAddress(pair_contract_address)}')}, of len:{len(recent_logs)}")

        logger.debug('Calculated 24h, 7d and fees_24h vol: %s, %s | contract: %s', total_volume_24h, fees_24h, pair_contract_address)
        prepared_snapshot = liquidityProcessedData(
            contractAddress=pair_contract_address,
            name=pair_token_metadata['pair']['symbol'],
            liquidity=f"US${round(abs(total_liquidity)):,}",
            volume_24h=f"US${round(abs(total_volume_24h)):,}",
            volume_7d=f"",
            fees_24h=f"US${round(abs(fees_24h)):,}",
            cid_volume_24h=cids_volume_24h[0], 
            cid_volume_7d="",
            block_height_total_reserve=int(liquidity_data['data']['payload']['chainHeightRange']['end']),
            block_heigh_trade_volume=int(dag_chain_24h[0]['data']['payload']['chainHeightRange']['end']),
            token0Liquidity=token0_liquidity,
            token1Liquidity=token1_liquidity,
            token0TradeVolume_24h=token0_volume_24h,
            token1TradeVolume_24h=token1_volume_24h,
            token0TradeVolume_7d=0.0,
            token1TradeVolume_7d=0.0
        )
    except Exception as e:
        logger.error('Error in process_pairs_trade_volume_and_reserves: %s', e, exc_info=True)
        raise(e)
    
    
    logger.debug('Prepared trades, token reserves snapshot and storing it in redis: %s', prepared_snapshot)
    r = await writer_redis_conn.set(
        redis_keys.get_uniswap_pair_contract_V2_pair_data(f"{Web3.toChecksumAddress(pair_contract_address)}"),
        prepared_snapshot.json())
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
    # data = loop.run_until_complete(pair_recent_metadata("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc"))
    # print("", "")
    # print(data)