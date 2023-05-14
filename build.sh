echo "Building common utils"
cd go || exit
go build ./...

echo "Building Payload Commit Service"
cd payload-commit || return
go build .

echo "Building Pruning Service"
cd ../pruning || return
go build .
