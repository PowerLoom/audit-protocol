echo "Building common utils"
cd goutils
go build ./...
cd ../

echo "Building Pruning Service"
cd pruning-archival
go build .
cd ../

echo "Building Payload Commit Service"
cd payload-commit
go build .

cd ../
echo "Building DAG Verifier Service"
cd dag-verifier
go build .

cd ../
echo "Building Token Aggregator Service"
cd token-aggregator
go build .
