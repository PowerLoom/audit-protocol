# Pruning-Archival Service

This service is responsible for archival and pruning of the DAG segments of projects' DAG chain.
Sequence diagram and more information about service can be found [here](../../docs/Introduction.md)

## Testing locally

- cd to `deploy` project
- deploy complete setup locally by running deploy project

```shell
./build-dev.sh
```

- come out of `deploy` project and cd to `audit-protocol/go/pruning-archival` service directory.

- Run `main.go`. This will start the service pruning-archival service and start listening for events.
```shell
# CONFIG_PATH="/home/user/audit-protocol"
CONFIG_PATH=absolute/path/to/audit-protocol go run main.go
```
- Run `./testing/simulate/simulate.go` to add data in local redis cache and create a rabbitmq task/message to simulate the pruning and archival process.
```shell
CONFIG_PATH=absolute/path/to/audit-protocol go run testing/simulate/simulate.go
```
- Check the logs to see if the task is consumed and processed by the service and then Run `./testing/test/test.go` which will check if the segments are pruned and output is as expected.
```shell
CONFIG_PATH=absolute/path/to/audit-protocol go run testing/test/test.go
```

## Testing improvements

- currently tests are run manually as we have to run `test.go` and `simulate.go` manually and keep an eye on the logs to see if the service has processed the task.

