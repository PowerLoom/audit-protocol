#!/bin/bash

NAMESPACE="SUSHISWAP-stg-1"

# Now take action
echo "Updating namespace to $NAMESPACE"

sed -i "s/const NAMESPACE string = \"UNISWAPV2\"/const NAMESPACE string = \"$NAMESPACE\"/g" dag_verifier/dag_verifier.go
sed -i "s/const NAMESPACE string = \"UNISWAPV2\"/const NAMESPACE string = \"$NAMESPACE\"/g" token-aggregator/main.go

sed -i "s/NAMESPACE = 'UNISWAPV2'/NAMESPACE = '$NAMESPACE'/g" pair_data_aggregation_service.py utils/redis_keys.py

sed -i "s/cache:indexesRequested/cache:indexesRequested:$NAMESPACE/g" proto_sliding_window_cacher_service.py

namespace=$(echo $NAMESPACE | tr '[:upper:]' '[:lower:]')

sed -i "s/ap-proto-indexer/ap-proto-indexer-$namespace/g" pm2.config.js
sed -i "s/ap-dag-verifier/ap-dag-verifier-$namespace/g" pm2.config.js
sed -i "s/ap-token-aggregator/ap-token-aggregator-$namespace/g" pm2.config.js

sed -i "s/NAMESPACE = \"UNISWAPV2\"/NAMESPACE = \"$NAMESPACE\"/g" register_pair_projects_for_indexing.py

echo "Building dag_verifier after changes"
cd dag_verifier
go build .
echo "Building token-aggregator after changes"
cd ../token-aggregator
go build .
cd ..
