#!/bin/bash

# Copy the maticvigil config files to the right location
#mkdir -p /root/.maticvigil
#cp docker/settings.json /root/.maticvigil/
#cp docker/account_info.json /root/.maticvigil/

#cp docker_settings.json settings.json

export PYTHONPATH=$(pwd)
#alias python='python3'

#echo sleeping for 7 secs
#sleep 7

# Run tests to check redis connection, maticvigil sdk and ipfs daemon connection
#python3 tests/test_maticvigil_connection.py

./build.sh

python init_rabbitmq.py

#echo 'waiting to register projects...';

#sleep 10;

#echo 'registering projects...';

#python cli_cmd.py updatestoredprojectids docker-UNISWAPV2-ph15-prod

#python register_pair_projects_for_indexing.py

echo 'starting pm2...';

pm2 start pm2.config.js

pm2 logs --lines 1000