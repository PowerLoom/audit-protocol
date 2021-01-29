#!/bin/bash

mkdir -p /root/.maticvigil
cp docker/settings.json /root/.maticvigil/
cp docker/account_info.json /root/.maticvigil/
cp docker_settings.example.json settings.json
export PYTHONPATH=$(pwd)
echo sleeping for 7 secs
sleep 7
python3 tests/test_maticvigil_connection.py

