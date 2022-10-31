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


class SnapshotSubmission(BaseModel):
    epoch: EpochBase
    projectID: str
    instanceID: str
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
    # if all peers have submitted their snapshots and 2/3 consensus has not been reached
    # if submission deadline has passed, all peers have not submitted and 2/3 not reached
    indeterminate = 'INDETERMINATE'


class SubmissionResponse(BaseModel):
    status: SubmissionAcceptanceStatus
    delayedSubmission: bool
    finalizedSnapshotCID: Optional[str] = None


class SettingsConf(BaseModel):
    submission_window: int
