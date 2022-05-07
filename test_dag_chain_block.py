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


async def verify_dag_block_calculations(dag_cid):
    aioredis_pool = RedisPool()
    await aioredis_pool.populate()
    redis_read_conn = aioredis_pool.reader_redis_pool

    dagBlockData = await get_dag_cid_output(dag_cid)

    payload_cid = dagBlockData['data']['cid']
    payloadData = await get_payload_cid_output(payload_cid)

    pair_tokens_data = await redis_read_conn.hgetall(redis_keys.get_uniswap_pair_contract_tokens_data(Web3.toChecksumAddress(payloadData['contract'])))
    token0_symbol = pair_tokens_data[b"token0_symbol"].decode('utf-8') if pair_tokens_data else "Token0"
    token1_symbol = pair_tokens_data[b"token1_symbol"].decode('utf-8') if pair_tokens_data else "Token1"

    if len(payloadData['recent_logs']) <= 0:
        print("No transactions in dag block")
        return

    print("\n")
    table = Table(show_header=True, header_style="bold magenta")
    table.add_column("Event", style="dim", width=12)
    table.add_column(token0_symbol + " amount", justify="center")
    table.add_column(token1_symbol + " amount", justify="center")
    table.add_column("Amount in usd", justify="center")
    table.add_column("Timestamp / Calculation Age", justify="center")
    for log in payloadData['recent_logs']:
        ts = int(round(datetime.now().timestamp())) - log["timestamp"]
        table.add_row(
            log["event"], 
            str(round(log["token0_amount"], 2)), 
            str(round(log["token1_amount"], 2)), 
            str(round(log["trade_amount_usd"], 2)),
            pretty_relative_time(ts)
        )

    console.print(table)
    console.print("\n")
    console.print("[bold magenta]DAG CID:[/bold magenta]", f"[bold bright_cyan]{dag_cid}[/bold bright_cyan]")
    console.print("[bold magenta]Contract:[/bold magenta]", payloadData['contract'])
    console.print("[bold magenta]Pair:[/bold magenta]", f"{token0_symbol} - {token1_symbol}")
    console.print("[bold magenta]Total trade Volume (USD):[/bold magenta]", str(round(payloadData['totalTrade'], 2)))
    console.print(f"[bold magenta]Total trade Volume ({token0_symbol}):[/bold magenta]", str(round(payloadData['token0TradeVolume'], 2)))
    console.print(f"[bold magenta]Total trade Volume ({token1_symbol}):[/bold magenta]", str(round(payloadData['token1TradeVolume'], 2)))
    console.print("[bold magenta]Timestamp:[/bold magenta]", str(payloadData['timestamp']))
    console.print("\n")

        

async def verify_trade_volume_cids(data_cid):

    data = await get_payload_cid_output(data_cid)
    trade_volume_24h_cids = data['resultant']['trade_volume_24h_cids']

    print(f"\nTotal trade volume 24h cids objects: {len(trade_volume_24h_cids)}\n")
    print(f"dag_cid: {trade_volume_24h_cids[25]['dagCid']}")

    for i in range(1):
        dag_cid = trade_volume_24h_cids[25]['dagCid']
        payload_cid = trade_volume_24h_cids[25]['payloadCid']

        dagBlockData = await get_dag_cid_output(dag_cid)

        payload_cid = dagBlockData['data']['cid']
        payloadData = await get_payload_cid_output(payload_cid)

        print(f"dagBlockData: {dagBlockData}")
        print(f"payloadData: {payloadData}")



if __name__ == "__main__":
    # trade_volume_24h_cid = 'QmcY1LbyTTN8PSVoij2sCRnbUuxcCHFHMWAoebgQueGX6j'
    dag_block_cid = str(sys.argv[1]) if len(sys.argv) > 1 else 'bafyreidwzdemrzfr5tgzwothh6dc2qidht4kacazd23rghbjjlynnwrf5a'
    tasks = asyncio.gather(
        verify_dag_block_calculations(dag_block_cid)
    )
    asyncio.get_event_loop().run_until_complete(tasks)