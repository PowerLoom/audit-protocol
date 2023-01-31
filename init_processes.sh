#!/bin/bash

export PYTHONPATH=$(pwd)

./build.sh

python init_rabbitmq.py

echo 'starting pm2...';

pm2 start pm2.config.js

pm2 logs --lines 1000