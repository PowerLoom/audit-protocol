from typing import Optional
from utils import redis_conn
from utils import redis_keys
import typer
import redis
import json

REDIS_CONN_CONF = redis_conn.REDIS_CONN_CONF
app = typer.Typer()

def read_json_file(file_path: str):
    """Read given json file and return its content as a dictionary."""
    try:
        f_ = open(file_path, 'r')
    except Exception as e:
        print(f"Unable to open the {file_path} file, error msg:{str(e)}")
    else:
        print(f"Reading {file_path} file")
        json_data = json.loads(f_.read())
    return json_data

@app.command()
def hello():
    print("cli cmd script for audit protocol")

@app.command()
def dagChainStatus(namespace: str = typer.Argument('UNISWAPV2'), dag_chain_height: int = typer.Argument(-1)):
    
    dag_chain_height = dag_chain_height if dag_chain_height > -1 else '-inf'

    r = redis.Redis(**REDIS_CONN_CONF, single_connection_client=True)
    pair_contracts = read_json_file('static/cached_pair_addresses.json')
    pair_projects = [
        'projectID:uniswap_pairContract_trade_volume_{}_'+namespace+':{}',
        'projectID:uniswap_pairContract_pair_total_reserves_{}_'+namespace+':{}'
    ]
    total_zsets = {}
    total_issue_count = {
        "CURRENT_DAG_CHAIN_HEIGHT": 0,
        "LAG_EXIST_IN_DAG_CHAIN": 0
    }

    # get highest dag chain height
    project_heights = r.hgetall(redis_keys.get_uniswap_projects_dag_verifier_status(namespace))
    if project_heights:
        for key, value in project_heights.items():
            if int(value.decode('utf-8')) > total_issue_count["CURRENT_DAG_CHAIN_HEIGHT"]:
                total_issue_count["CURRENT_DAG_CHAIN_HEIGHT"] = int(value.decode('utf-8'))
    
    def get_zset_data(key, min, max, pair_address):
        res = r.zrangebyscore(
            name=key.format(pair_address, "dagChainGaps"),
            min=min,
            max=max
        )
        tentative_block_height, block_height = r.mget(
            [
                key.format(f"{pair_address}", "tentativeBlockHeight"),
                key.format(f"{pair_address}", "blockHeight")
            ]
        )
        key_based_issue_stats = {
            "CURRENT_DAG_CHAIN_HEIGHT": 0
        }

        # add tentative and current block height
        if tentative_block_height:
            tentative_block_height = int(tentative_block_height.decode("utf-8")) if type(tentative_block_height) is bytes else int(tentative_block_height)
        else:
            tentative_block_height = None
        if block_height:
            block_height = int(block_height.decode("utf-8")) if type(block_height) is bytes else int(block_height)
        else:
            block_height = None
        key_based_issue_stats["CURRENT_LAG_IN_DAG_CHAIN_HEIGHT"] = tentative_block_height - block_height if tentative_block_height and block_height else None
        if key_based_issue_stats["CURRENT_LAG_IN_DAG_CHAIN_HEIGHT"] != None:
            total_issue_count["LAG_EXIST_IN_DAG_CHAIN"] = key_based_issue_stats["CURRENT_LAG_IN_DAG_CHAIN_HEIGHT"] if key_based_issue_stats["CURRENT_LAG_IN_DAG_CHAIN_HEIGHT"] > total_issue_count["LAG_EXIST_IN_DAG_CHAIN"] else total_issue_count["LAG_EXIST_IN_DAG_CHAIN"]
        
        if res:
            # parse zset entry
            parsed_res = []
            for entry in res:
                entry = json.loads(entry)
                
                # create/add issue entry in overall issue structure
                if not entry["issueType"] + "_ISSUE_COUNT" in total_issue_count:
                    total_issue_count[entry["issueType"] + "_ISSUE_COUNT" ] = 0
                if not entry["issueType"] + "_BLOCKS" in total_issue_count:
                    total_issue_count[entry["issueType"] + "_BLOCKS" ] = 0

                # create/add issue entry in "KEY BASED" issue structure
                if not entry["issueType"] + "_ISSUE_COUNT" in key_based_issue_stats:
                    key_based_issue_stats[entry["issueType"] + "_ISSUE_COUNT" ] = 0
                if not entry["issueType"] + "_BLOCKS" in key_based_issue_stats:
                    key_based_issue_stats[entry["issueType"] + "_BLOCKS" ] = 0
                
                    
                # gather overall issue stats
                total_issue_count[entry["issueType"] + "_ISSUE_COUNT" ] += 1
                key_based_issue_stats[entry["issueType"] + "_ISSUE_COUNT" ] +=  1

                if entry["missingBlockHeightEnd"] == entry["missingBlockHeightStart"]:
                    total_issue_count[entry["issueType"] + "_BLOCKS" ] +=  1
                    key_based_issue_stats[entry["issueType"] + "_BLOCKS" ] +=  1
                else:
                    total_issue_count[entry["issueType"] + "_BLOCKS" ] +=  entry["missingBlockHeightEnd"] - entry["missingBlockHeightStart"] + 1
                    key_based_issue_stats[entry["issueType"] + "_BLOCKS" ] +=  entry["missingBlockHeightEnd"] - entry["missingBlockHeightStart"] + 1
                    
                # store latest dag block height for projectId
                if entry["dagBlockHeight"] > key_based_issue_stats["CURRENT_DAG_CHAIN_HEIGHT"]:
                    key_based_issue_stats["CURRENT_DAG_CHAIN_HEIGHT"] = entry["dagBlockHeight"]

                parsed_res.append(entry)
            res = parsed_res

            print(f"{key.format(pair_address, '')} - ")
            for k, v in key_based_issue_stats.items():
                print(f"\t {k} : {v}")
        else:
            del key_based_issue_stats["CURRENT_DAG_CHAIN_HEIGHT"]
            key_based_issue_stats["tentative_block_height"] = tentative_block_height
            key_based_issue_stats["block_height"] = block_height
            print(f"{key.format(pair_address, '')} - ")
            for k, v in key_based_issue_stats.items():
                print(f"\t {k} : {v}")
            res = []
        return res
    
    def gather_all_zset(contracts, projects):
        for project in projects:
            for addr in contracts:
                zset_key = project.format(addr, "dagChainGaps")
                total_zsets[zset_key] = get_zset_data(project, dag_chain_height, '+inf', addr)
    
    gather_all_zset(pair_contracts, pair_projects)

    if total_issue_count["LAG_EXIST_IN_DAG_CHAIN"] > 3:
        total_issue_count["LAG_EXIST_IN_DAG_CHAIN"] = f"THERE IS A LAG WHILE PROCESSING CHAIN, BIGGEST LAG: {total_issue_count['LAG_EXIST_IN_DAG_CHAIN']}"
    else:
        total_issue_count["LAG_EXIST_IN_DAG_CHAIN"] = "NO LAG"
        

    print(f"\n======================================> OVERALL ISSUE STATS: \n")
    for k, v in total_issue_count.items():
        print(f"\t {k} : {v}")

    if len(total_issue_count) < 2:
        print(f"\n##################### NO GAPS FOUND IN CHAIN #####################\n")

    print("\n")

if __name__ == '__main__':
    app()
