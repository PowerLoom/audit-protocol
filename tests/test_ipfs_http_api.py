from IPFS_API import ipfshttpclient
from config import settings
import io
import json


ipfs_client = ipfshttpclient.connect(addr=settings.ipfs.url, session=False)
# ipfs_client = ipfshttpclient.connect(addr='/ip4/127.0.0.1/tcp/5001', session=False)

# raw payload upload

res = ipfs_client.add_json({'status': 'dummy', 'body': 'dummy'})
print(res)
assert res == 'QmPLRkUySPgkMF6TGq51VmPYmP9pj2CTY96GsU3KznYgtK'

# add DAG block
dag_body = {'prevDagCid': 'bafyreicjck5vjf3ahkie7ftak5oopy5s3xc5e7sknphsqu2jr7lbt4bvle', 'data': None}
res = ipfs_client.dag.put(io.BytesIO(json.dumps(dag_body).encode('utf-8')))
dag_addition_result = res.as_json()
print(dag_addition_result)
assert dag_addition_result['Cid']['/'] == 'bafyreidjertihk2n3kquy4lhjvxagfzurmbriajp74frxzwgjhheqfl3se'

fetched = ipfs_client.dag.get(dag_addition_result['Cid']['/'])
fetched_dag_block = fetched.as_json()
print(fetched_dag_block)
assert fetched_dag_block == dag_body

