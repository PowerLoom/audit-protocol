from pydantic import BaseModel, ValidationError, validator
from typing import Union
import json


class JobData(BaseModel):
    jobId: str
    jobStatus: str
    jobStatusDescription: str
    retries: int
    filecoinToken: str


class BloomFilterSettings(BaseModel):
    max_elements: int
    error_rate: float
    filename: str


class ContainerData(BaseModel):
    toHeight: int
    fromHeight: int
    containerCid: str
    projectId: str
    timestamp: int
    bloomFilterSettings: Union[dict, str, BloomFilterSettings]
    jobData: Union[dict, str, JobData]

    @validator('jobData', 'bloomFilterSettings')
    def validate_json_data(cls, data, values, **kwargs):

        if isinstance(data, str):
            try:
                data = json.loads(data)
            except json.JSONDecodeError as jdecerr:
                print(jdecerr)

        if isinstance(data, dict):

            if kwargs['field'].name == 'jobData':
                try:
                    data = JobData(**data)
                except ValidationError as verr:
                    print(verr)

            elif kwargs['field'].name == 'bloomFilterSettings':
                try:
                    data = BloomFilterSettings(**data)
                except ValidationError as verr:
                    print(verr)

        return data

    def convert_to_json(self):
        """
            - The jobData and bloomFilterSettings will be pydantic models and hence I am not putting any
            exception here. If they are not pydantic models, then the model is invalid
        """
        self.jobData = json.dumps(self.jobData.dict())
        self.bloomFilterSettings = json.dumps(self.bloomFilterSettings.dict())

