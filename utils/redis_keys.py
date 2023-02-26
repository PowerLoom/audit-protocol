# keep a mapping from payload commit ID to tx hash
def get_payload_commit_key(payload_commit_id: str):
    payload_commit_key = "payloadCommit:{}".format(payload_commit_id)
    return payload_commit_key


def get_project_first_epoch_end_height(project_id):
    return f'projectID:{project_id}:firstEpochEndHeight'


def get_project_epoch_size(project_id):
    return f'projectID:{project_id}:epochSize'


def get_pending_retrieval_requests_key():
    pending_retrieval_requests_key = "pendingRetrievalRequests"
    return pending_retrieval_requests_key


def get_retrieval_request_info_key(request_id: str):
    retrieval_request_info_key = "retrievalRequestInfo:{}".format(request_id)
    return retrieval_request_info_key


def get_pruning_status_key():
    last_pruned_key = "projects:pruningStatus"
    return last_pruned_key

def get_all_pruning_cycles_status_key():
    return 'pruningRunStatus'


def get_specific_pruning_cycle_run_information_pattern():
    return 'pruningProjectDetails:*'

def get_dag_cids_key(project_id: str):
    dag_cids_key = "projectID:{}:Cids".format(project_id)
    return dag_cids_key


def get_block_height_key(project_id: str):
    block_height_key = "projectID:{}:blockHeight".format(project_id)
    return block_height_key


def get_sliding_window_cache_head_marker(project_id: str, time_period: str):
    return f'projectID:{project_id}:slidingCache:{time_period}:head'


def get_sliding_window_cache_tail_marker(project_id: str, time_period: str):
    return f'projectID:{project_id}:slidingCache:{time_period}:tail'


def get_containers_created_key(project_id: str):
    containers_created_key = "projectID:{}:containers".format(project_id)
    return containers_created_key


def get_container_data_key(container_id: str):
    container_data_key = "containerData:{}".format(container_id)
    return container_data_key


def get_retrieval_request_files_key(request_id: str):
    retrieval_request_files_key = "retrievalRequestFiles:{}".format(request_id)
    return retrieval_request_files_key


def get_event_data_key(payload_commit_id: str):
    event_data_key = "eventData:{}".format(payload_commit_id)
    return event_data_key


# a ZSET
# webhook callbacks have arrived against the corresponding tentative block heights, but yet to be included in the
# DAG chain because of some missing callback in the past which will lead to a gap in the DAG chain
def get_pending_blocks_key(project_id: str):
    pending_blocks_key = "projectID:{}:pendingBlocks".format(project_id)
    return pending_blocks_key


def get_last_dag_cid_key(project_id: str):
    last_dag_cid_key = "projectID:{}:lastDagCid".format(project_id)
    return last_dag_cid_key


def get_diff_snapshots_key(project_id: str):
    diff_snapshots_key = "projectID:{}:diffSnapshots".format(project_id)
    return diff_snapshots_key


def get_last_seen_snapshots_key():
    last_seen_snapshots_key = "auditprotocol:lastSeenSnapshots"
    return last_seen_snapshots_key


def get_payload_cids_key(project_id: str):
    payload_cids_key = "projectID:{}:payloadCids".format(project_id)
    return payload_cids_key


def get_pending_block_creation_key(project_id: str):
    pending_block_creation_key = "projectID:{}:pendingBlockCreation".format(project_id)
    return pending_block_creation_key


def get_stored_project_ids_key():
    stored_project_ids_key = "storedProjectIds"
    return stored_project_ids_key


def get_project_dag_segments_key(project_id: str):
    return f'projectID:{project_id}:dagSegments'


def get_payload_commit_id_process_logs_zset_key(project_id, payload_commit_id):
    return f'projectID:{project_id}:payloadCommitID:{payload_commit_id}:processingLogs'


def get_hits_dag_block_key():
    hits_dag_block_key = "hitsDagBlock"
    return hits_dag_block_key


def get_last_snapshot_cid_key(project_id: str):
    last_snapshot_cid_key = "projectID:{}:lastSnapshotCid".format(project_id)
    return last_snapshot_cid_key


def get_tentative_block_height_key(project_id: str):
    tentative_block_height_key = "projectID:{}:tentativeBlockHeight".format(project_id)
    return tentative_block_height_key


