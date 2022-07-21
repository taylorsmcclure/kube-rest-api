#!/usr/bin/env bash

## Easy way to interact with kube-server via mTLS

set -e

curl --silent -k \
     --cacert certs/kube-server/ca.crt \
     --key certs/kube-server/client.key \
     --cert certs/kube-server/client.crt \
     "$@"