#!/usr/bin/env bash

set -e

mkdir -p certs/kube-server
mkdir -p certs/redis

## For kube-server application
# Generate self signed root CA cert
openssl req -nodes -x509 -newkey rsa:2048 -sha256 -keyout certs/kube-server/ca.key -out certs/kube-server/ca.crt -subj "/C=US/ST=HI/L=Waialua/O=kubeServer/OU=root/CN=kube-server.taylorm.cc"

# Generate server cert to be signed
openssl req -nodes -newkey rsa:2048 -sha256 -keyout certs/kube-server/server.key -out certs/kube-server/server.csr -subj "/C=US/ST=HI/L=Waialua/O=kubeServer/OU=server/CN=kube-server.taylorm.cc"

# Sign the server cert
openssl x509 -req -in certs/kube-server/server.csr -CA certs/kube-server/ca.crt -CAkey certs/kube-server/ca.key -CAcreateserial -out certs/kube-server/server.crt -sha256

# Create server PEM file
cat certs/kube-server/server.key certs/kube-server/server.crt > certs/kube-server/server.pem


# Generate client cert to be signed
openssl req -nodes -newkey rsa:2048 -sha256 -keyout certs/kube-server/client.key -out certs/kube-server/client.csr -subj "/C=US/ST=HI/L=Waialua/O=kubeServer/OU=client/CN=kube-server.taylorm.cc"

# Sign the client cert
openssl x509 -req -in certs/kube-server/client.csr -CA certs/kube-server/ca.crt -CAkey certs/kube-server/ca.key -CAserial certs/kube-server/ca.srl -out certs/kube-server/client.crt -sha256

# Create client PEM file
cat certs/kube-server/client.key certs/kube-server/client.crt > certs/kube-server/client.pem

## For Redis
# Generate self signed root CA cert
openssl req -nodes -x509 -newkey rsa:2048 -sha256 -keyout certs/redis/ca.key -out certs/redis/redis-ca.crt -subj "/C=US/ST=HI/L=Waialua/O=kubeServer/OU=root/CN=redis.taylorm.cc"

# Generate server cert to be signed
openssl req -nodes -newkey rsa:2048 -sha256 -keyout certs/redis/server.key -out certs/redis/server.csr -subj "/C=US/ST=HI/L=Waialua/O=kubeServer/OU=server/CN=redis.taylorm.cc"

# Sign the server cert
openssl x509 -req -sha256 -in certs/redis/server.csr -CA certs/redis/redis-ca.crt -CAkey certs/redis/ca.key -CAcreateserial -out certs/redis/server.crt

# Create server PEM file
cat certs/redis/server.key certs/redis/server.crt > certs/redis/server.pem

# Generate client cert to be signed
openssl req -nodes -newkey rsa:2048 -sha256 -keyout certs/redis/redis-client.key -out certs/redis/redis-client.csr -subj "/C=US/ST=HI/L=Waialua/O=kubeServer/OU=redis-client/CN=redis.taylorm.cc"

# Sign the client cert
openssl x509 -req -in certs/redis/redis-client.csr -CA certs/redis/redis-ca.crt -CAkey certs/redis/ca.key -CAserial certs/redis/redis-ca.srl -out certs/redis/redis-client.crt -sha256

# Create client PEM file
cat certs/redis/redis-client.key certs/redis/redis-client.crt > certs/redis/redis-client.pem
