from typing import Optional, Union
from fastapi import Depends, FastAPI, WebSocket, HTTPException, Security, Request, Response, BackgroundTasks, Cookie, Query, WebSocketDisconnect
from fastapi import status, Header
from fastapi.security.api_key import APIKeyQuery, APIKeyHeader, APIKey
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from starlette.responses import RedirectResponse, JSONResponse
# from pygate_grpc.client import PowerGateClient
# from pygate_grpc.ffs import get_file_bytes, bytes_to_chunks, chunks_to_bytes
# from google.protobuf.json_format import MessageToDict
# from pygate_grpc.ffs import bytes_to_chunks
from eth_utils import keccak
from io import BytesIO
from maticvigil.EVCore import EVCore
from uuid import uuid4
import sqlite3
import logging
import sys
import json
import aioredis
import io
import redis
import time
from skydb import SkydbTable
import ipfshttpclient
#import settings
from config import settings
from datetime import datetime
print(settings.as_dict())
ipfs_client = ipfshttpclient.connect()
ipfs_table = SkydbTable(table_name=settings.dag_table_name,columns=['cid'],
				seed=settings.SEED)


formatter = logging.Formatter(u"%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s")

stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setLevel(logging.DEBUG)
# stdout_handler.setFormatter(formatter)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setLevel(logging.ERROR)
# stderr_handler.setFormatter(formatter)
rest_logger = logging.getLogger(__name__)
rest_logger.setLevel(logging.DEBUG)
rest_logger.addHandler(stdout_handler)
rest_logger.addHandler(stderr_handler)

# setup CORS origins stuff
origins = ["*"]

redis_lock = redis.Redis()

app = FastAPI(docs_url=None, openapi_url=None, redoc_url=None)
app.add_middleware(
	CORSMiddleware,
	allow_origins=origins,
	allow_credentials=True,
	allow_methods=["*"],
	allow_headers=["*"]
)
app.mount('/static', StaticFiles(directory='static'), name='static')

evc = EVCore(verbose=True)
contract = evc.generate_contract_sdk(
	contract_address=settings.audit_contract,
	app_name='auditrecords'
)


REDIS_CONN_CONF = {
	"host": settings['REDIS']['HOST'],
	"port": settings['REDIS']['PORT'],
	"password": settings['REDIS']['PASSWORD'],
	"db": settings['REDIS']['DB']
}
#
STORAGE_CONFIG = {
  "hot": {
	"enabled": True,
	"allowUnfreeze": True,
	"ipfs": {
	  "addTimeout": 30
	}
  },
  "cold": {
	"enabled": True,
	"filecoin": {
	  "repFactor": 1,
	  "dealMinDuration": 518400,
	  "renew": {
	  },
	  "addr": "placeholderstring"
	}
  }
}


@app.on_event('startup')
async def startup_boilerplate():
	app.redis_pool: aioredis.Redis = await aioredis.create_redis_pool(
		address=(REDIS_CONN_CONF['host'], REDIS_CONN_CONF['port']),
		db=REDIS_CONN_CONF['db'],
		password=REDIS_CONN_CONF['password'],
		maxsize=5
	)
	# app.sqlite_conn = sqlite3.connect('auditprotocol_1.db')
	# app.sqlite_cursor = app.sqlite_conn.cursor()


async def load_user_from_auth(
		request: Request = None
) -> Union[dict, None]:
	api_key_in_header = request.headers['Auth-Token'] if 'Auth-Token' in request.headers else None
	if not api_key_in_header:
		return None
	rest_logger.debug(api_key_in_header)
	row = None
	while True:
		rest_logger.debug("Waiting for Lock")
		v = redis_lock.incr('my_lock')
		if v == 1:
			row = api_keys_table.fetch(condition={'api_key':api_key_in_header}, 
					start_index=api_keys_table.index-1,
					n_rows=1)
			v = redis_lock.decr('my_lock')
			break
		v = redis_lock.decr('my_lock')
		time.sleep(0.01)
	ffs_token = row[next(iter(row.keys()))]['token']
	return {'token': ffs_token, 'api_key': api_key_in_header}


