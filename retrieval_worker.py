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
from skydb import SkydbTable
import time

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

# Tables that are used to store data in the main.py file
accounting_records_table = SkydbTable(
			table_name='accounting_records',
			columns=['token','cid','localCID','txHash','confirmed','timestamp'],
			seed='qwerasdfzxcv'
		)

retreivals_single_table = SkydbTable(
			table_name='retreivals_single',
			columns=['requestID','cid','localCID','retreived_file','completed'],
			seed='qwerasdfzxcv'
		)

retreivals_bulk_table = SkydbTable(
			table_name='retreivals_bulk',
			columns=['requestID','api_key','token','retreived_file','completed'],
			seed='qwerasdfzxcv',
		)


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
		retrieval_worker_logger.debug("Looping...")
		c = sqlite_cursor.execute("""
			SELECT requestID, cid, localCID, retrievedFile FROM retrievals_single WHERE completed=0   
		""")
		while True:
			retrieval_worker_logger.debug('Waiting for Lock')
			v = r.incr('my_lock')
			if v == 1:
				retreivals_single_table.calibrate_index()
				time.sleep(0.1)
				rows = retreivals_single_table.fetch(
							condition={'completed':'0'},
							start_index=retreivals_single_table.index-1,
							n_rows=retreivals_single_table.index-1
						)
				v = r.decr('my_lock')
				break
			v = r.decr('my_lock')
			time.sleep(0.1)

		#single_retrieval_requests = c.fetchall()
		single_retrieval_requests = rows.keys()
		
		for retrieval_request in single_retrieval_requests:
			retrieval_worker_logger.debug(retrieval_request)
			request_id = rows[retrieval_request]['requestID']
			ffs_cid = rows[retrieval_request]['cid']
			local_cid = rows[retrieval_request]['localCID']
			s = sqlite_cursor.execute("""
				SELECT token FROM accounting_records WHERE localCID=?
			""", (local_cid, ))
			res = s.fetchone()
			while True:
				retrieval_worker_logger.debug('Waiting for Lock')
				v = r.incr('my_lock')
				if v == 1:
					accounting_records_table.calibrate_index()
					time.sleep(0.1)
					acc_row = accounting_records_table.fetch(condition={'localCID':local_cid},
									start_index=accounting_records_table.index-1,
									n_rows=1, 
									)
					v = r.decr('my_lock')
					break
				v = r.decr('my_lock')
				time.sleep(0.1)

			assert len(acc_row) == 1, f"No row found  for localCID: {local_cid}"
			acc_row = acc_row[next(iter(acc_row.keys()))]
			token = res[0]
			token = acc_row['token']
			retrieval_worker_logger.debug("Retrieving file " + ffs_cid + " from FFS.")
			file_ = pow_client.ffs.get(ffs_cid, token)
			file_name = f'static/{request_id}'
			retrieval_worker_logger.debug('Saving to ' + file_name)
			try:
				with open(file_name, 'wb') as f_:
					for _ in file_:
						f_.write(_)
			except Exception:
				retrieval_worker_logger.debug('File has alreadby been saved')	
			sqlite_cursor.execute("""
				UPDATE retrievals_single SET retrievedFile=?, completed=1 WHERE requestID=?
			""", ('/'+file_name, request_id))
			sqlite_conn.commit()
			retreivals_single_table.update_row(row_index=retrieval_request,
					data={'retreived_file':'/'+file_name, 'completed':1}
					)
		# bulk retrievals
		c = sqlite_cursor.execute("""
					SELECT requestID, api_key, token FROM retrievals_bulk WHERE completed=0   
				""")
		while True:
			retrieval_worker_logger.debug('Waiting for Lock')
			v = r.incr('my_lock')
			if v == 1:
				retreivals_bulk_table.calibrate_index()
				time.sleep(0.1)
				bulk_rows = retreivals_bulk_table.fetch(condition={'completed':'0'},
								start_index=retreivals_bulk_table.index - 1,
								n_rows=retreivals_bulk_table.index-1)
				v = r.decr('my_lock')
				break
			v = r.decr('my_lock')
			time.sleep(0.1)

		bulk_retrieval_requests = bulk_rows.keys()
		for bulk_retrieval_request in bulk_retrieval_requests:
			retrieval_worker_logger.debug(bulk_rows[bulk_retrieval_request])
			request_id = bulk_rows[bulk_retrieval_request]['requestID']
			api_key = bulk_rows[bulk_retrieval_request]['api_key']
			ffs_token = bulk_rows[bulk_retrieval_request]['token']
			retrieval_worker_logger.debug(f"{request_id}, {api_key}, {ffs_token}")
			records_c = sqlite_cursor.execute('''
					SELECT cid, localCID, txHash, confirmed FROM accounting_records WHERE token=? 
				''', (ffs_token,))

			file_name = f'static/{request_id}'
			fp = open(file_name, 'wb')

			while True:
				retrieval_worker_logger.debug('Waiting for Lock')
				v = r.incr('my_lock')
				if v == 1:
					accounting_records_table.calibrate_index()
					time.sleep(0.1)
					records_rows = accounting_records_table.fetch(
								condition={'token':ffs_token},
								n_rows=accounting_records_table.index-1,
								start_index=accounting_records_table.index-1
							)
					v = r.decr('my_lock')

					break
				v = r.decr('my_lock')
				time.sleep(0.1)

			records_c = records_rows.keys()
			for each_record in records_c:
				ffs_cid = records_rows[each_record]['cid']
				retrieval_worker_logger.debug("Bulk mode: Retrieving CID " + ffs_cid + " from FFS.")
				file_ = pow_client.ffs.get(ffs_cid, ffs_token)
				try:
					for _ in file_:
						fp.write(_)
				except Exception as e:
					print(e)
			fp.close()
			sqlite_cursor.execute("""
							UPDATE retrievals_bulk SET retrievedFile=?, completed=1 WHERE requestID=?
						""", ('/' + file_name, request_id))
			sqlite_conn.commit()
			retreivals_bulk_table.update_row(row_index=bulk_retrieval_request, 
						data={'retreived_file':'/'+file_name, 'completed':1}
					)
			retrieval_worker_logger.debug('Bulk mode: Marking request ID ' + request_id + ' as completed')
		time.sleep(5)


if __name__ == '__main__':
	main()
