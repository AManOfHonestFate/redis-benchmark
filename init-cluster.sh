#!/bin/bash
set -e
NODE_IPS="redis-node-1:6379 redis-node-2:6379 redis-node-3:6379 "
REPLICAS=1
PASSWORD="pass"
echo ">>> Creating cluster with nodes: $NODE_IPS"
echo ">>> Number of replicas: $REPLICAS"
docker exec -it redis-node-1 redis-cli -a "$PASSWORD" --cluster create $NODE_IPS --cluster-replicas $REPLICAS --cluster-yes
