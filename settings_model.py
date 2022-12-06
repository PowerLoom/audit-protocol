from pydantic import BaseModel, validator
from typing import Union, Optional, Dict
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


class RabbitMQQueueConfig(BaseModel):
    queue_name_prefix: str
    routing_key_prefix: str


class RabbitMQCoreConfig(BaseModel):
    exchange: str

class RabbitMQSetupConfig(BaseModel):
    core : RabbitMQCoreConfig
    queues: Dict[str,RabbitMQQueueConfig]

class RabbitMQConfig(BaseModel):
    user: str
    password: str
    host: str
    port: int
    setup: RabbitMQSetupConfig


class TableNames(BaseModel):
    api_keys: str
    accounting_records: str
    retreivals_single: str
    retreivals_bulk: str


class PruneSettings(BaseModel):
    segment_size: int = 700


class ConsensusConfigRateLimit(BaseModel):
    req_per_sec: int
    burst: int


# TODO: move to consensus service module
class ConsensusConfig(BaseModel):
    service_url: str
    rate_limit: ConsensusConfigRateLimit
    timeout_secs: int
    max_idle_conns: int
    idle_conn_timeout: int
    finalization_wait_time_secs: int

class Settings(BaseModel):
    instance_id: str
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
    webhook_listener: Union[WebhookListener, dict]
    redis: Union[RedisConfig, dict]
    redis_reader: Union[RedisConfig, dict]
    contract_addresses: Union[ContractAddresses, dict]
    calculate_diff: bool
    rpc_url: str
    consensus_config: ConsensusConfig