#@app.post('/create')
#async def create_filecoin_filesystem(
#		request: Request
#):
#	req_json = await request.json()
#	hot_enabled = req_json.get('hotEnabled', True)
#	pow_client = PowerGateClient(fast_settings.config.powergate_url, False)
#	new_ffs = pow_client.ffs.create()
#	rest_logger.info('Created new FFS')
#	rest_logger.info(new_ffs)
#	if not hot_enabled:
#		default_config = pow_client.ffs.default_config(new_ffs.token)
#		rest_logger.debug(default_config)
#		new_storage_config = STORAGE_CONFIG
#		new_storage_config['cold']['filecoin']['addr'] = default_config.default_storage_config.cold.filecoin.addr
#		new_storage_config['hot']['enabled'] = False
#		new_storage_config['hot']['allowUnfreeze'] = False
#		pow_client.ffs.set_default_config(json.dumps(new_storage_config), new_ffs.token)
#		rest_logger.debug('Set hot storage to False')
#		rest_logger.debug(new_storage_config)
#	# rest_logger.debug(type(default_config))
#	api_key = str(uuid4())
#	api_keys_table.add_row({'token':new_ffs.token,'api_key':api_key})
#
#	# Add row to skydb
#	
#	api_keys_table.add_row({'api_key':api_key, 'token':new_ffs.token})
#	rest_logger.debug("Added a row to api_keys_table")
#	return {'apiKey': api_key}



@app.get('/payload/{recordCid:str}')
async def record(request: Request, response:Response, recordCid: str):
	# record_chain = contract.getTokenRecordLogs('0x'+keccak(text=tokenId).hex())
	# skydb fetching
	row = None
	while True:
		rest_logger.debug("Waiting for Lock")
		v = redis_lock.incr('my_lock')
		if v == 1:
			row = accounting_records_table.fetch(condition={'localCID':recordCid}, start_index=accounting_records_table.index-1, n_rows=1)
			v = redis_lock.decr('my_lock')
			break
		v = redis_lock.decr('my_lock')
		time.sleep(0.01)
	assert len(row) >= 1, "No row found"
	index = list(row.keys())[0]
	row = row[index]
	confirmed = int(row['confirmed'])
	real_cid = row['cid']
	ffs_token = row['token']
	# pow_client = PowerGateClient(fast_settings.config.powergate_url, False)
	# check = pow_client.ffs.info(real_cid, ffs_token)
	# rest_logger.debug(check)

	row = None
	while True:
		rest_logger.debug("Waiting for Lock")
		v = redis_lock.incr('my_lock')
		if v == 1:
			row = retreivals_single_table.fetch(condition={'cid':real_cid}, start_index=retreivals_single_table.index-1, n_rows=1)
			v = redis_lock.decr('my_lock')
			break
		v = redis_lock.decr('my_lock')
		time.sleep(0.01)
	if row:
		row = row[next(iter(row.keys()))]
		request_id = row['requestID']
		request_status = int(row['completed'])
		if request_status == 0:
			request_status = 'InProcess'
		else:
			request_status = 'Completed'
	else:
		request_id = str(uuid4())
		request_status = 'Queued'
	# if real_cid in check.info.pins:
	#	 rest_logger.info('CID Found in Pinned!')
	#
	#	 rest_logger.debug("Retrieving file " + real_cid + " from FFS.")
	#	 file_ = pow_client.ffs.get(real_cid, ffs_token)
	#	 file_name = f'static/{request_id}'
	#	 rest_logger.debug('Saving to ' + file_name)
	#	 with open(file_name, 'wb') as f_:
	#		 for _ in file_:
	#			 f_.write(_)
	#	 return {'requestId': request_id, 'downloadFile': file_name}
	if confirmed == 0:
		# response.status_code = status.HTTP_404_NOT_FOUND
		payload_status = 'PendingPinning'
	elif confirmed == 1:
		payload_status = 'Pinned'
	elif confirmed == 2:
		payload_status = 'PinFailed'
	else:
		payload_status = 'unknown'
	if confirmed in range(0, 2):
		request.app.sqlite_cursor.execute("""
			INSERT INTO retrievals_single VALUES (?, ?, ?, "", 0)
		""", (request_id, real_cid, recordCid))
		#request.app.sqlite_cursor.connection.commit()

		retreivals_single_table.add_row({
					'requestID':request_id,
					'cid':real_cid,
					'localCID':recordCid,
					'retreived_file':"",
					'completed':0
				})

	return {'requestId': request_id, 'requestStatus': request_status, 'payloadStatus': payload_status}


