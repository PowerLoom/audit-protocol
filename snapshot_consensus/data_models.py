from pydantic import BaseModel
from typing import Union, List, Optional, Any, Dict
from enum import Enum
from settings_model import RedisConfig
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


class SubmissionStatus(str, Enum):
    within_schedule = 'WITHIN_SCHEDULE'
    delayed = 'DELAYED'


class EpochConsensusStatus(str, Enum):
    consensus_achieved = 'CONSENSUS_ACHIEVED'
    no_consensus = 'NO_CONSENSUS'


class SubmissionResponse(BaseModel):
    status: Union[SubmissionAcceptanceStatus, EpochConsensusStatus]
    delayedSubmission: bool
    finalizedSnapshotCID: Optional[str] = None


class ConsensusService(BaseModel):
    submission_window: int
    host: str
    port: str


class SettingsConf(BaseModel):
    consensus_service: ConsensusService
    redis: RedisConfig


# Data model for a list of snapshotters
class Snapshotters(BaseModel):
    projectId: str
    snapshotters: List[str]


class Epoch(BaseModel):
    sourcechainEndheight: int
    finalized: bool


# Data model for a list of epoch data
class EpochData(BaseModel):
    projectId: str
    epochs: List[Epoch]


# Data model for a submission
class Submission(BaseModel):
    snapshotterInstanceID: str
    snapshotCID: str
    submittedTS: int
    submissionStatus: SubmissionStatus


class Message(BaseModel):
    message: str
