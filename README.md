# audit-protocol-sia

make localnet
uvicorn main:app --port 9000
python3 deal_watcher.py
python3 retrieval_worker.py


