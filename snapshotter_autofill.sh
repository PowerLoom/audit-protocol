#!/bin/bash

#This script is run from high level docker-compose. Refer to https://github.com/PowerLoom/deploy

set -e

echo 'populating setting from environment values...';

if [ -z "$ANCHOR_CHAIN_RPC_URL" ]; then
    echo "ANCHOR_CHAIN_RPC_URL not found, please set this in your .env!";
    exit 1;
fi

if [ -z "$UUID" ]; then
    echo "UUID not found, please set this in your .env!";
    exit 1;
fi

echo "Got ANCHOR CHAIN RPC URL: ${ANCHOR_CHAIN_RPC_URL}"

echo "Got UUID: ${UUID}"

echo "Got WEB3_STORAGE_TOKEN: ${WEB3_STORAGE_TOKEN}"

cp settings.example.json settings.json

export namespace=UNISWAPV2-ph15-prod

echo "Using Namespace: ${namespace}"

sed -i "s|relevant-namespace|$namespace|" settings.json

sed -i "s|https://rpc-url|$ANCHOR_CHAIN_RPC_URL|" settings.json

sed -i "s|generated-uuid|$UUID|" settings.json

sed -i "s|web3-storage-token|$WEB3_STORAGE_TOKEN|" settings.json

echo 'settings has been populated!'