module github.com/powerloom/audit-prototol-private/token-aggregator

go 1.19

require (
	github.com/ethereum/go-ethereum v1.10.16
	github.com/go-redis/redis/v8 v8.11.5
	github.com/sirupsen/logrus v1.9.0
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/powerloom/audit-prototol-private/goutils v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)

replace github.com/powerloom/audit-prototol-private/goutils => ../goutils
