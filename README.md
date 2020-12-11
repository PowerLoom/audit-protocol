# audit-protocol-private

```
make localnet

python3 gunicorn_main_launcher.py
python3 gunicorn_webhook_launcher.py
python3 payload_commit_service.py
python3 retrieval_service.py
python3 pruning_service.py

```

#### Options
- "max_ipfs_blocks": This number represents the latest max_ipfs_blocks need to fetched using ipfs_client
- "max_pending_payload_commits": This variable is not yet used anywhere. This variable represents the limit at which, 
you need to block further calls to the commit_payload endpoint. Once the pending_payload_commits Queue reach a certain 
limit, it needs to be alerted.
- "block_storage": Either IPFS to FILECOIN. This variable represents where each of the block or payload needs to be stored
- "payload_storage": Same as that of the block_storage
- "container_height": The no.of DAG blocks each container needs to hold.
- "backup_target": Defaults to FILECOIN for now. This variable represents where we want to backup the containers

