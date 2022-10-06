from httpx import AsyncClient, Timeout, Limits
from starlette.background import BackgroundTask
from starlette.responses import StreamingResponse
from dag import DAGSection
import utils
import json


class AsyncIPFSClient:
    def __init__(
            self,
            addr,
            api_base='api/v0'

    ):
        self._base_url, self._host_numeric = utils.addr.multiaddr_to_url_data(addr, api_base)
        print(self._base_url, self._host_numeric)
        self.dag = None

    async def init_session(self):
        self._client = AsyncClient(
            base_url=self._base_url,
            timeout=Timeout(timeout=5.0),
            follow_redirects=False,
            limits=Limits(max_connections=100, max_keepalive_connections=20, keepalive_expiry=5.0)
        )
        self.dag = DAGSection(self._client)

    async def add_bytes(self, data: bytes, **kwargs):
        # TODO
        pass

    def add_str(self, string, **kwargs):
        # TODO
        pass

    def add_json(self, json_obj, **kwargs):
        # TODO
        pass

    async def cat(self, cid, **kwargs):
        response_body = ''
        async with self._client.stream(method='POST', url=f'/cat?arg={cid}') as response:
            async for chunk in response.aiter_text():
                response_body += chunk
        return response_body

    async def get_json(self, cid, **kwargs):
        return json.loads(await self.cat(cid))
