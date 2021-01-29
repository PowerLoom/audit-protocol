from maticvigil.EVCore import EVCore
from config import settings
from utils.redis_conn import get_reader_redis_conn, get_writer_redis_conn
import ipfshttpclient
import asyncio
import logging
import pwd
import os

logger = logging.getLogger(__name__)
logger.setLevel(level="DEBUG")


async def test_ipfs_connection():
    try:
        client = ipfshttpclient.connect(settings.ipfs_url)
    except Exception as e:
        print("Failed ipfs connection ")
        logger.error(e,exc_info=True)
        raise e
    else:
        print("Test for ipfs connection success")


async def test_redis_connection():
    try:
        reader_redis_conn = await get_reader_redis_conn()
        writer_redis_conn = await get_writer_redis_conn()
    except Exception as e:
        print("Failed redis connection")
        logger.error(e,exc_info=True)
        raise e
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
        logger.error(e, exc_info=True)
        raise e
    else:
        print("Test for maticvigil connection success")


async def test_all():
    tasks = [test_redis_connection(), test_maticvigil_connection(), test_ipfs_connection()]
    await asyncio.gather(*tasks)

if __name__ == "__main__":
    loop = asyncio.get_event_loop()
    print("Teh output for pwd command: ")
    print(pwd.getpwuid(os.getuid()).pw_dir)
    try:
        loop.run_until_complete(test_all())
    except Exception as e:
        pass
    finally:
        loop.stop()
