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
from skydb.skydb_utils import _value_in
import time

deal_watcher_logger = logging.getLogger(__name__)
formatter = logging.Formatter(u"%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s")

stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setLevel(logging.DEBUG)
# stdout_handler.setFormatter(formatter)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setLevel(logging.ERROR)

deal_watcher_logger.addHandler(stdout_handler)
deal_watcher_logger.addHandler(stderr_handler)
coloredlogs.install(level='DEBUG', logger=deal_watcher_logger, stream=sys.stdout)

# Tables that are used to store data in the main.py file
accounting_records_table = SkydbTable(
			table_name='accounting_records',
			columns=['c1'],
			column_split=['token','cid','localCID','txHash','confirmed','timestamp'],
			seed='abc'
		)

retreivals_single_table = SkydbTable(
			table_name='retreivals_single',
			columns=['c1'],
			column_split=['requestID','cid','localCID','retreived_file','completed'],
			seed='abc'
		)

retreivals_bulk_table = SkydbTable(
			table_name='retreivals_bulk',
			columns=['c1'],
			column_split=['requestID','api_key','token','retreived_file','completed'],
			seed='abc',
		)



with open('settings.json') as f:
	settings = json.load(f)

deals = dict()
deals_lock = threading.Lock()
redis_lock = redis.Redis()


def main():
	r = redis.StrictRedis(
		host=settings['REDIS']['HOST'],
		port=settings['REDIS']['PORT'],
		db=settings['REDIS']['DB'],
		password=settings['REDIS']['PASSWORD']
	)
	p = r.pubsub()
	p.subscribe('new_deals')
	while True:
		deal_watcher_logger.debug("Main Looping....")
		update = p.get_message(ignore_subscribe_messages=True)
		if update:
			deal_watcher_logger.debug('Got new deal update')
			deal_watcher_logger.debug(update)
			deal_to_be_watched = json.loads(update['data'])
			deals_lock.acquire()
			deals[deal_to_be_watched['jid']] = deal_to_be_watched
			deals_lock.release()
			deal_watcher_logger.debug('Current Deals to be watched set')
			deal_watcher_logger.debug(deals)
		time.sleep(5)


def job_checker():

	pow_client = PowerGateClient(fast_settings.config.powergate_url, False)
	while True:
		deal_watcher_logger.debug("Job checker Looping....")
		done_deal_jids = list()
		deals_lock.acquire()
		for deal_jid, deal in deals.items():
			deal_watcher_logger.debug(f"Getting data for deal {deal['cid']} from SkyDB")
			# Get the status of the storage job from powergate
			j_stat = pow_client.ffs.get_storage_job(jid=deal['jid'], token=deal['token'])
			# print(j_stat.job)
			if j_stat.job.status == 5:  # 'JOB_STATUS_SUCCESS':
				deal_watcher_logger.debug('Hurrah. Removing from deals to be watched...')
				deal_watcher_logger.debug(j_stat.job)

				done_deal_jids.append(deal_jid)
				# update status

				required_row = None
				while True:
					while True:
						deal_watcher_logger.debug('Waiting for Lock')
						v = redis_lock.incr('my_lock')
						required_row = None
						if v == 1:
							accounting_records_table.calibrate_index()
							required_row = accounting_records_table.fetch(
												condition = {'c1':['cid',deal['cid']]}, 
												start_index=accounting_records_table.index-1,
												n_rows=1, 
												condition_func=_value_in,
											)
							v = redis_lock.decr('my_lock')
							print(required_row)
							break
						v = redis_lock.decr('my_lock')
						time.sleep(0.1)
					try:
						assert len(required_row.keys()) >= 1, "No rows found that match the condition"
						break
					except AssertionError:
						deal_watcher_logger.debug("Refetching Data")
				update_row_index = list(required_row.keys())[0]

				# Update status on SkyDB and also log it to console
				update_data = required_row[update_row_index]['c1'].split(';')
				update_data[accounting_records_table.column_split.index('confirmed')] = 1
				update_data = ';'.join(update_data)
				accounting_records_table.update_row(row_index=update_row_index, data={'c1':update_data})
				deal_watcher_logger.debug(f"Accounting records table updated.!")

			elif j_stat.job.status == 3:
				deal_watcher_logger.error('Job failed. Removing from deals to be watched...')
				deal_watcher_logger.error(j_stat.job)
				done_deal_jids.append(deal_jid)
				# update status
				required_row == None
				while True:
					
					deal_watcher_logger.debug('Waiting for Lock')
					v = redis_lock.incr('my_lock')
					if v == 1:
						required_row = accounting_records_table.fetch({'cid':deal['cid']}, 
												start_index=accounting_records_table.index-1,
												n_rows=1, 
												num_workers=1)
						v = redis_lock.decr('my_lock')
						break
					v = redis_lock.decr('my_lock')
					time.sleep(0.1)
				assert len(required_row.keys()) >= 1, "No rows found that match the condition"
				update_row_index = list(required_row.keys())[0]
				# Update status on SkyDB and also log it to console
				update_data = required_row[update_row_index]['c1'].split(';')
				update_data[accounting_records_table.column_split.index('confirmed')] = 2
				update_data = ';'.join(update_data)
				accounting_records_table.update_row(row_index=update_row_index, data={'c1':update_data})
				deal_watcher_logger.debug(f"Accounting records table updated.!")
		for each_done_jid in done_deal_jids:
			del deals[each_done_jid]
		deals_lock.release()
		time.sleep(5)


if __name__ == '__main__':
	t1 = threading.Thread(target=main)
	t2 = threading.Thread(target=job_checker)
	t1.start()
	t2.start()
	t1.join()
	t2.join()
