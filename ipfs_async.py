import ipfshttpclient
import asyncio
from greenletio import async_
from config import settings


class AsyncIpfsClient:
    def __init__(self, get, put, add_json, add_str, cat, pin_rm=None):
        self.get = get
        self.put = put
        self.add_json = add_json
        self.add_str = add_str
        self.cat = cat
        self.pin_rm = pin_rm

    async def async_get(self, *args, **kwargs):
        return await async_(self.get)(*args, **kwargs)

    async def async_put(self, *args, **kwargs):
        return await async_(self.put)(*args, **kwargs)

    async def async_add_json(self, *args, **kwargs):
        return await async_(self.add_json)(*args, **kwargs)

    async def async_add_str(self, string):
        return await async_(self.add_str)(string)

    async def async_cat(self, *args, **kwargs):
        return await async_(self.cat)(*args, **kwargs)

    async def pin_rm(self, *args, **kwargs):
        if self.pin_rm:
            return await async_(self.pin_rm)(*args, **kwargs)
        else:
            raise ValueError("pin.rm method was not passed while initializing this object")


client = ipfshttpclient.connect(settings.IPFS_URL)

async_ipfs_client = AsyncIpfsClient(
    get=client.dag.get,
    put=client.dag.put,
    add_json=client.add_json,
    add_str=client.add_str,
    cat=client.cat,
    pin_rm=client.pin.rm,
)

# Monkey patch the ipfs client
client.dag.get = async_ipfs_client.async_get
client.dag.put = async_ipfs_client.async_put
client.cat = async_ipfs_client.async_cat
client.add_json = async_ipfs_client.async_add_json
client.add_str = async_ipfs_client.async_add_str
client.pin.rm = async_ipfs_client.pin_rm

# out = asyncio.run(client.add_str("ASYNC add_str function"))
# print(out)