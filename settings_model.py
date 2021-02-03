from pydantic import BaseModel, validator
from data_models import BloomFilterSettings
from typing import Union, List, Optional
import json


class WebhookListener(BaseModel):
    host: str
    port: int


class RedisConfig(BaseModel):
    host: str
    port: int
    db: int
    password: Optional[str]


class TableNames(BaseModel):
    api_keys: str
    accounting_records: str
    retreivals_single: str
    retreivals_bulk: str


class Settings(BaseModel):
    host: str
    port: str
    ipfs_url: str
    snapshot_interval: int
    metadata_cache: str
    dag_table_name: str
    seed: str
    dag_structure: dict
    audit_contract: str
    app_name: str
    powergate_client_addr: str
    max_ipfs_blocks: int
    max_pending_payload_commits: int
    block_storage: str
    payload_storage: str
    container_height: int
    payload_commit_interval: int
    pruning_service_interval: int
    retrieval_service_interval: int
    deal_watcher_service_interval: int
    api_key: str
    backup_targets: list
    unpin_mode: str
    max_pending_events: int
    max_payload_commits: int
    ipfs_timeout: int
    span_expire_timeout: int

    bloom_filter_settings: Union[BloomFilterSettings, dict]
    webhook_listener: Union[WebhookListener, dict]
    redis: Union[RedisConfig, dict]
    redis_reader: Union[RedisConfig, dict]

    @validator("bloom_filter_settings", "webhook_listener", "redis", "redis_reader")
    def convert_to_models(cls, data, values, **kwargs):

        if isinstance(data, dict):
            if kwargs['field'].name == "bloom_filter_settings":
                data = BloomFilterSettings(**data)

            elif kwargs['field'].name == "webhook_listener":
                data = WebhookListener(**data)

            elif (kwargs['field'].name == "redis") or (kwargs['field'] == "redis_reader"):
                data = RedisConfig(**data)

        return data