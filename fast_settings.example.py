from pydantic import BaseSettings


class Config(BaseSettings):
    powergate_url: str = '192.168.99.100:5002'
    audit_contract: str = '0x5a07f5bdc9f82096948c53aee6fc19c4769ffb9d'


config = Config()

