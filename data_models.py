from pydantic import BaseModel, validator
from typing import Union, List, Optional
import json


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
