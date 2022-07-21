#!/usr/bin/env bash

## Self-signed certs for kube-server
kubectl delete secret certificates-tls-secret -n kube-server-dev ; \
kubectl create secret generic certificates-tls-secret -n kube-server-dev \
--from-file=./certs/kube-server/server.crt \
--from-file=./certs/kube-server/server.key \
--from-file=./certs/kube-server/ca.crt

kubectl delete secret certificates-tls-secret -n kube-server ; \
kubectl create secret generic certificates-tls-secret -n kube-server \
--from-file=./certs/kube-server/server.crt \
--from-file=./certs/kube-server/server.key \
--from-file=./certs/kube-server/ca.crt

## Self-signed client certs for kube-server to allow health checks
kubectl delete secret client-tls-secret -n kube-server ; \
kubectl create secret generic client-tls-secret -n kube-server \
--from-file=./certs/kube-server/client.crt \
--from-file=./certs/kube-server/client.key \
--from-file=./certs/kube-server/ca.crt

kubectl delete secret client-tls-secret -n kube-server-dev ; \
kubectl create secret generic client-tls-secret -n kube-server-dev \
--from-file=./certs/kube-server/client.crt \
--from-file=./certs/kube-server/client.key \
--from-file=./certs/kube-server/ca.crt

## Self-signed client certs for accessing Redis
kubectl delete secret redis-tls-client -n kube-server-dev ; \
kubectl create secret generic redis-tls-client -n kube-server-dev \
--from-file=./certs/redis/redis-client.crt \
--from-file=./certs/redis/redis-client.key \
--from-file=./certs/redis/redis-ca.crt

kubectl delete secret redis-tls-client -n kube-server ; \
kubectl create secret generic redis-tls-client -n kube-server \
--from-file=./certs/redis/redis-client.crt \
--from-file=./certs/redis/redis-client.key \
--from-file=./certs/redis/redis-ca.crt


## Self-signed certs for redis
kubectl delete secret certificates-tls-secret -n redis ; \
kubectl create secret generic certificates-tls-secret -n redis \
--from-file=./certs/redis/server.crt \
--from-file=./certs/redis/server.key \
--from-file=./certs/redis/redis-ca.crt