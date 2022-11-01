from pydantic import BaseModel
from typing import Union, List, Optional, Any, Dict
from enum import Enum
import json


class PeerRegistrationRequest(BaseModel):
    projectID: str
    instanceID: str


class EpochBase(BaseModel):
    begin: int
    end: int


class SnapshotBase(BaseModel):
    epoch: EpochBase
    projectID: str
    instanceID: str


class SnapshotSubmission(SnapshotBase):
    snapshotCID: str


class SubmissionSchedule(BaseModel):
    begin: int
    end: int


class SubmissionDataStoreEntry(BaseModel):
    snapshotCID: str
    submittedTS: int


class SubmissionAcceptanceStatus(str, Enum):
    accepted = 'ACCEPTED'
    finalized = 'FINALIZED'
    # if the peer never submitted yet comes around checking for status, trying to work around the system
    notsubmitted = 'NOTSUBMITTED'
    # if all peers have submitted their snapshots and 2/3 consensus has not been reached
    # if submission deadline has passed, all peers have not submitted and 2/3 not reached
    indeterminate = 'INDETERMINATE'


class EpochConsensusStatus(str, Enum):
    consensus_achieved = 'CONSENSUS_ACHIEVED'
    no_consensus = 'NO_CONSENSUS'


class SubmissionResponse(BaseModel):
    status: Union[SubmissionAcceptanceStatus, EpochConsensusStatus]
    delayedSubmission: bool
    finalizedSnapshotCID: Optional[str] = None


class SettingsConf(BaseModel):
    submission_window: int
