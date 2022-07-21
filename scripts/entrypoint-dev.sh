#!/usr/bin/env bash

# entrypoint for dev deployment

set -e

echo "Starting server..."

go run cmd/kube-server/main.go --port=${PORT} --verbose \
--ca=server-certs/ca.crt \
--cert=server-certs/server.crt \
--key=server-certs/server.key \
--raddr=${REDIS_ADDR} \
--rca=redis-certs/redis-ca.crt \
--rcert=redis-certs/redis-client.crt \
--rkey=redis-certs/redis-client.key 
