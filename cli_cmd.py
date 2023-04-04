import bz2
import io
import pydantic
import typer
import concurrent.futures
from itertools import cycle
from textwrap import wrap
from typing import Dict, Optional
from collections import defaultdict
from settings_model import Settings
from utils import redis_conn
from utils import redis_keys
from data_models import AggregateState, DailyStatSummaryState, PairsSummaryState, ProtocolState, ProjectSpecificProtocolState, ProjectDAGChainSegmentMetadata, PruningCycleForProjectDetails, PruningCycleOverallDetails, TimeSeriesIndex, TokensSummaryState
from config import settings
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


@app.command()
def projectStatus(namespace: str = typer.Option("", "--namespace"), projectId: str = typer.Option(None, "--projectId")):
    r = redis.Redis(**REDIS_CONN_CONF, single_connection_client=True)

    index_status = None
    key = 'projects:IndexStatus'
    if projectId:
        index_status = r.hget(key, projectId)
        if not index_status:
            console.log(f"\n[bold red]Project is not indexed[bold red]: \n{projectId}\n")
            return
        index_status = [{projectId: index_status.decode('utf-8')}]
    else:
        index_status = r.hgetall(key)
        if not index_status:
            console.log("\n[bold red]Indexes map doesn't exist [bold red]: 'projects:IndexStatus'\n")
            return
        index_status = dict(filter(lambda elem: namespace in elem[0].decode('utf-8'), index_status.items()))
        index_status = [{k.decode('utf-8'): v.decode('utf-8')} for k, v in index_status.items()]

    table = Table(show_header=True, header_style="bold magenta")
    table.add_column("ProjectId", justify="center")
    table.add_column("First Epoch Start Height", justify="center")
    table.add_column("Current Epoch End Height", justify="center")
    table.add_column("Dag chain issues", justify="center")
    table.add_column("Epochs pending finalization", justify="center")

    for project_indexes in index_status:
        k, v = project_indexes.popitem()
        v = json.loads(v)

        chain_issues_key = f"projectID:{k}:dagChainGaps"
        chain_issues = r.zrangebyscore(
            chain_issues_key,
            min='-inf',
            max='+inf',
            withscores=True)

        finalized_height_key = f"projectID:{k}:blockHeight"
        finalized_height=r.get(finalized_height_key)

        payload_cids_key = f"projectID:{k}:payloadCids"
        payload_cids = r.zrangebyscore(
            payload_cids_key,
            min=finalized_height,
            max='+inf',
            withscores=True)
        unfinalized_epochs = int(len(payload_cids)/2)

        table.add_row(
            Text(k, justify="left", overflow="ellipsis"),
            str(v["startSourceChainHeight"]),
            str(v["currentSourceChainHeight"]),
            str(int(len(chain_issues)/2)),
            str(unfinalized_epochs)
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


def export_project_state(project_id: str, r: redis.Redis):
    console.log(f'Exporting state for project {project_id}')
    local_exceptions = list()
    project_specific_state = ProjectSpecificProtocolState.construct()
    try:
        project_specific_state.lastFinalizedDAgBlockHeight = int(r.get(redis_keys.get_block_height_key(project_id)))
    except Exception as e:
        local_exceptions.append({'lastFinalizedDAgBlockHeight': str(e)})
    else:
        console.log('\t Exported lastFinalizedDAgBlockHeight')

    try:
        project_specific_state.firstEpochEndHeight = int(r.get(redis_keys.get_project_first_epoch_end_height(project_id)))
    except Exception as e:
        local_exceptions.append({'firstEpochEndHeight': str(e)})
        if 'Summary' or 'Daily' in project_id:
            project_specific_state.firstEpochEndHeight = 0
    else:
        console.log('\t Exported firstEpochEndHeight')

    try:
        project_specific_state.epochSize = int(r.get(redis_keys.get_project_epoch_size(project_id)))
    except Exception as e:
        local_exceptions.append({'epochSize': str(e)})
        if 'Summary' or 'Daily' in project_id:
            project_specific_state.epochSize = 0
    else:
        console.log('\t Exported epochSize')

    project_dag_cids = r.zrangebyscore(
        name=redis_keys.get_dag_cids_key(project_id),
        min='-inf',
        max='+inf',
        withscores=True,
    )

    if len(project_dag_cids) > 0:
        project_specific_state.dagCidsZset = {int(k[1]): k[0] for k in project_dag_cids}
        console.log('\t Exported dagCidsZset of length ', len(project_specific_state.dagCidsZset))
    else:
        project_specific_state.dagCidsZset = dict()
        local_exceptions.append({'dagCidsZset': 'empty'})

    snapshot_cids = r.zrangebyscore(
        name=redis_keys.get_payload_cids_key(project_id),
        min='-inf',
        max='+inf',
        withscores=True,
    )
    if len(snapshot_cids) > 0:
        project_specific_state.snapshotCidsZset = {int(k[1]): k[0] for k in snapshot_cids}
        console.log('\t Exported snapshotCidsZset of length ', len(project_specific_state.snapshotCidsZset))
    else:
        project_specific_state.snapshotCidsZset = dict()
        local_exceptions.append({'snapshotCidsZset': 'empty'})

    project_dag_segments = r.hgetall(redis_keys.get_project_dag_segments_key(project_id))
    if len(project_dag_segments) > 0:
        project_specific_state.dagSegments = {k: ProjectDAGChainSegmentMetadata.parse_raw(v) for k, v in project_dag_segments.items()}
    else:
        project_specific_state.dagSegments = dict()
        local_exceptions.append({'dagSegments': 'empty'})

    # time series indexes
    # TODO: use the cache:indexesRequested set to get the list of indexes to export for a project since the project might not require all indexes
    time_series_index_identifiers = ['0', '24h', '7d']
    project_specific_state.timeSeriesData = dict()
    for time_period in time_series_index_identifiers:
        idx_head_key = redis_keys.get_sliding_window_cache_head_marker(project_id, time_period)
        idx_tail_key = redis_keys.get_sliding_window_cache_tail_marker(project_id, time_period)
        # this takes care of only head marker being set for '0' indexes to the latest snapshot's epoch max height
        idx_head = r.get(idx_head_key)
        if idx_head:
            idx_head = int(idx_head)  
            idx_tail = r.get(idx_tail_key)
            idx_tail = int(idx_tail) if idx_tail else None
            project_specific_state.timeSeriesData[time_period] = TimeSeriesIndex(
                head=idx_head,
                tail=idx_tail,
            )
            console.log('\t Exported timeSeriesData for ', time_period, ' with head ', idx_head, ' and tail ', idx_tail)
        else:
            local_exceptions.append({f'timeSeriesData.{time_period}': 'empty'})
    return project_specific_state, local_exceptions


def export_pruning_cycle_details(cycle_id: str, r: redis.Redis):
    console.log(f'Exporting pruning cycle details for cycle {cycle_id}')
    cycle_state = dict()
    local_exceptions = list()
    detailed_pruning_run_stats = r.hgetall(redis_keys.get_specific_pruning_cycle_run_information(cycle_id))
    if len(detailed_pruning_run_stats) > 0:
        cycle_state = {k: PruningCycleForProjectDetails.parse_raw(v) for k, v in detailed_pruning_run_stats.items()}
        console.log('Exported pruning cycle details for cycle ', cycle_id)
    else:
        local_exceptions.append({'detailed_pruning_run_stats': 'empty'})
    return cycle_state, local_exceptions


def export_aggregates(r: redis.Redis):
    aggregates = AggregateState.construct()
    # aggregate: pairs summary keys
    single_pair_cached_summary_key_pattern = 'uniswap:pairContract:'+settings.pooler_namespace+':*:contractV2PairCachedData'
    pairs_summary_project_id_key = redis_keys.get_uniswap_pairs_summary_snapshot_project_id(settings.pooler_namespace)
    pairs_summary_zset_key = redis_keys.get_uniswap_pair_snapshot_summary_zset(settings.pooler_namespace)
    pairs_summary_timestamp_map_zset_key = redis_keys.get_uniswap_pair_snapshot_timestamp_zset(settings.pooler_namespace)
    # pairs_summary_at_block_height_key_pattern = 'uniswap:V2PairsSummarySnapshot:'+settings.pooler_namespace+':snapshot:*'
    pairs_summary_last_cached_blockheight_key = redis_keys.get_uniswap_pair_snapshot_last_block_height(settings.pooler_namespace)
    pair_contract_recent_logs_key_pattern = f'uniswap:pairContract:{settings.pooler_namespace}:*:recentLogs'
    pair_contract_sliding_window_data_key_pattern = f'uniswap:pairContract:{settings.pooler_namespace}:*:slidingWindowData'
    # aggregate: daily stats keys
    pair_daily_stats_project_id_key = redis_keys.get_uniswap_pairs_v2_daily_snapshot_project_id(settings.pooler_namespace)
    pairs_daily_stats_zset_key = redis_keys.get_uniswap_pair_daily_stats_snapshot_zset(settings.pooler_namespace)
    daily_stats_cache_key_pattern = f'uniswap:pairContract:{settings.pooler_namespace}:*:dailyCache'
    # aggregate: token stats and other keys
    token_summary_project_id_key = redis_keys.get_uniswap_token_summary_snapshot_project_id(settings.pooler_namespace)
    pair_contract_tokens_data_key_pattern = f'uniswap:pairContract:{settings.pooler_namespace}:*:PairContractTokensData'  # hashtable
    token_price_history_key_pattern = f'uniswap:tokenInfo:{settings.pooler_namespace}:*:priceHistory'
    pair_token_addresses_key_pattern = f'uniswap:pairContract:{settings.pooler_namespace}:*:PairContractTokensAddresses'
    token_price_at_block_height_key_pattern = f'uniswap:pairContract:{settings.pooler_namespace}:*:cachedPairBlockHeightTokenPrice'
    token_summary_snapshots_zset_key = redis_keys.get_uniswap_token_summary_snapshot_zset(settings.pooler_namespace)

    single_pairs_summary_map = dict()
    for k in r.scan_iter(match=single_pair_cached_summary_key_pattern):
        pair_contract_address = k.split(':')[3]
        single_pairs_summary_map[pair_contract_address] = r.get(k)
    aggregates.singlePairsSummary = single_pairs_summary_map
    console.log('Exported cached summary data for ', len(single_pairs_summary_map), ' pairs')
    pairs_summary_state = PairsSummaryState.construct()
    try:
        pairs_summary_state.lastCachedBlock = int(r.get(pairs_summary_last_cached_blockheight_key))
    except TypeError:
        pairs_summary_state.lastCachedBlock = 0
        console.log('No last cached block height found for pairs summary')
    else:
        console.log('Exported last cached block height for pairs summary: ', pairs_summary_state.lastCachedBlock)
    pairs_summary_state.timestampSummaryMap = {
        k[1]: k[0] for k in r.zrangebyscore(
            pairs_summary_timestamp_map_zset_key, '-inf', '+inf', withscores=True, score_cast_func=int
        )
    }
    console.log('Exported ', len(pairs_summary_state.timestampSummaryMap), ' timestamp summary map entries for pairs summary')
    pairs_summary_state.blockHeightSummaryMap = {
        k[1]: k[0] for k in r.zrangebyscore(
            pairs_summary_zset_key, '-inf', '+inf', withscores=True, score_cast_func=int
        )
    }
    console.log('Exported ', len(pairs_summary_state.blockHeightSummaryMap), ' block height summary map entries for pairs summary')
    pairs_summary_state.tentativeBlockHeight = int(r.get(redis_keys.get_tentative_block_height_key(pairs_summary_project_id_key)))
    console.log('Exported tentative block height for pairs summary: ', pairs_summary_state.tentativeBlockHeight)
    pairs_summary_state.sildingWindowData = dict()
    for sliding_window_data_contract_key in r.scan_iter(match=pair_contract_sliding_window_data_key_pattern, count=50):
        pair_contract_address = sliding_window_data_contract_key.split(':')[3]
        pairs_summary_state.sildingWindowData.update({pair_contract_address: r.get(sliding_window_data_contract_key)})
    console.log('Exported ', len(pairs_summary_state.sildingWindowData), ' sliding window data entries for pairs summary')
    pairs_summary_state.recentLogs = dict()
    for recent_logs_contract_key in r.scan_iter(match=pair_contract_recent_logs_key_pattern, count=50):
        pair_contract_address = recent_logs_contract_key.split(':')[3]
        pairs_summary_state.recentLogs.update({pair_contract_address: r.get(recent_logs_contract_key)})
    console.log('Exported ', len(pairs_summary_state.recentLogs), ' recent logs entries for pairs summary')
    aggregates.pairsSummary = pairs_summary_state

    daily_stats_summary_state = DailyStatSummaryState.construct()
    daily_stats_summary_state.tentativeBlockHeight = int(r.get(redis_keys.get_tentative_block_height_key(pair_daily_stats_project_id_key)))
    console.log('Exported tentative block height for daily stats summary: ', daily_stats_summary_state.tentativeBlockHeight)
    daily_stats_summary_state.blockHeightSummaryMap = {
        k[1]: k[0] for k in r.zrangebyscore(
            pairs_daily_stats_zset_key, '-inf', '+inf', withscores=True, score_cast_func=int
        )
    }
    console.log('Exported ', len(daily_stats_summary_state.blockHeightSummaryMap), ' block height summary map entries for daily stats summary')
    daily_stats_summary_state.dailyCache = dict()
    for daily_cache_contract_key in r.scan_iter(match=daily_stats_cache_key_pattern, count=50):
        pair_contract_address = daily_cache_contract_key.split(':')[3]
        daily_stats_summary_state.dailyCache[pair_contract_address] = {
            k[1]: k[0] for k in r.zrangebyscore(
                daily_cache_contract_key, '-inf', '+inf', withscores=True, score_cast_func=int
            )
        }
    console.log('Exported ', len(daily_stats_summary_state.dailyCache), ' daily cache entries for daily stats summary')
    aggregates.dailyStatsSummary = daily_stats_summary_state

    token_summary_state = TokensSummaryState.construct()
    token_summary_state.tentativeBlockHeight = int(r.get(redis_keys.get_tentative_block_height_key(token_summary_project_id_key)))
    token_summary_state.contractTokenMetadata = dict()
    for pair_contract_tokens_metadata_key in r.scan_iter(match=pair_contract_tokens_data_key_pattern, count=50):
        pair_contract_address = pair_contract_tokens_metadata_key.split(':')[3]
        token_summary_state.contractTokenMetadata.update({pair_contract_address: r.hgetall(pair_contract_tokens_metadata_key)})
    
    console.log('Exported ', len(token_summary_state.contractTokenMetadata), ' token metadata entries for tokens summary')
    token_summary_state.contractTokenPriceHistory = dict()
    for token_contract_price_history_key in r.scan_iter(match=token_price_history_key_pattern, count=50):
        token_contract_address = token_contract_price_history_key.split(':')[3]
        token_summary_state.contractTokenPriceHistory[token_contract_address] = {
            k[1]: k[0] for k in r.zrangebyscore(
                token_contract_price_history_key, '-inf', '+inf', withscores=True, score_cast_func=int
            )
        }
    console.log('Exported ', len(token_summary_state.contractTokenPriceHistory), ' token price history entries for tokens summary')
    token_summary_state.contractTokenAddresses = dict()
    for pair_contract_token_addrs_key in r.scan_iter(match=pair_token_addresses_key_pattern, count=50):
        pair_contract_address = pair_contract_token_addrs_key.split(':')[3]
        token_summary_state.contractTokenAddresses[pair_contract_address] = r.hgetall(pair_contract_token_addrs_key)
    console.log('Exported ', len(token_summary_state.contractTokenAddresses), ' pair contract\'s token addresses entries for tokens summary')
    token_summary_state.pairContractPriceHistory = dict()
    for pair_contract_price_history_zset_key in r.scan_iter(match=token_price_at_block_height_key_pattern, count=50):
        pair_contract_address = pair_contract_price_history_zset_key.split(':')[3]
        token_summary_state.pairContractPriceHistory[pair_contract_address] = {
            k[1]: k[0] for k in r.zrangebyscore(
                pair_contract_price_history_zset_key, '-inf', '+inf', withscores=True, score_cast_func=int
            )
        }
    console.log('Exported ', len(token_summary_state.pairContractPriceHistory), ' pair contract\'s price history entries for tokens summary')
    token_summary_state.blockHeightSummaryMap = {
        k[1]: k[0] for k in r.zrangebyscore(
            token_summary_snapshots_zset_key, '-inf', '+inf', withscores=True, score_cast_func=int
        )
    }
    console.log('Exported ', len(token_summary_state.blockHeightSummaryMap), ' block height summary map entries for tokens summary')
    aggregates.tokensSummary = token_summary_state
    return aggregates


@app.command(name='exportState')
def export_state():
    r = redis.Redis(**REDIS_CONN_CONF, max_connections=20, decode_responses=True)
    # TODO: incrementally byte stream the final state JSON into the bz2 compressor to achieve better memory efficiency
    # references: https://docs.python.org/3/glossary.html#term-bytes-like-object
    # https://docs.python.org/3/library/io.html#binary-i-o
    # https://pymotw.com/3/bz2/#incremental-compression-and-decompression
    # compressor = bz2.BZ2Compressor()
    console.log("Exporting state to file")
    state = ProtocolState.construct()
    exceptions = defaultdict(list)  # collect excs while iterating over projects
    exceptions['project'] = dict()
    exceptions['pruning'] = dict()

    state.projectSpecificStates = dict()
    all_projects = r.smembers(redis_keys.get_stored_project_ids_key())
    with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
        future_to_project = {executor.submit(export_project_state, project_id, r): project_id for project_id in all_projects}
        for future in concurrent.futures.as_completed(future_to_project):
            project_id = future_to_project[future]
            try:
                project_specific_state, local_exceptions = future.result()
            except Exception as exc:
                exceptions['project'].update({project_id: str(exc)})
            else:
                if local_exceptions:
                    exceptions['project'].update({project_id: local_exceptions})
                state.projectSpecificStates[project_id] = project_specific_state
    
    pruning_project_status = r.hgetall(redis_keys.get_pruning_status_key())
    if len(pruning_project_status) > 0:
        state.pruningProjectStatus = {k: int(v) for k, v in pruning_project_status.items()}
        console.log('Backed up last pruned DAG height for projects: ', ', '.join(state.pruningProjectStatus.keys()))
    else:
        state.pruningProjectStatus = dict()
        exceptions['pruning'].update({'pruningProjectStatus': 'empty'})

    pruning_cycle_run_stats = r.zrangebyscore(
        name=redis_keys.get_all_pruning_cycles_status_key(),
        min='-inf',
        max='+inf',
        withscores=True,
    )
    if len(pruning_cycle_run_stats) > 0:
        state.pruningCycleRunStatus = {int(k[1]): PruningCycleOverallDetails.parse_raw(k[0]) for k in pruning_cycle_run_stats}
    else:
        state.pruningCycleRunStatus = dict()
        exceptions['pruning'].update({'pruningCycleRunStats': 'empty'})

    exceptions['pruning'].update({'detailedRunStats': dict()})
    state.pruningProjectDetails = dict()
    cycles_collected = [k.pruningCycleID for k in state.pruningCycleRunStatus.values()]
    with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
        future_to_cycle = {executor.submit(export_pruning_cycle_details, cycle_id, r): cycle_id for cycle_id in cycles_collected}
        for future in concurrent.futures.as_completed(future_to_cycle):
            cycle_id = future_to_cycle[future]
            try:
                cycle_specific_state, local_exceptions = future.result()
            except Exception as exc:
                exceptions['pruning']['detailedRunStats'].update({cycle_id: str(exc)})
            else:
                state.pruningProjectDetails[cycle_id] = cycle_specific_state
                exceptions['pruning']['detailedRunStats'].update({cycle_id: local_exceptions})
    state.aggregates = export_aggregates(r)
    state_json = state.json()
    console.log('Collected exceptions while processing: ')
    console.print_json(json.dumps(exceptions))
    with bz2.open('state.json.bz2', 'wb') as f:
        with io.TextIOWrapper(f, encoding='utf-8') as enc:
            enc.write(state_json)
    console.log('Exported state.json.bz2')


def load_project_specific_state(project_id: str, project_state: ProjectSpecificProtocolState, r: redis.Redis):
    console.log('Loading project specific state for project ', project_id)
    r.sadd(redis_keys.get_stored_project_ids_key(), project_id)
    console.log('\tLoaded project id in all projects set')
    r.set(redis_keys.get_block_height_key(project_id), project_state.lastFinalizedDAgBlockHeight)
    console.log('\tLoaded last finalized DAG block height')
    r.set(redis_keys.get_project_first_epoch_end_height(project_id), project_state.firstEpochEndHeight)
    console.log('\tLoaded first epoch end height')
    r.set(redis_keys.get_project_epoch_size(project_id), project_state.epochSize)
    console.log('\tLoaded epoch size')
    # transform to element-score mapping for easy addition to zset
    dag_cid_map = {v: k for k, v in project_state.dagCidsZset.items()}
    r.zadd(name=redis_keys.get_dag_cids_key(project_id), mapping=dag_cid_map)
    console.log('\tLoaded DAG CIDs')
    snapshot_cid_map = {v: k for k, v in project_state.snapshotCidsZset.items()}
    r.zadd(name=redis_keys.get_payload_cids_key(project_id), mapping=snapshot_cid_map)
    console.log('\tLoaded snapshot CIDs')
    if project_state.dagSegments:
        r.hset(
            name=redis_keys.get_project_dag_segments_key(project_id),
            mapping={k: v.json() for k, v in project_state.dagSegments.items()}
        )
    console.log('\tLoaded DAG segments')
    for time_series_id in project_state.timeSeriesData:
        idx_head_key = redis_keys.get_sliding_window_cache_head_marker(project_id, time_series_id)
        idx_tail_key = redis_keys.get_sliding_window_cache_tail_marker(project_id, time_series_id)
        r.set(idx_head_key, project_state.timeSeriesData[time_series_id].head)
        if project_state.timeSeriesData[time_series_id].tail is not None:
            r.set(idx_tail_key, project_state.timeSeriesData[time_series_id].tail)
        console.log(
            'Loaded head', project_state.timeSeriesData[time_series_id].head, ' and tail ', project_state.timeSeriesData[time_series_id].tail, 
            ' for time series ', time_series_id, ' for project ', project_id
        )

def load_pruning_detailed_stats(cycle_id: int, project_id_cycle_details_map: Dict[str, PruningCycleForProjectDetails], r: redis.Redis):
    console.log('Loading pruning cycle details for cycle ', cycle_id)
    if project_id_cycle_details_map:
        r.hset(
            name=redis_keys.get_specific_pruning_cycle_run_information(cycle_id),
            mapping={k: v.json() for k, v in project_id_cycle_details_map.items()}
        )
        console.log('\tLoaded pruning cycle details for cycle ', cycle_id, ' with ', len(project_id_cycle_details_map), ' projects')

def load_aggregates(aggregates: AggregateState, r: redis.Redis):
    # pair contract summary stuff
    for pair_contract_address in aggregates.singlePairsSummary:
        r.set(
            redis_keys.get_uniswap_pair_contract_V2_pair_data(pair_contract_address, settings.pooler_namespace),
            aggregates.singlePairsSummary[pair_contract_address]
        )
        console.log('Loaded single pair V2 summary for ', pair_contract_address)
    pairs_summary_project_id = redis_keys.get_uniswap_pairs_summary_snapshot_project_id(settings.pooler_namespace)
    r.set(
        redis_keys.get_tentative_block_height_key(pairs_summary_project_id),
        aggregates.pairsSummary.tentativeBlockHeight
    )
    console.log('Loaded tentative block height for V2 pairs summary ', aggregates.pairsSummary.tentativeBlockHeight)
    pairs_summary_last_cached_blockheight_key = redis_keys.get_uniswap_pair_snapshot_last_block_height(settings.pooler_namespace)
    r.set(
        pairs_summary_last_cached_blockheight_key,
        aggregates.pairsSummary.lastCachedBlock
    )
    console.log('Loaded last cached block height for V2 pairs summary ', aggregates.pairsSummary.lastCachedBlock)
    pairs_summary_zset_key = redis_keys.get_uniswap_pair_snapshot_summary_zset(settings.pooler_namespace)
    r.zadd(
        name=pairs_summary_zset_key,
        mapping={
            v: k for k, v in aggregates.pairsSummary.blockHeightSummaryMap.items()
        }
    )
    console.log('Loaded V2 pairs blockheight mapped summary zset with ', len(aggregates.pairsSummary.blockHeightSummaryMap), ' elements')
    pairs_summary_timestamp_map_zset_key = redis_keys.get_uniswap_pair_snapshot_timestamp_zset(settings.pooler_namespace)
    r.zadd(
        name=pairs_summary_timestamp_map_zset_key,
        mapping={
            v: k for k, v in aggregates.pairsSummary.timestampSummaryMap.items()
        }
    )
    console.log('Loaded V2 pairs timestamp mapped summary zset with ', len(aggregates.pairsSummary.timestampSummaryMap), ' elements')
    for pair_contract_address in aggregates.pairsSummary.recentLogs:
        r.set(
            redis_keys.get_uniswap_pair_cached_recent_logs(pair_contract_address, settings.pooler_namespace),
            aggregates.pairsSummary.recentLogs[pair_contract_address]
        )
        console.log('Loaded V2 recent logs for ', pair_contract_address)
    for pair_contract_address in aggregates.pairsSummary.sildingWindowData:
        r.set(
            redis_keys.get_uniswap_pair_cache_sliding_window_data(pair_contract_address, settings.pooler_namespace),
            aggregates.pairsSummary.sildingWindowData[pair_contract_address]
        )
        console.log('Loaded V2 sliding window data for ', pair_contract_address)
    # daily stats
    pair_daily_stats_project_id_key = redis_keys.get_uniswap_pairs_v2_daily_snapshot_project_id(settings.pooler_namespace)
    r.set(
        redis_keys.get_tentative_block_height_key(pair_daily_stats_project_id_key),
        aggregates.dailyStatsSummary.tentativeBlockHeight
    )
    console.log('Loaded tentative block height for V2 daily stats summary ', aggregates.dailyStatsSummary.tentativeBlockHeight)
    pairs_daily_stats_zset_key = redis_keys.get_uniswap_pair_daily_stats_snapshot_zset(settings.pooler_namespace)
    r.zadd(
        name=pairs_daily_stats_zset_key,
        mapping={
            v: k for k, v in aggregates.dailyStatsSummary.blockHeightSummaryMap.items()
        }
    )
    console.log('Loaded V2 pairs daily stats blockheight mapped summary zset with ', len(aggregates.dailyStatsSummary.blockHeightSummaryMap), ' elements')
    for pair_contract_address, daily_cache_mapping in aggregates.dailyStatsSummary.dailyCache.items():
        pair_daily_cache_daily_stats_key = redis_keys.get_uniswap_pair_cache_daily_stats(pair_contract_address, settings.pooler_namespace)
        r.zadd(
            name=pair_daily_cache_daily_stats_key,
            mapping={
                v: k for k, v in daily_cache_mapping.items()
            }
        )
        console.log('Loaded V2 daily stats cache for ', pair_contract_address, ' with ', len(daily_cache_mapping), ' elements')
    # token summaries
    token_summary_project_id_key = redis_keys.get_uniswap_token_summary_snapshot_zset(settings.pooler_namespace)
    r.set(
        redis_keys.get_tentative_block_height_key(token_summary_project_id_key),
        aggregates.tokensSummary.tentativeBlockHeight
    )
    console.log('Loaded tentative block height for V2 token summary ', aggregates.tokensSummary.tentativeBlockHeight)
    for pair_contract_address, metadata_dict in aggregates.tokensSummary.contractTokenMetadata.items():
        if metadata_dict:
            pair_contract_tokens_data_key = f'uniswap:pairContract:{settings.pooler_namespace}:{pair_contract_address}:PairContractTokensData'  # hashtable
            r.hset(
                name=pair_contract_tokens_data_key,
                mapping=metadata_dict
            )
            console.log('Loaded V2 token metadata for ', pair_contract_address)
    for pair_contract_address, price_history_mapping in aggregates.tokensSummary.contractTokenPriceHistory.items():
        token_price_history_key = f'uniswap:tokenInfo:{settings.pooler_namespace}:{pair_contract_address}:priceHistory'
        r.zadd(
            name=token_price_history_key,
            mapping={
                v: k for k, v in price_history_mapping.items()
            }
        )
        console.log('Loaded V2 token price history for ', pair_contract_address, ' with ', len(price_history_mapping), ' elements')
    for pair_contract_address, token_addresses in aggregates.tokensSummary.contractTokenAddresses.items():
        if token_addresses:
            pair_token_addresses_key = f'uniswap:pairContract:{settings.pooler_namespace}:{pair_contract_address}:PairContractTokensAddresses'
            r.hset(
                name=pair_token_addresses_key,
                mapping=token_addresses
            )
            console.log('Loaded V2 token addresses for ', pair_contract_address, ' with mapping ', token_addresses)
    for pair_contract_address, pair_contract_pricing_map in aggregates.tokensSummary.pairContractPriceHistory.items():
        pair_contract_price_at_block_height_key = f'uniswap:pairContract:{settings.pooler_namespace}:{pair_contract_address}:cachedPairBlockHeightTokenPrice'
        r.zadd(
            name=pair_contract_price_at_block_height_key,
            mapping={
                v: k for k, v in pair_contract_pricing_map.items()
            }
        )
        console.log('Loaded V2 pair contract pricing history for ', pair_contract_address, ' with ', len(pair_contract_pricing_map), ' elements')
    token_summary_snapshots_zset_key = redis_keys.get_uniswap_token_summary_snapshot_zset(settings.pooler_namespace)
    r.zadd(
        name=token_summary_snapshots_zset_key,
        mapping={
            v: k for k, v in aggregates.tokensSummary.blockHeightSummaryMap.items()
        }
    )
    console.log('Loaded V2 token summary blockheight mapped summary zset with ', len(aggregates.tokensSummary.blockHeightSummaryMap), ' elements')
    

@app.command(name='loadState')
def load_state_from_file(file_name: Optional[str] = typer.Argument('state.json.bz2')):
    r = redis.Redis(**REDIS_CONN_CONF, max_connections=20, decode_responses=True)
    console.log('Loading state from file ', file_name)
    with bz2.open(file_name, 'rb') as f:
        state_json = f.read()
    try:
        state = ProtocolState.parse_raw(state_json)
    except pydantic.ValidationError as e:
        console.log('Error while parsing state file: ', e)
        with open('state_parse_error.json', 'w') as f:
            json.dump(e.errors(), f)
        return
    console.log(state.pruningCycleRunStatus)
    with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
        future_to_project = {executor.submit(load_project_specific_state, project_id, project_state, r): project_id for project_id, project_state in state.projectSpecificStates.items()}
        for future in concurrent.futures.as_completed(future_to_project):
            project_id = future_to_project[future]
            try:
                future.result()
            except Exception as exc:
                console.log('Error while loading project state: ', project_id, exc)
            else:
                console.log('Loaded project state: ', project_id)
    if state.pruningProjectStatus:
        r.hset(
            name=redis_keys.get_pruning_status_key(), 
            mapping=state.pruningProjectStatus
        )
        console.log('Loaded pruning project status in key ', redis_keys.get_pruning_status_key())
    r.zadd(
        name=redis_keys.get_all_pruning_cycles_status_key(),
        mapping={v.json(): k for k, v in state.pruningCycleRunStatus.items()}
    )
    console.log('Loaded pruning run cycle status in key ', redis_keys.get_all_pruning_cycles_status_key())
    with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
        future_to_cycle = {executor.submit(load_pruning_detailed_stats, cycle_id, project_id_cycle_details_map, r): cycle_id for cycle_id, project_id_cycle_details_map in state.pruningProjectDetails.items()}
        for future in concurrent.futures.as_completed(future_to_cycle):
            cycle_id = future_to_cycle[future]
            try:
                future.result()
            except Exception as exc:
                console.log('Error while loading pruning cycle details: ', cycle_id, exc)
    load_aggregates(state.aggregates, r)

@app.command()
def skip_pair_projects_verified_heights():
    r = redis.Redis(**REDIS_CONN_CONF, single_connection_client=True)

    print("\nThis command will force-push Summary project's redis state ahead.\n ")
    count = 0

    verification_status_key = f"projects:dagVerificationStatus"
    projects = r.hgetall(verification_status_key)
    console.log("project count:",len(projects))
    for project,verified_height in projects.items():
        project_str = project.decode('utf-8')
        if project_str.find('Snapshot') > 0:
            console.log("Found project which is Summary project",project_str, " and skipping it")
            continue
        block_height_key = f"projectID:{project_str}:blockHeight"
        console.log("Project Id is ",project_str)
        project_height = r.get(block_height_key)
        project_height = project_height.decode('utf-8')
        console.log("Project height is ",project_height)
        project_height = int(project_height)
        if project_height > int(verified_height)+10:
            console.log("difference in height for project %s is %s",project_str, (project_height - int(verified_height)))
            count+=1
            projects[project] = int(verified_height)+4
    if count > 0:
        all([r.hset(verification_status_key, k, v) for k, v in projects.items()])
        console.log("updated project verification heights successfully for %d projects",count)
    else:
        console.log("No need to update project verification heights as all projects have been verified till their current height.")


if __name__ == '__main__':
    app()
