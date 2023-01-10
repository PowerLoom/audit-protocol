def get_epoch_submissions_htable_key(project_id, epoch_end):
    """
    projectid_epoch_submissions_htable -> instance ID - > JSON {snapshotCID:,  submittedTS:}
    }
    """
    return f'projectID:{project_id}:{epoch_end}:centralizedConsensus:epochSubmissions'


def get_epoch_submission_schedule_key(project_id, epoch_end):
    # {'start': , 'end': }
    return f'projectID:{project_id}:{epoch_end}:centralizedConsensus:submissionSchedule'


def get_project_finalized_epochs_key(project_id):
    return f'projectID:{project_id}:centralizedConsensus:finalizedEpochs'


def get_project_registered_peers_set_key(project_id):
    return f'projectID:{project_id}:centralizedConsensus:peers'


def get_project_epoch_specific_accepted_peers_key(project_id, epoch_end):
    return f'projectID:{project_id}:{epoch_end}:centralizedConsensus:acceptedPeers'


def get_epoch_generator_last_epoch():
    return "epochGenerator:lastEpoch"

def get_epoch_generator_epoch_history():
    return "epochGenerator:epochHistory"

def get_snapshotter_issues_reported_key(snapshotter_id):
    return f'snapshotterData:{snapshotter_id}:issuesReported'