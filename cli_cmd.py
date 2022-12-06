from itertools import cycle
from textwrap import wrap
from typing import Optional
from settings_model import Settings
from utils import redis_conn
from utils import redis_keys
import typer
import redis
import json
from rich.console import Console
from rich.table import Table
from rich.text import Text
from datetime import datetime
from rich.pretty import pretty_repr

import re

console = Console()

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

@app.command()
def updateStoredProjectIds(namespace: str = typer.Argument('UNISWAPV2')):
    r = redis.Redis(**REDIS_CONN_CONF, single_connection_client=True)

    all_contracts = read_json_file('static/cached_pair_addresses.json')
    all_projectIds = []
    for contract in all_contracts:
        pair_reserve_template = f"uniswap_pairContract_pair_total_reserves_{contract}_{namespace}"
        pair_trade_volume_template = f"uniswap_pairContract_trade_volume_{contract}_{namespace}"
        all_projectIds.append(pair_reserve_template)
        all_projectIds.append(pair_trade_volume_template)

    all_projectIds.append(f"uniswap_V2PairsSummarySnapshot_{namespace}")
    all_projectIds.append(f"uniswap_V2TokensSummarySnapshot_{namespace}")
    all_projectIds.append(f"uniswap_V2DailyStatsSnapshot_{namespace}")

    r.sadd('storedProjectIds', *all_projectIds)

@app.command()
def projectIndexStatus(namespace: str = typer.Option("UNISWAPV2", "--namespace"), projectId: str = typer.Option(None, "--projectId")):
    r = redis.Redis(**REDIS_CONN_CONF, single_connection_client=True)

    indexeStatus = None
    if projectId:
        indexeStatus = r.hget(f'projects:{namespace}:IndexStatus', projectId)
        if not indexeStatus:
            console.log(f"\n[bold red]Project is not indexed[bold red]: \n{projectId}\n")
            return
        indexeStatus = [{projectId: indexeStatus.decode('utf-8')}]
    else:
        indexeStatus = r.hgetall(f'projects:{namespace}:IndexStatus')
        if not indexeStatus:
            console.log(f"\n[bold red]Indexes map doesn't exist [bold red]: 'projects:{namespace}:IndexStatus'\n")
            return
        indexeStatus = [{k.decode('utf-8'): v.decode('utf-8')} for k, v in indexeStatus.items()]

    table = Table(show_header=True, header_style="bold magenta")
    table.add_column("ProjecId", justify="center")
    table.add_column("Start source chain height", justify="center")
    table.add_column("Current source chain height", justify="center")

    for project_indexes in indexeStatus:
        k, v = project_indexes.popitem()
        v = json.loads(v)
        table.add_row(
            Text(k, justify="left", overflow="ellipsis"),
            str(v["startSourceChainHeight"]),
            str(v["currentSourceChainHeight"]),
        )

    console.print(table)


@app.command()
def pruning_cycles_status(cycles: int = typer.Option(3, "--cycles")):

    r = redis.Redis(**REDIS_CONN_CONF, single_connection_client=True)

    cycles = 20 if cycles > 20 else cycles
    cycles = 3 if cycles < 1 else cycles

    pruningStatusZset = r.zrangebyscore(
        name='pruningRunStatus',
        min='-inf',
        max='+inf',
        withscores=True
    )

    table = Table(show_header=True, header_style="bold magenta", show_lines=True )
    table.add_column("Timestamp", justify="center", vertical="middle")
    table.add_column("Cycle status", justify="center")

    for entry in reversed(pruningStatusZset):
        payload, timestamp = entry
        payload = json.loads(payload.decode('utf-8')) if payload else {}
        timestamp = int(timestamp/1000) if timestamp and len(str(int(timestamp))) > 10  else timestamp

        payload_text = Text()
        payload_text.append(f"pruningCycleID: {payload.get('pruningCycleID')}\n")
        payload_text.append(f"cycleStartTime: {payload.get('cycleStartTime')}\n")
        payload_text.append(f"cycleEndTime: {payload.get('cycleEndTime')}\n")
        payload_text.append(f"projectsCount: {payload.get('projectsCount')}\n")
        payload_text.append(f"projectsProcessSuccessCount: {payload.get('projectsProcessSuccessCount')}\n", style="bold green")
        payload_text.append(f"projectsProcessFailedCount: {payload.get('projectsProcessFailedCount')}\n", style="bold red")
        payload_text.append(f"projectsNotProcessedCount: {payload.get('projectsNotProcessedCount')}", style="bold yellow")

        table.add_row(
            Text(f"{datetime.fromtimestamp(timestamp)} ( {timestamp} )"),
            payload_text
        )

        cycles -= 1
        if cycles <= 0:
            break

    console.print(table)


