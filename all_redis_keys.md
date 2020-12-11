### All the Redis Keys that are being used across all the services

##### filecoinToken:{projectId} - STRING
- This redis key is used to store a filecoin user token against each of the unique projectIds
- As soon as the /commit_payload endpoint is called, the function get_project_token will try to retrieve the filecoin token
for that projectId if it exists else it will create a new token for that projectId

##### storedProjectIds - SET
- This redis key holds the set of all unique projectIds which exist in our database
- As soon as a filecoin token is create for a projectId, that projectId will be put into storedProjectIds set.

##### projectID:{projectId}:tentativeBlockHeight - STRING
- Keep the latest tentative block height.

##### projectID:{projectId}:lastDagCid - STRING
- Holds the cid of latest DAG block

##### projectID:{projectId}:blockHeight - STRING
- Holds the height of the latest DAG Block

##### projectID:{projectId}:payloadCids - ZSET
- A Zset which will hold the list of cids of the payloads that are pushed to filecoin and the scores will be their 
tentative_block_height

##### projectID:{project_id}:Cids - ZSET
- A Zset which will hold the list of cids of DAG blocks with block height as the score

##### jobStatus:{snapshot_cid} - STRING
- This key:value pair will hold the jobId for the snapshot_cid

##### TRANSACTION:{txHash} - HASH-SET
- This hash-set holds 3 fields: 
	- project_id: The project_id for this transaction
	- tentative_block_height: The tentative block height for this snapshot
	- prev_dag_cid: The cid of the latest DAG block that has been committed

- This hash-set is create as soon as a snapshot has been committed to smart contract. This hash-set will be used in webhook 
listener for creating the block DAG

##### pendingRetrievalRequests - SET
- This set will hold all the pending requestIds.
- Once a request has been completed, that requestId will be removed from this set.

##### retrievalRequestFiles:{requestId} - ZSET
- A Zset which will paths to requested files that have been retrieved by the retrieval service, with scores being block 
height of each of the block

##### retrievalRequestInfo:{requestId} - HASH-SET
- This hash-set holds 4 fields:
	- project_id: The project_id that is associated with the request
	- from_height: The height of the block from where we need to start fetching the blocks
	- to_height: The height till where we need to fecth the blocks
	- data: A flag which can be
		- '0': Retrieve only the DAG block with no payload data.
		- '1': Retrieve the DAG Block along with the payload data.
		- '2': Retrieve only the payload data and not the DAG block

##### CidDiff:{prev_payload_cid}:{current_payload_cid} - SET
- This set holds the entire diff_map between the current payload and its previous snapshot

##### blockFilecoinStorage:{projectId}:{tentative_block_height} - HASH-SET
- This Hash set is responsible holding the information of block DAG once data has been pushed to filecoin
- The fields are:
	- blockStageCid: This is cid of you get after staging the block dag onto ipfs
	- blockDagCid: The Cid of the block's DAG
	- jobId: The jobId of the storage deal of the DAG block once it has been pushed to filecoin

##### projectID:{projectId}:diffSnapshots - ZSET
- This Zset will hold the diff_map between two consectutive snapshots, with the score being the block height of the target snapshot
