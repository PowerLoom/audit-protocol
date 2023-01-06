def get_epoch_submissions_htable_key(project_id, epoch_end):
    """
    projectid_epoch_submissions_htable -> instance ID - > JSON {snapshotCID:,  submittedTS:}
    }
    """
    return f'projectID:{project_id}:{epoch_end}:centralizedConsensus:epochSubmissions'


def get_epoch_submission_schedule_key(project_id, epoch_end):
    # {'start': , 'end': }
    return f'projectID:{project_id}:{epoch_end}:centralizedConsensus:submissionSchedule'


def get_epoch_project_status_key(project_id, epoch_end):
    return f'projectID:{project_id}:centralizedConsensus:{epoch_end}:epochStatus'

def get_epoch_finalized_projects_key(epoch_end):
    return f'projectID:*:centralizedConsensus:{epoch_end}:epochStatus'

def get_project_registered_peers_set_key(project_id):
    return f'projectID:{project_id}:centralizedConsensus:peers'


def get_project_epoch_specific_accepted_peers_key(project_id, epoch_end):
    return f'projectID:{project_id}:{epoch_end}:centralizedConsensus:acceptedPeers'


def get_project_epochs(project_id):
    # Using :centralizedConsensus:epochSubmissions to avoid processing of duplicate keys under projectID:project_id:epoch_end*]
    return "projectID:"+project_id+":[0-9]*:centralizedConsensus:epochSubmissions"


def get_project_ids():
    # Using :centralizedConsensus:peers to avoid fetching duplicate keys under projectID:project_id:*
    return "projectID:*:centralizedConsensus:peers"

def get_epoch_generator_last_epoch():
    return "epochGenerator:lastEpoch"

def get_epoch_generator_epoch_history():
    return "epochGenerator:epochHistory"

def get_snapshotter_issues_reported_key(snapshotter_id):
    return f'snapshotterData:{snapshotter_id}:issuesReported'