@app.command()
def pruning_cycle_project_report(cycleId: str = typer.Option(None, "--cycleId")):

    r = redis.Redis(**REDIS_CONN_CONF, single_connection_client=True)
    cycleDetails = {}

    if not cycleId:
        cycleDetails = r.zrevrange(
            name='pruningRunStatus',
            start=0,
            end=0
        )
        cycleDetails = json.loads(cycleDetails[0].decode('utf-8')) if len(cycleDetails)>0 else {}
    else:
        allCycles = r.zrangebyscore(
            name='pruningRunStatus',
            min='-inf',
            max='+inf'
        )
        for cycle in allCycles:
            cycle = json.loads(cycle.decode('utf-8')) if cycle else {}
            if cycle.get('pruningCycleID', None) == cycleId:
                cycleDetails = cycle
                break

    cycleStartTime = cycleDetails.get('cycleStartTime', None)
    cycleStartTime = int(cycleStartTime/1000) if cycleStartTime and len(str(int(cycleStartTime))) > 10  else cycleStartTime
    cycleEndTime = cycleDetails.get('cycleEndTime', None)
    cycleEndTime = int(cycleEndTime/1000) if cycleEndTime and len(str(int(cycleEndTime))) > 10  else cycleEndTime

    allProjectsDetails = r.hgetall(f"pruningProjectDetails:{cycleDetails.get('pruningCycleID', None)}")
    if not allProjectsDetails:
        console.print(f"[bold red]Can't find project details- pruningProjectDetails:{cycleDetails.get('pruningCycleID', None)}[/bold red]\n")
    else:
        table = Table(show_header=True, header_style="bold magenta", show_lines=True )
        table.add_column("ProjectId", overflow="fold", justify="center", vertical="middle")
        table.add_column("Details", justify="center")

        for projectId, projectDetails in allProjectsDetails.items():
            projectId = projectId.decode('utf-8') if projectId else {}
            projectDetails = json.loads(projectDetails.decode('utf-8')) if projectDetails else {}

            payload_text = Text()
            payload_text.append(f"DAGSegmentsProcessed: {projectDetails.get('DAGSegmentsProcessed')}\n", style="bold green")
            payload_text.append(f"DAGSegmentsArchived: {projectDetails.get('DAGSegmentsArchived')}\n", style="bold green")
            payload_text.append(f"CIDsUnPinned: {projectDetails.get('CIDsUnPinned')}\n")
            if projectDetails.get('DAGSegmentsArchivalFailed', False):
                payload_text.append(f"DAGSegmentsArchivalFailed: {projectDetails.get('DAGSegmentsArchivalFailed')}\n", style="bold red")
            if projectDetails.get('failureCause', False):
                payload_text.append(f"failureCause: {projectDetails.get('failureCause')}\n", style="bold red")
            if projectDetails.get('unPinFailed', False):
                payload_text.append(f"unPinFailed: {projectDetails.get('unPinFailed')}\n", style="bold red")

            table.add_row(
                Text(projectId, style='bright_cyan'),
                payload_text
            )

        console.print(table)


    console.print("\n\n[bold magenta]Pruning cycleId:[/bold magenta]", f"[bold bright_cyan]{cycleDetails.get('pruningCycleID', None)}[/bold bright_cyan]")
    console.print("[bold magenta]Start timestamp:[/bold magenta]", f"[white] {datetime.fromtimestamp(cycleStartTime)} ( {cycleStartTime} )[/white]")
    console.print("[bold magenta]End timestamp:[/bold magenta]", f"[white]{datetime.fromtimestamp(cycleEndTime)} ( {cycleEndTime} )[/white]")
    console.print("[bold blue]Projects count:[/bold blue]", f"[bold blue]{cycleDetails.get('projectsCount', None)}[/bold blue]")
    console.print("[bold green]Success count:[/bold green]", f"[bold green]{cycleDetails.get('projectsProcessSuccessCount', None)}[/bold green]")
    console.print("[bold red]Failure counts:[/bold red]", f"[bold red]{cycleDetails.get('projectsProcessFailedCount', None)}[/bold red]")
    console.print("[bold yellow]Unprocessed Project count:[/bold yellow]", f"[bold yellow]{cycleDetails.get('projectsNotProcessedCount', None)}[/bold yellow]\n\n")

@app.command()
def identify_projects_that_require_force_resubmission(namespace: str = typer.Option(None, "--namespace")):
    r = redis.Redis(**REDIS_CONN_CONF, single_connection_client=True)

    if not namespace:
        console.log("No namespace provided, hence checking for all namespaces")
        namespace = "*"

    projectId = "*"
    count = 0

    blockHeightProjects_pattern = f"projectID:*{projectId}*{namespace}*:blockHeight"
    for project in r.scan_iter(match=blockHeightProjects_pattern):
        try:
            project = project.decode('utf-8')

            projectHeight = r.get(project)
            projectHeight = projectHeight.decode('utf-8')

            blockHeightProject = int(projectHeight)+1
            pendingTxProject = re.sub(r'blockHeight', 'pendingTransactions', project)

            # check if pending transaction exist at project's block-height
            pendingTransactions = r.zrangebyscore(
                name=pendingTxProject,
                min=blockHeightProject,
                max=blockHeightProject,
                withscores=True,
            )
            if len(pendingTransactions) > 0:
                projectId = re.sub(r':pendingTransactions', '', pendingTxProject)
                pendingTxScore = int(pendingTransactions[0][1])
                pendingTxValue = json.loads((pendingTransactions[0][0]).decode('utf-8'))
                if pendingTxValue["lastTouchedBlock"] != 0:
                    count = count+1
                else:
                    continue
                console.log(f"\n[bold magenta]{project}[bold magenta]:")
                console.log(f"[bold blue]Forced project resubmission at height:[bold blue] [white]{pendingTxScore}[white]\n")

        except Exception as e:
            print(f"Error: {str(e)} | project: {project}")
    if count == 0:
       console.log("No projects require force resubmission")

