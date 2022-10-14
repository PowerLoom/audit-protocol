from pydantic import BaseModel, validator
from data_models import BloomFilterSettings
from typing import Union, List, Optional
import json


class ContractAddresses(BaseModel):
    iuniswap_v2_factory: str
    iuniswap_v2_router: str
    iuniswap_v2_pair: str
    USDT: str
    DAI: str
    USDC: str
    WETH: str
    MAKER: str


class WebhookListener(BaseModel):
    host: str
    port: int
    validate_header_sig: bool = False
    keepalive_secs: int = 600
    redis_lock_lifetime: int


class HTTPClientConnection(BaseModel):
    sock_read: int
    sock_connect: int
    connect: int


class RedisConfig(BaseModel):
    host: str
    port: int
    db: int
    password: Optional[str]


class RabbitMQConfig(BaseModel):
    user: str
    password: str
    host: str
    port: int
    setup: dict


class TableNames(BaseModel):
    api_keys: str
    accounting_records: str
    retreivals_single: str
    retreivals_bulk: str


class PruneSettings(BaseModel):
    segment_size: int = 700

class Settings(BaseModel):
    host: str
    port: str
    keepalive_secs: int = 600
    rlimit: dict
    ipfs_url: str
    ipfs_reader_url: str
    rabbitmq: RabbitMQConfig
    snapshot_interval: int
    seed: str
    audit_contract: str
    contract_call_backend: str
    powergate_client_addr: str
    max_ipfs_blocks: int
    max_pending_payload_commits: int
    container_height: int
    payload_commit_interval: int
    pruning_service_interval: int  # TODO: this field will be deprecated and references should be removed
    pruning: PruneSettings
    retrieval_service_interval: int
    deal_watcher_service_interval: int
    api_key: str
    backup_targets: list
    unpin_mode: str
    max_pending_events: int
    max_payload_commits: int
    ipfs_timeout: int
    span_expire_timeout: int
    aiohtttp_timeouts: Union[HTTPClientConnection, dict]
    bloom_filter_settings: Union[BloomFilterSettings, dict]
    webhook_listener: Union[WebhookListener, dict]
    redis: Union[RedisConfig, dict]
    redis_reader: Union[RedisConfig, dict]
    contract_addresses: Union[ContractAddresses, dict]
    calculate_diff: bool
    rpc_url: str

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
