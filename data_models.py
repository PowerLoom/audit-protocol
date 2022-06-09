from pydantic import BaseModel, validator
from typing import Union, List, Optional, Any, Dict
import json


class AuditRecordTxEventData(BaseModel):
    txHash: str
    projectId: str
    apiKeyHash: str
    timestamp: float
    payloadCommitId: str
    snapshotCid: str
    tentativeBlockHeight: int


class PendingTransaction(BaseModel):
    txHash: str
    lastTouchedBlock: int = 0
    event_data: Optional[AuditRecordTxEventData] = dict()


class PayloadCommit(BaseModel):
    projectId: str
    commitId: str
    payload: Optional[dict] = None
    # following two can be used to substitute for not supplying the payload but the CID and hash itself
    snapshotCID: Optional[str] = None
    apiKeyHash: Optional[str] = None
    tentativeBlockHeight: int
    resubmitted: bool = False
    resubmissionBlock: int = 0  # corresponds to lastTouchedBlock in PendingTransaction model


class FilecoinJobData(BaseModel):
    stagedCid: str = ""
    jobId: str = ""
    jobStatus: str = ""
    jobStatusDescription: str = ""
    retries: int = 0
    filecoinToken: str = ""


class BloomFilterSettings(BaseModel):
    max_elements: int = 0
    error_rate: float = 0.0
    filename: Optional[str] = ""


class SiaData(BaseModel):
    fileHash: str = ""
    skylink: str = ""


class SiaSkynetData(BaseModel):
    skylink: str = ""


class SiaRenterData(BaseModel):
    fileHash: str = ""


class BackupMetaData(BaseModel):
    sia_skynet: Optional[SiaSkynetData] = SiaSkynetData()  # Create empty placeholders
    sia_renter: Optional[SiaRenterData] = SiaRenterData()  # Create empty placeholders
    filecoin: Optional[FilecoinJobData] = FilecoinJobData()  # Create empty placeholders

    @validator("sia_skynet", "sia_renter", "filecoin")
    def validate_json_data(cls, data, values, **kwargs):
        if isinstance(data, str):
            try:
                data = json.loads(data)
            except json.JSONDecodeError as jdecerr:
                print(jdecerr)

        if isinstance(data, dict):
            if kwargs['field'].name == "sia_skynet":
                data = SiaSkynetData(**data)
            elif kwargs['field'].name == "sia_renter":
                data = SiaRenterData(**data)
            elif kwargs['field'].name == "filecoin":
                data = FilecoinJobData(**data)
        return data


class ContainerData(BaseModel):
    toHeight: int
    fromHeight: int
    projectId: str
    timestamp: int
    backupTargets: Union[str, List[str]]
    backupMetaData: Union[dict, str, BackupMetaData]
    bloomFilterSettings: Union[dict, str, BloomFilterSettings]

    @validator('backupMetaData', 'bloomFilterSettings', 'backupTargets')
    def validate_json_data(cls, data, values, **kwargs):

        if isinstance(data, str):
            try:
                data = json.loads(data)
            except json.JSONDecodeError as jdecerr:
                print(jdecerr)

        if isinstance(data, dict):

            if kwargs['field'].name == 'backupMetaData':
                data = BackupMetaData(**data)

            elif kwargs['field'].name == 'bloomFilterSettings':
                data = BloomFilterSettings(**data)

        return data

    def convert_to_json(self):
        self.backupTargets = json.dumps(self.backupTargets)
        self.backupMetaData = json.dumps(self.backupMetaData.dict())
        self.bloomFilterSettings = json.dumps(self.bloomFilterSettings.dict())


class liquidityProcessedData(BaseModel):
    contractAddress: str
    name: str
    liquidity: str
    volume_24h: str
    volume_7d: str
    cid_volume_24h: str
    cid_volume_7d: str
    fees_24h: str
    block_height: int
    block_timestamp: int
    token0Liquidity: float
    token1Liquidity: float
    token0LiquidityUSD: float
    token1LiquidityUSD: float
    token0TradeVolume_24h: float
    token1TradeVolume_24h: float
    token0TradeVolumeUSD_24h: float
    token1TradeVolumeUSD_24h: float
    token0TradeVolume_7d: float
    token1TradeVolume_7d: float
    token0TradeVolumeUSD_7d: float
    token1TradeVolumeUSD_7d: float


class DAGBlockPayloadLinkedPath(BaseModel):
    cid: Dict[str, str]
    type: str


class DAGBlock(BaseModel):
    height: int
    prevCid: Optional[Dict[str, str]]
    data: DAGBlockPayloadLinkedPath
    txHash: str
    timestamp: int


class DiffCalculationRequest(BaseModel):
    project_id: str
    dagCid: str
    lastDagCid: Optional[str] = None
    payloadCid: str
    txHash: str
    tentative_block_height: int
    timestamp: int