@app.command()
def force_project_height_into_resubmission(namespace: str = typer.Option(None, "--namespace"), projectId: str = typer.Option(None, "--project")):
    r = redis.Redis(**REDIS_CONN_CONF, single_connection_client=True)

    print("\nThis command will overwrite project's redis state and force resubmission,.\n ")
    count = 0

    if not namespace:
        console.log("No namespace provided, hence forcing resubmission for all namespaces")
        namespace = "*"

    if not projectId:
        projectId = "*"

    blockHeightProjects_pattern = f"projectID:*{projectId}*{namespace}*:blockHeight"
    for project in r.scan_iter(match=blockHeightProjects_pattern):
        try:
            project = project.decode('utf-8')

            projectHeight = r.get(project)
            projectHeight = projectHeight.decode('utf-8')

            blockHeightProject = int(projectHeight)+1
            pendingTxProject = re.sub(r'blockHeight', 'pendingTransactions', project)

            # check if pending transaction exist at project's block-height
            pendingTransactions = r.zrangebyscore(
                name=pendingTxProject,
                min=blockHeightProject,
                max=blockHeightProject
            )
            if len(pendingTransactions) > 0:

                # get lastest pending transaction after current project height
                pendingLastestTransactions = r.zrangebyscore(
                    name=pendingTxProject,
                    min=blockHeightProject,
                    max=blockHeightProject,
                    withscores=True,
                )

                if len(pendingLastestTransactions) > 0:
                    projectId = re.sub(r':pendingTransactions', '', pendingTxProject)
                    pendingTxScore = int(pendingLastestTransactions[0][1])
                    pendingTxValue = json.loads((pendingLastestTransactions[0][0]).decode('utf-8'))
                    if pendingTxValue["lastTouchedBlock"] != 0:
                        pendingTxValue["lastTouchedBlock"] = 0
                        count = count+1
                    else:
                        continue
                    # remove old entry from zset
                    r.zremrangebyscore(
                        name=pendingTxProject,
                        min=pendingTxScore,
                        max=pendingTxScore
                    )

                    # add new entry to zset (lastTouchedBlockHeight = 0)
                    r.zadd(
                        name=pendingTxProject,
                        mapping={json.dumps(pendingTxValue): int(pendingTxScore)}
                    )

                    console.log(f"\n[bold magenta]{projectId}[bold magenta]:")
                    console.log(f"[bold blue]Forced project resubmission at height:[bold blue] [white]{pendingTxScore}[white]\n")

        except Exception as e:
            print(f"Error: {str(e)} | project: {project}")
    if count == 0:
       console.log("No projects required force resubmission")

@app.command()
def force_summary_project_height_ahead(namespace: str = typer.Option(None, "--namespace")):
    r = redis.Redis(**REDIS_CONN_CONF, single_connection_client=True)

    print("\nThis command will force-push Summary project's redis state ahead.\n ")
    count = 0

    if not namespace:
        console.log("No namespace provided, please provide a namespace")
        return
    projects = [
        f"projectID:uniswap_V2TokensSummarySnapshot_{namespace}",
        f"projectID:uniswap_V2PairsSummarySnapshot_{namespace}",
        f"projectID:uniswap_V2DailyStatsSnapshot_{namespace}"
        ]
    for project in projects:
        try:
            block_height_key = f"{project}:blockHeight"
            tentative_bh_key = f"{project}:tentativeBlockHeight"
            console.log(f"\n[bold magenta]{project}[bold magenta]:")

            project_height = r.get(block_height_key)
            project_height = project_height.decode('utf-8')

            tentative_project_height = r.get(tentative_bh_key)
            tentative_project_height = tentative_project_height.decode('utf-8')

            project_height = int(project_height)
            tentative_height_project = int(tentative_project_height)
            if project_height < tentative_height_project:
                count+=1
                result = r.set(block_height_key,tentative_project_height)
                console.log(f"[bold blue]Force pushed {result} project blockHeight from {project_height} to height:[bold blue] [white]{tentative_project_height}[white]\n")
            else:
                console.log(f"No need to force-push height ahead for project: {project}")
        except Exception as exc:
            print(f"Error: {str(exc)} | project: {project}")
    if count == 0:
       console.log("No projects required force pushing heights")


if __name__ == '__main__':
    app()
