def get_epoch_submissions_htable_key(project_id, epoch_end):
    """
    projectid_epoch_submissions_htable -> instance ID - > JSON {snapshotCID:,  submittedTS:}
    }
    """
    return f'projectID:{project_id}:{epoch_end}:centralizedConsensus:epochSubmissions'


def get_epoch_submission_schedule_key(project_id, epoch_end):
    # {'start': , 'end': }
    return f'projectID:{project_id}:{epoch_end}:centralizedConsensus:submissionSchedule'


def get_project_registered_peers_set_key(project_id):
    return f'projectID:{project_id}:centralizedConsensus:peers'


def get_project_epoch_specific_accepted_peers_key(project_id, epoch_end):
    return f'projectID:{project_id}:{epoch_end}:centralizedConsensus:acceptedPeers'
