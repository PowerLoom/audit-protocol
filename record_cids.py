import requests
import time
import ipfshttpclient
from datetime import datetime
import json
import io
import settings
from skydb import SkydbTable

client = ipfshttpclient.connect()
table = SkydbTable(table_name='new_links',columns=['c1'],seed="MY SEED")
prevCid = None
if table.index > 0:
	prevCid = table.fetch_row(row_index=table.index-1)['c1']
	print(f"Found Latest Cid: {prevCid}")

def main():
	global prevCid, table, client
	while True:
		response = requests.get("https://strapi-matic.poly.market/markets?active=true&_sort=volume:desc")
		if response.status_code == 200:
			# Get the description of the id=93 page
			payload_data = response.json()
			data = ""
			for p in payload_data:
				if p['id'] == 93:
					data = p['description']
			
			if prevCid:
				dag = client.dag.get(prevCid).as_json()
			else:
				dag = settings.get_dag_dict()
			timestamp = datetime.strftime(datetime.now(),"%Y%m%d%H%M%S%f")

			fs = open(f'files/{timestamp}', 'w')
			fs.write(data)
			fs.close()

			snapshot = client.add('files/'+timestamp)
			snapshot = {k:v for k,v in snapshot.items()}

			dag['Data'] = "Some Data"
			dag['Height'] = table.index
			dag['prevCid'] = prevCid
			dag['Links'].append({
					'Name':'Link', 
					'Cid':snapshot['Hash'], 
					'Timestamp':timestamp,
					'Type': 'HOT_IPFS',

				})
			json_string = json.dumps(dag).encode('utf-8')
			data = client.dag.put(io.BytesIO(json_string))
			prevCid = data['Cid']['/']
			print(prevCid)
			table.add_row({'c1':prevCid})
			time.sleep(settings.snapshot_interval)

if __name__ == "__main__":
	main()
