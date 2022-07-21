#!/usr/bin/env bash

## Liveness probe check for kube-server

set -e

curl -k \
     --cacert client-certs/ca.crt \
     --key client-certs/client.key \
     --cert client-certs/client.crt \
     "https://localhost:8443/v1/healthz"