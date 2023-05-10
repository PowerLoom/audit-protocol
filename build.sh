echo "Building common utils"
cd go
go build ./...

echo "Building Payload Commit Service"
cd payload-commit
go build .

