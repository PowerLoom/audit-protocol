echo "Building common utils"
cd goutils/settings
go build .
cd ../..

echo "Building Pruning Service"
cd go-pruning-archival-service
go build .
cd ../

echo "Building Payload Commit Service"
cd go-payload-commit-service
go build .

cd ../
echo "Building DAG Verifier Service"
cd dag_verifier
go build .
