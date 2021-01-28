from maticvigil.EVCore import EVCore
from config import settings
from utils.redis_conn import provide_async_reader_conn_inst, provide_async_writer_conn_inst
import ipfshttpclient
import asyncio
import time


async def test_ipfs_connection():
    try:
        client = ipfshttpclient.connect(settings.ipfs_url)
    except Exception as e:
        print("Failed ipfs connection ")
        print(e)
    else:
        print("Test for ipfs connection success")


async def test_redis_connection():
    try:
        reader_redis_conn = await provide_async_reader_conn_inst()
        writer_redis_conn = await provide_async_writer_conn_inst()
    except Exception as e:
        print("Failed redis connection")
        print(e)
    else:
        print("Test for redis connection success")


async def test_maticvigil_connection():
    evc = EVCore(verbose=True)
    try:
        contract = evc.generate_contract_sdk(
            contract_address=settings.audit_contract,
            app_name='auditrecords'
        )
    except Exception as e:
        print("Failed maticvigil connection")
        print(e)
    else:
        print("Test for maticvigil connection success")


async def test_all():
    tasks = [test_redis_connection(), test_maticvigil_connection(), test_ipfs_connection()]
    await asyncio.gather(*tasks)

if __name__ == "__main__":
    loop = asyncio.get_event_loop()
    time.sleep(7)  # Wait for ipfs daemon to run
    try:
        loop.run_until_complete(test_all())
    except Exception as e:
        pass
    finally:
        loop.stop()
