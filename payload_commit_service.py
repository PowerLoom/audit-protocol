import aioredis
from config import settings
import logging
import asyncio
import sys
import json
import io
import time
from main import get_project_token
from eth_utils import keccak
from main import make_transaction
from maticvigil.EVCore import EVCore
from redis_conn import RedisPool, provide_async_reader_conn_inst, provide_async_writer_conn_inst

""" Powergate Imports """
from pygate_grpc.client import PowerGateClient


formatter = logging.Formatter(u"%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s")

stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setLevel(logging.DEBUG)
# stdout_handler.setFormatter(formatter)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setLevel(logging.ERROR)
# stderr_handler.setFormatter(formatter)
payload_logger = logging.getLogger(__name__)
payload_logger.setLevel(logging.DEBUG)
payload_logger.addHandler(stdout_handler)
payload_logger.addHandler(stderr_handler)

payload_logger.debug("Initialized Payload")

evc = EVCore(verbose=True)
contract = evc.generate_contract_sdk(
    contract_address=settings.audit_contract,
    app_name='auditrecords'
)


@provide_async_reader_conn_inst
@provide_async_writer_conn_inst
async def commit_pending_payloads(
        reader_redis_conn=None,
        writer_redis_conn=None
):
    """
        - The goal of this function will be to check if there are any pending
        payloads left to commit, and take action on them
    """
    global contract

    payload_logger.debug("Checking for pending payloads to commit...")
    pending_payload_commits_key = f"pendingPayloadCommits"
    pending_payloads = await reader_redis_conn.lrange(pending_payload_commits_key, 0, -1)
    if len(pending_payloads) > 0:
        payload_logger.debug("Pending payloads found: ")
        payload_logger.debug(pending_payloads)

        """ Processing each of the pending payloads """
        while True:

            payload_commit_id = await writer_redis_conn.rpop(pending_payload_commits_key)
            if not payload_commit_id:
                payload_logger.debug("All payloads committed...")
                break
            payload_commit_id = payload_commit_id.decode('utf-8')
            payload_logger.debug("Processing payload: "+payload_commit_id)
            
            payload_commit_key = f"payloadCommit:{payload_commit_id}"
            out = await reader_redis_conn.hgetall(payload_commit_key)
            payload_data = {k.decode('utf-8'):v.decode('utf-8') for k,v in out.items()}
            payload_logger.debug(payload_data)

            snapshot = dict()
            snapshot['cid'] = payload_data['snapshotCid']
            snapshot['type'] = "COLD_FILECOIN"

            token_hash = '0x' + keccak(text=json.dumps(snapshot)).hex()
            _ = await make_transaction(
                                    snapshot_cid=payload_data['snapshotCid'], 
                                    token_hash=token_hash, 
                                    payload_commit_id=payload_commit_id,
                                    last_tentative_block_height=payload_data['tentativeBlockHeight'],
                                    project_id=payload_data['projectId'],
                                    writer_redis_conn=writer_redis_conn,
                                    contract=contract
                        )
            payload_logger.debug("The payload: "+payload_commit_id+" has been succesfully committed...")
                        

if __name__ == "__main__":
    payload_logger.debug("Starting the loop")
    while True:
        asyncio.run(commit_pending_payloads())
        payload_logger.debug("Sleeping for 20 seconds...")
        time.sleep(settings.payload_commit_interval)