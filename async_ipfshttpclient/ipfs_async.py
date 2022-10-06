from main import AsyncIPFSClient
import asyncio
from config import settings

ipfs_write_client = AsyncIPFSClient(addr=settings.ipfs_url)
ipfs_read_client = AsyncIPFSClient(addr=settings.ipfs_reader_url)

async def init_ipfs_client():
    await ipfs_write_client.init_session()
    await ipfs_read_client.init_session()
    print(f"Initialized IPFS clients!!")


if __name__ == "__main__":
    tasks = asyncio.gather(init_ipfs_client())
    asyncio.get_event_loop().run_until_complete(tasks)