def get_job_status_key(snapshot_cid: str):
    job_status_key = "jobStatus:{}".format(snapshot_cid)
    return job_status_key


def get_diff_rules_key(project_id: str):
    diff_rules_key = "projectID:{}:diffRules".format(project_id)
    return diff_rules_key


def get_pending_transactions_key(project_id: str):
    pending_transaction_key = "projectID:{}:pendingTransactions".format(project_id)
    return pending_transaction_key


# will be used to store input data against a transaction for later re-processing if required
def get_pending_tx_input_data_key(tx_hash: str):
    return f'txHash:{tx_hash}:inputData'


def get_discarded_transactions_key(project_id: str):
    discarded_transactions_key = "projectID:{}:discardedTransactions".format(project_id)
    return discarded_transactions_key


def get_live_spans_key(project_id: str, span_id: str):
    live_spans_key = "projectID:{}:liveSpans:{}".format(project_id, span_id)
    return live_spans_key


def get_cached_containers_key(container_id: str):
    cached_containers_key = "cachedContainers:{}".format(container_id)
    return cached_containers_key


def get_projects_registered_for_cache_indexing_key():
    return 'cache:indexesRequested'

def get_uniswap_pair_contract_tokens_data(pair_address, pooler_namespace):
    return 'uniswap:pairContract:'+pooler_namespace+':{}:PairContractMetaData'.format(pair_address)


def get_uniswap_pair_contract_V2_pair_data(pair_address, pooler_namespace):
    return 'uniswap:pairContract:'+pooler_namespace+':{}:contractV2PairCachedData'.format(pair_address)


def get_uniswap_pair_snapshot_last_block_height(pooler_namespace):
    return 'uniswap:V2PairsSummarySnapshot:'+pooler_namespace+':lastBlockHeight'


def get_uniswap_pair_snapshot_summary_zset(pooler_namespace):
    return 'uniswap:V2PairsSummarySnapshot:'+pooler_namespace+':snapshotsZset'

def get_uniswap_pair_snapshot_payload_at_blockheight(block_height, pooler_namespace):
    return 'uniswap:V2PairsSummarySnapshot:'+pooler_namespace+f':snapshot:{block_height}'

def get_uniswap_pair_daily_stats_snapshot_zset(pooler_namespace):
    return 'uniswap:V2DailyStatsSnapshot:'+pooler_namespace+':snapshotsZset'

def get_uniswap_pair_daily_stats_payload_at_blockheight(block_height, pooler_namespace):
    return 'uniswap:V2DailyStatsSnapshot:'+pooler_namespace+f':snapshot:{block_height}'


def get_uniswap_pairs_summary_snapshot_project_id(pooler_namespace):
    return 'uniswap_V2PairsSummarySnapshot_'+pooler_namespace

def get_uniswap_pairs_v2_daily_snapshot_project_id(pooler_namespace):
    return 'uniswap_V2DailyStatsSnapshot_'+pooler_namespace

def get_uniswap_pair_snapshot_timestamp_zset(pooler_namespace):
    return 'uniswap:V2PairsSummarySnapshot:'+pooler_namespace+':snapshotTimestampZset'


def get_uniswap_pair_cached_token_price(pair_symbol, pooler_namespace):
    return 'uniswap:pairContract:'+pooler_namespace+':{}:cachedPairPrice'.format(pair_symbol)


def get_uniswap_pair_cached_recent_logs(pair_address, pooler_namespace):
    return 'uniswap:pairContract:'+pooler_namespace+':{}:recentLogs'.format(pair_address)


def get_uniswap_pair_cache_daily_stats(pair_address, pooler_namespace):
    return 'uniswap:pairContract:'+pooler_namespace+':{}:dailyCache'.format(pair_address)


def get_uniswap_pair_cache_sliding_window_data(pair_address, pooler_namespace):
    return 'uniswap:pairContract:'+pooler_namespace+':{}:slidingWindowData'.format(pair_address)

def get_uniswap_projects_dag_verifier_status():
    return "projects:dagVerificationStatus"

def get_projects_registered_for_cache_indexing_key_with_namespace(pooler_namespace):
    return f'cache:indexesRequested:{pooler_namespace}'