from pygate_grpc.client import PowerGateClient
from pygate_grpc.ffs import get_file_bytes, bytes_to_chunks, chunks_to_bytes
from google.protobuf.json_format import MessageToDict
from pygate_grpc.ffs import bytes_to_chunks

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
            row = api_keys_table.fetch(condition={'api_key': api_key_in_header},
                                       start_index=api_keys_table.index - 1,
                                       n_rows=1)
            v = redis_lock.decr('my_lock')
            break
        v = redis_lock.decr('my_lock')
        time.sleep(0.01)
    ffs_token = row[next(iter(row.keys()))]['token']
    return {'token': ffs_token, 'api_key': api_key_in_header}


 @app.post('/create')
 async def create_filecoin_filesystem(
		request: Request
 ):
	req_json = await request.json()
	hot_enabled = req_json.get('hotEnabled', True)
	pow_client = PowerGateClient(fast_settings.config.powergate_url, False)
	new_ffs = pow_client.ffs.create()
	rest_logger.info('Created new FFS')
	rest_logger.info(new_ffs)
	if not hot_enabled:
		default_config = pow_client.ffs.default_config(new_ffs.token)
		rest_logger.debug(default_config)
		new_storage_config = STORAGE_CONFIG
		new_storage_config['cold']['filecoin']['addr'] = default_config.default_storage_config.cold.filecoin.addr
		new_storage_config['hot']['enabled'] = False
		new_storage_config['hot']['allowUnfreeze'] = False
		pow_client.ffs.set_default_config(json.dumps(new_storage_config), new_ffs.token)
		rest_logger.debug('Set hot storage to False')
		rest_logger.debug(new_storage_config)
	# rest_logger.debug(type(default_config))
	api_key = str(uuid4())
	api_keys_table.add_row({'token':new_ffs.token,'api_key':api_key})

	# Add row to skydb

	api_keys_table.add_row({'api_key':api_key, 'token':new_ffs.token})
	rest_logger.debug("Added a row to api_keys_table")
	return {'apiKey': api_key}


@app.get('/payload/{recordCid:str}')
async def record(request: Request, response: Response, recordCid: str):
    # record_chain = contract.getTokenRecordLogs('0x'+keccak(text=tokenId).hex())
    # skydb fetching
    row = None
    while True:
        rest_logger.debug("Waiting for Lock")
        v = redis_lock.incr('my_lock')
        if v == 1:
            row = accounting_records_table.fetch(condition={'localCID': recordCid},
                                                 start_index=accounting_records_table.index - 1, n_rows=1)
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
            row = retreivals_single_table.fetch(condition={'cid': real_cid},
                                                start_index=retreivals_single_table.index - 1, n_rows=1)
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
        # request.app.sqlite_cursor.connection.commit()

        retreivals_single_table.add_row({
            'requestID': request_id,
            'cid': real_cid,
            'localCID': recordCid,
            'retreived_file': "",
            'completed': 0
        })

    return {'requestId': request_id, 'requestStatus': request_status, 'payloadStatus': payload_status}

@app.get('/requests/{requestId:str}')
async def request_status(request: Request, requestId: str):
    row = None
    while True:
        rest_logger.debug("Waiting for Lock")
        v = redis_lock.incr('my_lock')
        if v == 1:
            row = retreivals_single_table.fetch(condition={'requestID': requestId},
                                                start_index=retreivals_single_table.index - 1, n_rows=1)
            v = redis_lock.decr('my_lock')
            break
        v = redis_lock.decr('my_lock')
        time.sleep(0.01)
    row = row[next(iter(row.keys()))]

    if row:
        # return {'requestID': requestId, 'completed': bool(res[4]), "downloadFile": res[3]}
        return {'requestID': requestId, 'completed': bool(int(row['completed'])), "downloadFile": row['retreived_file']}
    else:
        c_bulk_row = None
        while True:
            rest_logger.debug("Waiting for Lock")
            v = redis_lock.incr('my_lock')
            if v == 1:
                c_bulk_row = retreivals_bulk_table.fetch(condition={'requestID': requestId},
                                                         start_index=retreivals_bulk_table.index - 1)
                v = redis_lock.decr('my_lock')
                break
            v = redis_lock.decr('my_lock')
            time.sleep(0.01)
        return {'requestID': requestId, 'completed': bool(int(c_bulk_row['completed'])),
                "downloadFile": c_bulk_row['retreived_file']}