@app.get('/requests/{requestId:str}')
async def request_status(request: Request, requestId: str):

	row = None
	while True:
		rest_logger.debug("Waiting for Lock")
		v = redis_lock.incr('my_lock')
		if v == 1:
			row = retreivals_single_table.fetch(condition={'requestID':requestId}, start_index=retreivals_single_table.index-1, n_rows=1)
			v = redis_lock.decr('my_lock')
			break
		v = redis_lock.decr('my_lock')
		time.sleep(0.01)
	row = row[next(iter(row.keys()))]

	if row:
		#return {'requestID': requestId, 'completed': bool(res[4]), "downloadFile": res[3]}
		return {'requestID':requestId, 'completed': bool(int(row['completed'])), "downloadFile":row['retreived_file']}
	else:
		c_bulk_row = None
		while True:
			rest_logger.debug("Waiting for Lock")
			v = redis_lock.incr('my_lock')
			if v == 1:
				c_bulk_row = retreivals_bulk_table.fetch(condition={'requestID':requestId}, start_index=retreivals_bulk_table.index-1)
				v = redis_lock.decr('my_lock')
				break
			v = redis_lock.decr('my_lock')
			time.sleep(0.01)
		return {'requestID': requestId, 'completed': bool(int(c_bulk_row['completed'])), 
				"downloadFile": c_bulk_row['retreived_file']}



def get_prev_cid():
	global ipfs_table
	if ipfs_table.index == 0:
		return ""
	prev_cid = ipfs_table.fetch_row(row_index=ipfs_table.index-1)['cid']
	return prev_cid

@app.post('/commit_payload')
async def commit_payload(
		request: Request,
		response: Response,
):
	req_args = await request.json()
	try:
		payload = req_args['payload']
		projectId = req_args['projectId']
	except Exception as e:
		return {'error': "Either payload or projectId"}
		 
	
	ipfs_table = SkydbTable(
				table_name=f"{settings.dag_table_name}:{projectId}",
				columns = ['cid'],
				seed = settings.seed,
				verbose=1
			)

	payload_changed = False
	prev_payload_cid = None

	if ipfs_table.index == 0:
		prev_cid = ""
	else:
		prev_cid = ipfs_table.fetch_row(row_index=ipfs_table.index-1)['cid']
		prev_payload_cid = ipfs_client.dag.get(prev_cid).as_json()['Data']['Cid']
	rest_logger.debug('Previous IPLD CID in the DAG: ')
	rest_logger.debug(prev_cid)

	dag = settings.dag_structure.to_dict()
	timestamp = datetime.strftime(datetime.now(), "%Y%m%d%H%M%S%f")

	if type(payload) is dict:
		snapshot_cid = ipfs_client.add_json(payload)
	else:
		try:
			snapshot_cid = ipfs_client.add_str(str(payload))
		except:
			response.status_code = 400
			return {'success': False, 'error': 'PayloadNotSuppported'}
	rest_logger.debug('Payload CID')
	rest_logger.debug(snapshot_cid)
	snapshot = dict()
	snapshot['Cid'] = snapshot_cid
	snapshot['Type'] = "HOT_IPFS"
	if prev_payload_cid:
		if prev_payload_cid != snapshot['Cid']:
			payload_changed = True
	rest_logger.debug(snapshot)
	dag['Height'] = ipfs_table.index
	dag['prevCid'] = prev_cid
	dag['Data'] = snapshot
	ipfs_cid = snapshot['Cid']
	token_hash = '0x' + keccak(text=json.dumps(snapshot)).hex()
	tx_hash_obj = contract.commitRecord(**dict(
		ipfsCid=ipfs_cid,
		apiKeyHash=token_hash,
	))
	dag['TxHash'] = tx_hash_obj[0]['txHash']
	timestamp = datetime.strftime(datetime.now(), "%Y%m%d%H%M%S%f")
	dag['Timestamp'] = timestamp
	rest_logger.debug(dag)
	json_string = json.dumps(dag).encode('utf-8')
	data = ipfs_client.dag.put(io.BytesIO(json_string))
	rest_logger.debug(data)
	rest_logger.debug(data['Cid']['/'])
	ipfs_table.add_row({'cid': data['Cid']['/']})
	return {
		'Cid': data['Cid']['/'],
		# 'payloadCid': snapshot['Cid'],
		'payloadChanged': payload_changed,
		'Height': dag['Height']
	}


