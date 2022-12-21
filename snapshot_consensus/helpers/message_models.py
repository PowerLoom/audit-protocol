from pydantic import BaseModel
from typing import List


class RPCNodesObject(BaseModel):
    NODES: List[str]
    RETRY_LIMIT: int
