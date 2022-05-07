from utils.ipfs_async import client as ipfs_client
from utils.redis_conn import RedisPool
from utils import redis_keys
from web3 import Web3
import asyncio
import json
from datetime import datetime
import sys
from rich.console import Console
from rich.table import Table


console = Console()

def pretty_relative_time(time_diff_secs):
    weeks_per_month = 365.242 / 12 / 7
    intervals = [('minute', 60), ('hour', 60), ('day', 24), ('week', 7),
                 ('month', weeks_per_month), ('year', 12)]

    unit, number = 'second', abs(time_diff_secs)
    for new_unit, ratio in intervals:
        new_number = float(number) / ratio
        # If the new number is too small, don't go to the next unit.
        if new_number < 2:
            break
        unit, number = new_unit, new_number
    shown_num = int(number)
    return '{} {}'.format(shown_num, unit + (' ago' if shown_num == 1 else 's ago'))

async def get_dag_cid_output(dag_cid):
    out = await ipfs_client.dag.get(dag_cid)
    if not out:
        return {}
    return out.as_json()

async def get_payload_cid_output(cid):
    out = await ipfs_client.cat(cid)
    if not out:
        return {}
    return json.loads(out.decode('utf-8'))

        

async def verify_trade_volume_cids(data_cid):

    data = await get_payload_cid_output(data_cid)
    trade_volume_24h_cids = data['resultant']['trade_volume_24h_cids']

    aioredis_pool = RedisPool()
    await aioredis_pool.populate()
    redis_read_conn = aioredis_pool.reader_redis_pool

    #TODO: proceed only CID has dagCids and payload cids
    payloadCid_0 = trade_volume_24h_cids[0]['payloadCid']
    payloadData_0 = await get_payload_cid_output(payloadCid_0)
    pair_tokens_data = await redis_read_conn.hgetall(redis_keys.get_uniswap_pair_contract_tokens_data(Web3.toChecksumAddress(payloadData_0['contract'])))
    token0_symbol = pair_tokens_data[b"token0_symbol"].decode('utf-8') if pair_tokens_data else "Token0"
    token1_symbol = pair_tokens_data[b"token1_symbol"].decode('utf-8') if pair_tokens_data else "Token1"

    table = Table(show_header=True, header_style="bold magenta")
    table.add_column("Chain Height Range", justify="center")
    table.add_column("No. of Tx", justify="center")
    table.add_column(token0_symbol + " amount", justify="center")
    table.add_column(token1_symbol + " amount", justify="center")
    table.add_column("Total Trade Volume (USD)", justify="center")
    table.add_column("Timestamp / Calculation Age", justify="center")
    total_trade_volume_usd = 0

    for i in range(len(trade_volume_24h_cids)):
        dag_cid = trade_volume_24h_cids[i]['dagCid']
        dagBlockData = await get_dag_cid_output(dag_cid)

        payload_cid = dagBlockData['data']['cid']
        payloadData = await get_payload_cid_output(payload_cid)
        
        ts = int(round(datetime.now().timestamp())) - payloadData["timestamp"]
        table.add_row(
            f"{payloadData['chainHeightRange']['begin']} - {payloadData['chainHeightRange']['end']}", 
            str(len(payloadData["recent_logs"])), 
            str(round(payloadData["token0TradeVolume"], 2)), 
            str(round(payloadData["token1TradeVolume"], 2)), 
            str(round(payloadData["totalTrade"], 2)),
            pretty_relative_time(ts)
        )
            

        total_trade_volume_usd += float(payloadData["totalTrade"])
    
    console.print("\n")
    console.print(table)
    console.print("\n")
    console.print("[bold magenta]MASTER CID:[/bold magenta]", f"[bold bright_cyan]{data_cid}[/bold bright_cyan]")
    console.print("[bold magenta]Contract:[/bold magenta]", payloadData_0['contract'])
    console.print("[bold magenta]Pair:[/bold magenta]", f"{token0_symbol} - {token1_symbol}")
    console.print("[bold magenta]Total Trade Volume Combined:[/bold magenta]", f"{total_trade_volume_usd}")
    console.print("\n")
    


if __name__ == "__main__":
    trade_volume_24h_cid = str(sys.argv[1]) if len(sys.argv) > 1 else 'QmTDH4txPmx49UQtDuNb16pGXXBCr2n4atnBDSzbapHYtx'
    tasks = asyncio.gather(
        verify_trade_volume_cids(trade_volume_24h_cid)
    )
    asyncio.get_event_loop().run_until_complete(tasks)