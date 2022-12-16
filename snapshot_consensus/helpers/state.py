from snapshot_consensus.data_models import (
    SubmissionDataStoreEntry, SnapshotSubmission, SubmissionSchedule, SubmissionAcceptanceStatus, SnapshotBase
)
from snapshot_consensus.conf import settings
from .redis_keys import *
from typing import Tuple, Union
from redis import asyncio as aioredis
import time
import json


async def get_submission_schedule(
        project_id,
        epoch_end,
        redis_conn: aioredis.Redis
):
    schedule = await redis_conn.get(
        get_epoch_submission_schedule_key(
            project_id=project_id,
            epoch_end=epoch_end
        )
    )
    if not schedule:
        return None
    else:
        return SubmissionSchedule.parse_raw(schedule)


async def set_submission_schedule(
        project_id,
        epoch_end,
        redis_conn: aioredis.Redis
):
    cur_ts = int(time.time())
    await redis_conn.set(
        name=get_epoch_submission_schedule_key(
            project_id=project_id,
            epoch_end=epoch_end
        ),
        value=SubmissionSchedule(begin=cur_ts, end=cur_ts+settings.consensus_service.submission_window).json(),
        ex=3600
    )


async def set_submission_accepted_peers(
        project_id,
        epoch_end,
        redis_conn: aioredis.Redis
):
    await redis_conn.copy(
        get_project_registered_peers_set_key(project_id),
        get_project_epoch_specific_accepted_peers_key(project_id, epoch_end)
    )
    await redis_conn.expire(
        get_project_epoch_specific_accepted_peers_key(project_id, epoch_end),
        3600
    )


async def submission_delayed(project_id, epoch_end, auto_init_schedule, redis_conn: aioredis.Redis):
    schedule = await get_submission_schedule(project_id, epoch_end, redis_conn)
    if not schedule:
        if auto_init_schedule:
            await set_submission_accepted_peers(project_id, epoch_end, redis_conn)
            await set_submission_schedule(project_id, epoch_end, redis_conn)
        return False
    else:
        return int(time.time()) > schedule.end

async def check_consensus(
        project_id: str,
        epoch_end: int,
        redis_conn: aioredis.Redis
)-> Tuple[SubmissionAcceptanceStatus, Union[str, None]]:
    all_submissions = await redis_conn.hgetall(
        name=get_epoch_submissions_htable_key(
            project_id=project_id,
            epoch_end=epoch_end,
        )
    )
    cid_submission_map = dict()
    for instance_id_b, submission_b in all_submissions.items():
        sub_entry: SubmissionDataStoreEntry = SubmissionDataStoreEntry.parse_raw(submission_b)
        instance_id = instance_id_b.decode('utf-8')
        if sub_entry.snapshotCID not in cid_submission_map:
            cid_submission_map[sub_entry.snapshotCID] = [instance_id]
        else:
            cid_submission_map[sub_entry.snapshotCID].append(instance_id)

    num_expected_peers = await redis_conn.scard(get_project_epoch_specific_accepted_peers_key(project_id, epoch_end))
    # best case scenario
    if len(cid_submission_map.keys()) == 1:
        if len(list(cid_submission_map.values())[0]) / num_expected_peers >= 2/3:
            return SubmissionAcceptanceStatus.finalized, list(cid_submission_map.keys())[0]
        else:
            return SubmissionAcceptanceStatus.accepted, None
    else:
        sub_count_map = {k: len(cid_submission_map[k]) for k in cid_submission_map.keys()}
        num_submissions = sum(sub_count_map.values())
        for cid, sub_count in sub_count_map.items():
            # found one CID on which consensus has been reached
            if sub_count/num_submissions >= 2/3:
                return SubmissionAcceptanceStatus.finalized, cid
        else:
            # find if all peers have submitted and yet no consensus reached
            if num_submissions == num_expected_peers:
                return SubmissionAcceptanceStatus.indeterminate, None
            else:
                return SubmissionAcceptanceStatus.accepted, None


    
async def check_submissions_consensus(
        submission: Union[SnapshotSubmission, SnapshotBase],
        redis_conn: aioredis.Redis,
        epoch_consensus_check=False
) -> Tuple[SubmissionAcceptanceStatus, Union[str, None]]:
    # get all submissions
    all_submissions = await redis_conn.hgetall(
        name=get_epoch_submissions_htable_key(
            project_id=submission.projectID,
            epoch_end=submission.epoch.end,
        )
    )
    # map snapshot CID to instance ID list
    if not epoch_consensus_check and \
            submission.instanceID not in map(lambda x: x.decode('utf-8'), all_submissions.keys()):
        return SubmissionAcceptanceStatus.notsubmitted, None
    
    return await check_consensus(submission.projectID, submission.epoch.end, redis_conn)


async def register_submission(
        submission: SnapshotSubmission,
        cur_ts: int,
        redis_conn: aioredis.Redis
):
    await redis_conn.hset(
        name=get_epoch_submissions_htable_key(
            project_id=submission.projectID,
            epoch_end=submission.epoch.end,
        ),
        key=submission.instanceID,
        value=SubmissionDataStoreEntry(snapshotCID=submission.snapshotCID, submittedTS=cur_ts).json()
    )
    if await redis_conn.ttl(name=get_epoch_submissions_htable_key(
            project_id=submission.projectID,
            epoch_end=submission.epoch.end,
    )) == -1:
        await redis_conn.expire(
            name=get_epoch_submissions_htable_key(
                project_id=submission.projectID,
                epoch_end=submission.epoch.end,
            ),
            time=3600
        )
    return await check_submissions_consensus(submission, redis_conn)
