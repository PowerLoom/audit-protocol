echo "Building common utils"
cd go || exit
go build ./...

echo "Building Payload Commit Service"
cd payload-commit || exit
go build .

echo "Building Pruning Service"
cd ../pruning || exit
go build .
