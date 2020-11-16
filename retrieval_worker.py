from pygate_grpc.client import PowerGateClient
from pygate_grpc.ffs import get_file_bytes, bytes_to_chunks, chunks_to_bytes
from google.protobuf.json_format import MessageToDict
from pygate_grpc.ffs import bytes_to_chunks
import redis
import json
import time
import threading
import fast_settings
import queue
import sqlite3
import coloredlogs
import logging
import sys

retrieval_worker_logger = logging.getLogger(__name__)
formatter = logging.Formatter(u"%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s")

stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setLevel(logging.DEBUG)
# stdout_handler.setFormatter(formatter)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setLevel(logging.ERROR)

retrieval_worker_logger.addHandler(stdout_handler)
retrieval_worker_logger.addHandler(stderr_handler)
coloredlogs.install(level='DEBUG', logger=retrieval_worker_logger, stream=sys.stdout)

with open('settings.json') as f:
    settings = json.load(f)


def main():
    sqlite_conn = sqlite3.connect('auditprotocol_1.db')
    sqlite_cursor = sqlite_conn.cursor()
    r = redis.StrictRedis(
        host=settings['REDIS']['HOST'],
        port=settings['REDIS']['PORT'],
        db=settings['REDIS']['DB'],
        password=settings['REDIS']['PASSWORD']
    )
    pow_client = PowerGateClient(fast_settings.config.powergate_url, False)
    awaited_deals = list()
    while True:
        c = sqlite_cursor.execute("""
            SELECT requestID, cid, localCID, retrievedFile FROM retrievals_single WHERE completed=0   
        """)
        single_retrieval_requests = c.fetchall()
        for retrieval_request in single_retrieval_requests:
            retrieval_worker_logger.debug(retrieval_request)
            request_id = retrieval_request[0]
            ffs_cid = retrieval_request[1]
            local_cid = retrieval_request[2]
            s = sqlite_cursor.execute("""
                SELECT token FROM accounting_records WHERE localCID=?
            """, (local_cid, ))
            res = s.fetchone()
            token = res[0]
            retrieval_worker_logger.debug("Retrieving file " + ffs_cid + " from FFS.")
            file_ = pow_client.ffs.get(ffs_cid, token)
            file_name = f'static/{request_id}'
            retrieval_worker_logger.debug('Saving to ' + file_name)
            with open(file_name, 'wb') as f_:
                for _ in file_:
                    f_.write(_)
            sqlite_cursor.execute("""
                UPDATE retrievals_single SET retrievedFile=?, completed=1 WHERE requestID=?
            """, ('/'+file_name, request_id))
            sqlite_conn.commit()
        # bulk retrievals
        c = sqlite_cursor.execute("""
                    SELECT requestID, api_key, token FROM retrievals_bulk WHERE completed=0   
                """)
        bulk_retrieval_requests = c.fetchall()
        for bulk_retrieval_request in bulk_retrieval_requests:
            retrieval_worker_logger.debug(bulk_retrieval_request)
            request_id = bulk_retrieval_request[0]
            api_key = bulk_retrieval_request[1]
            ffs_token = bulk_retrieval_request[2]
            records_c = sqlite_cursor.execute('''
                    SELECT cid, localCID, txHash, confirmed FROM accounting_records WHERE token=? 
                ''', (ffs_token,))
            file_name = f'static/{request_id}'
            fp = open(file_name, 'wb')
            for each_record in records_c:
                ffs_cid = each_record[0]
                retrieval_worker_logger.debug("Bulk mode: Retrieving CID " + ffs_cid + " from FFS.")
                file_ = pow_client.ffs.get(ffs_cid, ffs_token)
                for _ in file_:
                    fp.write(_)
            fp.close()
            sqlite_cursor.execute("""
                            UPDATE retrievals_bulk SET retrievedFile=?, completed=1 WHERE requestID=?
                        """, ('/' + file_name, request_id))
            sqlite_conn.commit()
            retrieval_worker_logger.debug('Bulk mode: Marking request ID ' + request_id + ' as completed')
        time.sleep(5)


if __name__ == '__main__':
    main()
