import sys
import os
current = os.path.dirname(os.path.realpath(__file__))
parent = os.path.dirname(current)
sys.path.append(parent)

from httpx import AsyncClient, Timeout, Limits, AsyncHTTPTransport
from httpx_auth import Basic
from async_ipfshttpclient.dag import DAGSection, IPFSAsyncClientError
from config import settings
import async_ipfshttpclient.utils as utils
import json
import asyncio

class AsyncIPFSClient:
    def __init__(
            self,
            addr,
            basic_auth_username:str,
            basic_auth_passkey:str,
            api_base='api/v0'

    ):
        self._base_url, self._host_numeric = utils.addr.multiaddr_to_url_data(addr, api_base)
        self.basic_auth_username=basic_auth_username
        self.basic_auth_passkey=basic_auth_passkey
        self.dag = None
        self._client = None

    async def init_session(self):
        if not self._client:
            self._async_transport = AsyncHTTPTransport(
                limits=Limits(max_connections=300, max_keepalive_connections=100, keepalive_expiry=30)
            )
            if self.basic_auth_username is not None and self.basic_auth_passkey is not None:
                auth_config =Basic(self.basic_auth_username, self.basic_auth_passkey)
            else:
                auth_config = None

            self._client = AsyncClient(
                base_url=self._base_url,
                timeout=Timeout(timeout=5.0),
                follow_redirects=False,
                transport=self._async_transport,
                auth=auth_config
            )
            self.dag = DAGSection(self._client)

    def add_str(self, string, **kwargs):
        # TODO
        pass

    async def add_bytes(self, data: bytes, **kwargs):
        files = {'': data}
        r = await self._client.post(
            url='/add?cid-version=1',
            files=files
        )
        if r.status_code != 200:
            raise IPFSAsyncClientError(f"IPFS client error: add_bytes operation, response:{r}")

        try:
            return json.loads(r.text)
        except json.JSONDecodeError:
            return r.text


    async def add_json(self, json_obj, **kwargs):
        try:
            json_data = json.dumps(json_obj).encode('utf-8')
        except Exception as e:
            raise e

        cid = await self.add_bytes(json_data, **kwargs)
        return cid['Hash']

    async def cat(self, cid, **kwargs):
        response_body = ''
        async with self._client.stream(method='POST', url=f'/cat?arg={cid}') as response:
            if response.status_code != 200:
                raise IPFSAsyncClientError(f"IPFS client error: cat operation, response:{response}")

            async for chunk in response.aiter_text():
                response_body += chunk
        return response_body

    async def get_json(self, cid, **kwargs):
        json_data = await self.cat(cid)
        try:
            return json.loads(json_data)
        except json.JSONDecodeError:
            return json_data


class AsyncIPFSClientSingleton:
    def __init__(self):
        self._ipfs_write_client = AsyncIPFSClient(
            addr=settings.ipfs_url,
            basic_auth_username=settings.ipfs.basic_auth.username,
            basic_auth_passkey=settings.ipfs.basic_auth.passkey)
        self._ipfs_read_client = AsyncIPFSClient(
            addr=settings.ipfs_reader_url,
            basic_auth_username=settings.ipfs_basic_auth_username,
            basic_auth_passkey=settings.ipfs_basic_auth_passkey)
        self._initialized = False

    async def init_sessions(self):
        if self._initialized:
            return
        await self._ipfs_write_client.init_session()
        await self._ipfs_read_client.init_session()
        self._initialized = True
