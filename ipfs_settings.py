snapshot_interval = 5
dag_table_name = "DAG_CID"
SEED = "THISISRANDOMSEED"

def get_dag_dict():
	dag_structure = {
			'Height':-1,
			'prevCid':'',
			'Data': {
					'Cid':"",
					'Type':"",
				},
			'TxHash':"",
			'Timestamp':"",
			}
	return dag_structure
