from pydantic import BaseModel, validator
from typing import Union, Optional, Dict
import json


class ContractAddresses(BaseModel):
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


class BurstRateLimit(BaseModel):
    req_per_sec: int
    burst: int


class DAGVerifierSettings(BaseModel):
    issue_reporter_port: int
    slack_notify_URL: str
    notify_suppress_time_secs: int
    concurrency: int
    ipfs_rate_limit: BurstRateLimit
    redis_pool_size: int
    run_interval_secs: int
    additional_projects_to_track_prefixes: list
    pruning_verification: bool


class PruneSettings(BaseModel):
    segment_size: int = 700


# TODO: move to consensus service module
class ConsensusConfig(BaseModel):
    service_url: str
    rate_limit: BurstRateLimit
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
    contract_call_backend: str
    dag_verifier: DAGVerifierSettings
    local_cache_path: str
    pruning: PruneSettings
    ipfs_timeout: int
    webhook_listener: Union[WebhookListener, dict]
    redis: Union[RedisConfig, dict]
    redis_reader: Union[RedisConfig, dict]
    contract_addresses: Union[ContractAddresses, dict]
    calculate_diff: bool
    rpc_url: str
    use_consensus: bool = False
    consensus_config: ConsensusConfig
    skip_summary_projects_anchor_proof: bool=False
    pooler_namespace: str
