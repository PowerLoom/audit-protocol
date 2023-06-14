#!/bin/bash

#This script is run from high level docker-compose. Refer to https://github.com/PowerLoom/deploy
set -e

echo 'populating setting from environment values...';

if [ -z "$SIGNER_ACCOUNT_ADDRESS" ]; then
    echo "SIGNER_ACCOUNT_ADDRESS not found, please set this in your .env!";
    exit 1;
fi

if [ -z "$SIGNER_ACCOUNT_PRIVATE_KEY" ]; then
    echo "SIGNER_ACCOUNT_PRIVATE_KEY not found, please set this in your .env!";
    exit 1;
fi

echo "Found SIGNER ACCOUNT ADDRESS ${SIGNER_ACCOUNT_ADDRESS}";

if [ "$PROST_RPC_URL" ]; then
    echo "Found PROST_RPC_URL ${PROST_RPC_URL}";
fi

if [ "$IPFS_URL" ]; then
    echo "Found IPFS_URL ${IPFS_URL}";
fi

if [ "$PROTOCOL_STATE_CONTRACT" ]; then
    echo "Found PROTOCOL_STATE_CONTRACT ${PROTOCOL_STATE_CONTRACT}";
fi

if [ "$RELAYER_HOST" ]; then
    echo "Found RELAYER_HOST ${RELAYER_HOST}";
fi

if [ "$SLACK_REPORTING_URL" ]; then
    echo "Found SLACK_REPORTING_URL ${SLACK_REPORTING_URL}";
fi

if [ "$POWERLOOM_REPORTING_URL" ]; then
    echo "Found SLACK_REPORTING_URL ${POWERLOOM_REPORTING_URL}";
fi

if [ "$WEB3_STORAGE_TOKEN" ]; then
    echo "Found WEB3_STORAGE_TOKEN ${WEB3_STORAGE_TOKEN}";
fi

cp settings.example.json settings.json

export namespace=UNISWAPV2
export prost_rpc_url="${PROST_RPC_URL:-https://rpc-prost1b.powerloom.io}"

export ipfs_url="${IPFS_URL:-/dns/ipfs/tcp/5001}"
export ipfs_api_key="${IPFS_API_KEY:-}"
export ipfs_api_secret="${IPFS_API_SECRET:-}"

export protocol_state_contract="${PROTOCOL_STATE_CONTRACT:-0x19B7f6c458bBd6faDc36D39Ae77B132CC3b5B159}"
export relayer_host="${RELAYER_HOST:-https://relayer-test1b.powerloom.io}"

export slack_reporting_url="${SLACK_REPORTING_URL:-}"
export powerloom_reporting_url="${POWERLOOM_REPORTING_URL:-}"

export web3_storage_token="${WEB3_STORAGE_TOKEN:-}"

# If IPFS_URL is empty, clear IPFS API key and secret
if [ -z "$IPFS_URL" ]; then
    ipfs_api_key=""
    ipfs_api_secret=""
fi

echo "Using Namespace: ${namespace}"
echo "Using Prost RPC URL: ${prost_rpc_url}"
echo "Using IPFS URL: ${ipfs_url}"
echo "Using IPFS API KEY: ${ipfs_api_key}"
echo "Using protocol state contract: ${protocol_state_contract}"
echo "Using relayer host: ${relayer_host}"
echo "Using slack reporting url: ${slack_reporting_url}"
echo "Using powerloom reporting url: ${powerloom_reporting_url}"
echo "Using web3 storage token: ${web3_storage_token}"

sed -i'.backup' "s#relevant-namespace#$namespace#" settings.json

sed -i'.backup' "s#https://prost-rpc-url#$prost_rpc_url#" settings.json

sed -i'.backup' "s#ipfs-writer-url#$ipfs_url#" settings.json
sed -i'.backup' "s#ipfs-writer-key#$ipfs_api_key#" settings.json
sed -i'.backup' "s#ipfs-writer-secret#$ipfs_api_secret#" settings.json

sed -i'.backup' "s#ipfs-reader-url#$ipfs_url#" settings.json
sed -i'.backup' "s#ipfs-reader-key#$ipfs_api_key#" settings.json
sed -i'.backup' "s#ipfs-reader-secret#$ipfs_api_secret#" settings.json

sed -i'.backup' "s#web3-storage-token#$web3_storage_token#" settings.json
sed -i'.backup' "s#protocol-state-contract#$protocol_state_contract#" settings.json

sed -i'.backup' "s#signer-account-address#$SIGNER_ACCOUNT_ADDRESS#" settings.json
sed -i'.backup' "s#signer-account-private-key#$SIGNER_ACCOUNT_PRIVATE_KEY#" settings.json

sed -i'.backup' "s#account-address#$SIGNER_ACCOUNT_ADDRESS#" settings.json

sed -i'.backup' "s#https://relayer-url#$relayer_host#" settings.json
sed -i'.backup' "s#https://slack-reporting-url#$slack_reporting_url#" settings.json
sed -i'.backup' "s#https://powerloom-reporting-url#$powerloom_reporting_url#" settings.json

echo 'settings has been populated!'
