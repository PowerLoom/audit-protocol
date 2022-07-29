#!/bin/bash

NAMESPACE="SUSHISWAP-stg-1"

# Now take action
echo "Updating namespace to $NAMESPACE"

sed -i "s/const NAMESPACE string = \"UNISWAPV2\"/const NAMESPACE string = \"$NAMESPACE\"/g" dag_verifier/dag_verifier.go

sed -i "s/NAMESPACE = 'UNISWAPV2'/NAMESPACE = $NAMESPACE/g" pair_data_aggregation_service.py utils/redis_keys.py

sed -i "s/cache:indexesRequested/cache:indexesRequested:$NAMESPACE/g" proto_sliding_window_cacher_service.py

sed -i "s/audit-protocol-proto-indexer/$NAMESPACE-audit-protocol-proto-indexer/g" pm2.config.js
sed -i "s/audit-protocol-dag-verifier/$NAMESPACE-audit-protocol-dag-verifier/g" pm2.config.js

echo "Building dag_verifier after changes"
cd dag_verifier
go build .
cd ..
