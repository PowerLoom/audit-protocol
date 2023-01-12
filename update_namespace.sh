#!/bin/bash

NAMESPACE="UNISWAPV2-ph15-prod"

# Now take action

grep $NAMESPACE token-aggregator/main.go
if [ $? -eq 0 ]; then
    echo "Namespace $NAMESPACE" is already updated
    exit 0
else
    echo "Updating namespace to $NAMESPACE"
fi

sed -i "s/const NAMESPACE string = \"UNISWAPV2\"/const NAMESPACE string = \"$NAMESPACE\"/g" token-aggregator/main.go

sed -i "s/NAMESPACE = 'UNISWAPV2'/NAMESPACE = '$NAMESPACE'/g" pair_data_aggregation_service.py utils/redis_keys.py

sed -i "s/cache:indexesRequested/cache:indexesRequested:$NAMESPACE/g" proto_sliding_window_cacher_service.py

echo "Building token-aggregator after changes"
cd token-aggregator
go build .
cd ..
