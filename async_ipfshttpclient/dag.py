from httpx import AsyncClient
from io import BytesIO
import json

class IPFSAsyncClientError(Exception):
    pass

class DAGBlock:
    def __init__(self, json_body: str):
        self._dag_block_json = json_body

    def as_json(self):
        return json.loads(self._dag_block_json)

    def __str__(self):
        return self._dag_block_json


class DAGSection:
    def __init__(self, async_client: AsyncClient):
        self._client: AsyncClient = async_client

    async def put(self, bytes_body: BytesIO, pin=True):
        files = {'': bytes_body}
        r = await self._client.post(
            url=f'/dag/put?pin={str(pin).lower()}',
            files=files
        )
        if r.status_code != 200:
            raise IPFSAsyncClientError(f"IPFS client error: dag-put operation, response:{r}")
        try:
            return json.loads(r.text)
        except json.JSONDecodeError:
            return r.text

    async def get(self, dag_cid):
        response_body = ''
        async with self._client.stream(method='POST', url=f'/dag/get?arg={dag_cid}') as response:
            if response.status_code != 200:
                raise IPFSAsyncClientError(f"IPFS client error: dag-get operation, response:{response}")
            async for chunk in response.aiter_text():
                response_body += chunk
        return DAGBlock(response_body)
