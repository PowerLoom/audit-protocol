module go-cli

go 1.18

replace github.com/powerloom/goutils => ../goutils

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/powerloom/goutils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.0
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)
