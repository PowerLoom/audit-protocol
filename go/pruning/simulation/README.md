# Running simulation for the Pruning service

### Prerequisites

1. Docker should be installed
2. IPFS node running on docker with exposed port `5001`
3. Golang should be installed

### Running simulation

For running simulation first we export states of the source instance which is already running PowerLoom system.
States here are `protocol state` and `ipfs local cache`

1. Run `pre_simulation_setup.sh` to setup the environment for pruning

```shell
cd audit-protocol/go/pruning/simulation

# replace "host" with source instance host and "ip" with source instance ip
# replace "name_of_ipfs_container" with the name of ipfs container running on current instance (ipfs local cache state is copied here)
./pre_simulation_setup.sh -s host@ip -n name_of_ipfs_container
```

2. Run `simulation.go` to run the simulation

```shell
cd audit-protocol/go/pruning/simulation

# if you are running locally (ipfs is running on localhost)
CONFIG_PATH="../../../" LOCAL_TESTING=true go run simulation.go

# if you are running on remote instance with powerloom setup or ipfs is running on different host
# make sure to change the host of ipfs url in settings.json accordingly
CONFIG_PATH="../../../" go run simulation.go
```