def get_block_height():
	return ipfs_table.index-1


@app.get('/{projectId:int}/payloads')
async def get_payloads(
	request: Request,
	response: Response,
	projectId:int,
	from_height:int = Query(None),
	to_height:int = Query(None),
	data:Optional[str]=Query(None)
):
	ipfs_table = SkydbTable(table_name=f"{settings.dag_table_name}:{projectId}",
			columns=['cid'],
			seed=settings.seed,
			verbose=1)
	if data:
		if data.lower() == 'true':
			data = True
		else:
			data = False

	if (from_height < 0) or (to_height > ipfs_table.index-1) or (from_height > to_height):
		return {'error': 'Invalid Height'}

	blocks = {}
	current_height = to_height
	prevCid = ""
	while current_height >= from_height:
		rest_logger.debug("Fetching block at height: "+str(current_height))
		if not prevCid:
			prevCid = ipfs_table.fetch_row(row_index=current_height)['cid']
		block = ipfs_client.dag.get(prevCid).as_json()
		if data:
			block['Data']['payload'] = ipfs_client.cat(block['Data']['Cid']).decode()
		blocks.update({prevCid:block})
		prevCid = block['prevCid']
		current_height = current_height - 1

	return blocks



@app.get('/{projectId}/payload/height')
async def payload_height(request: Request, response:Response, projectId:int):
	ipfs_table = SkydbTable(
				table_name=f"{settings.dag_table_name}:{projectId}",
				columns=['cid'],
				seed=settings.seed
			)
	height = ipfs_table.index - 1
	return {"height": height}

@app.get('/{projectId}/payload/{block_height}')
async def get_block(request:Request,
		response:Response,
		projectId:int,
		block_height:int,
):
	ipfs_table = SkydbTable(table_name=f"{settings.dag_table_name}:{projectId}",
			columns=['cid'],
			seed=settings.seed,
			verbose=1)

	if (block_height > ipfs_table.index-1) or (block_height < 0):
		return {'error':'Invalid block Height'}


	row = ipfs_table.fetch_row(row_index=block_height)
	block = ipfs_client.dag.get(row['cid']).as_json()
	return {row['cid']:block}


@app.get('/{projectId:int}/payload/{block_height:int}/data')
async def get_block_data(
	request:Request,
	response:Response,
	projectId:int,
	block_height:int,
):
	ipfs_table = SkydbTable(table_name=f"{settings.dag_table_name}:{projectId}",
			columns=['cid'],
			seed=settings.seed,
			verbose=1)
	if (block_height > ipfs_table.index - 1) or (block_height < 0):
		return {'error':'Invalid block Height'}
	row = ipfs_table.fetch_row(row_index=block_height)
	block = ipfs_client.dag.get(row['cid']).as_json()
	block['Data']['payload'] = ipfs_client.cat(block['Data']['Cid']).decode()
	return {row['cid']:block['Data']}
