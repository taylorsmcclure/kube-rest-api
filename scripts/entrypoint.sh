#!/usr/bin/env bash

## entrypoint for kube-server deployment

set -e

echo "Starting server..."

./kube-server-linux --port=${PORT} \
--ca=server-certs/ca.crt \
--cert=server-certs/server.crt \
--key=server-certs/server.key \
--raddr=${REDIS_ADDR} \
--rca=redis-certs/redis-ca.crt \
--rcert=redis-certs/redis-client.crt \
--rkey=redis-certs/redis-client.key 
