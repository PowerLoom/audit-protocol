echo "Building common utils"
cd go
go build ./...

echo "Building Pruning Service"
cd pruning-archival
go build .
cd ../

echo "Building Payload Commit Service"
cd payload-commit
go build .

cd ../
echo "Building DAG status reporter Service"
cd dag-status-reporter
go build .

cd ../
echo "Building Token Aggregator Service"
cd token-aggregator
go build .
