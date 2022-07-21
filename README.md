# kube-server

This is a Go based REST server that interacts with the Kubernetes API, specifically with Deployments and their replicas.

## Prerequisites

- Go
- [minikube](https://minikube.sigs.k8s.io/docs/start/)
- [Helm](https://helm.sh/docs/intro/install/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

## Usage

For a quickstart you can `make start` This will:

- Build the kube-server binary for Linux 64-bit
- Start the minikube container and give you a kubeconfig
- Builds a kube-server docker image within minikube for local use
- Creates self-signed certs for mTLS for kube-server and Redis
- Creates multiple demonstration Deployments in minikube
- Deploys the certificates as secrets in the kube-server and Redis namespaces
- Installs Redis helm chart
- Installs local kube-server helm chart

It may take a couple moments for Redis to initialize. When it's healthy you can access the kube-server with an mTLS wrapper script.

Here's some examples using the script:

```shell
./scripts/client-tls.sh https://localhost:8443/v1/healthz
./scripts/client-tls.sh https://localhost:8443/v1/deployments
./scripts/client-tls.sh https://localhost:8443/v1/replicas/busybox-test/busybox-deployment
./scripts/client-tls.sh -X POST https://localhost:8443/v1/replicas/busybox-test/busybox-deployment -H 'Content-Type: application/json' -d '{replica_size:5}'
./scripts/client-tls.sh https://localhost:8443/v1/replicas/busybox-test/busybox-deployment 
```

## Local Development

For local development you can use the following workflow:

1. Initialize with `minikube start`
2. Make changes to the code
3. Use `go run-dev` to spawn a server in your teminal
4. Use `https://localhost:8888` to send requests to test
5. Run unit-tests `make unit-tests`
6. Deploy to minikube with `make deploy`
7. `curl` endpoints you want to test
8. Run integration test `make int-tests`

## Endpoints

Here's the available endpoints

### `v1/healthz`

**GET**

**Response**
```json
{
  "http_response_code": 200,
  "kubernetes_api_status": "ok",
  "application_version": "0.1.0"
}
```

### `v1/deployments`

You can also filter deployments by namespace like: `/v1/replicas/deployments?namespace=busybox-test`

**GET**

**Response**

```json
{
  "http_response_code": 200,
  "deployments": [
    {
      "deployment_name": "busybox-deployment0",
      "namespace": "busybox-test"
    },
    {
      "deployment_name": "busybox-deployment1",
      "namespace": "busybox-test"
    },
  ]
}
```

### `v1/replicas/:namespace/:deployment`

**GET**

Gets the replicas of the specified deployment.

**Response**

```json
{
  "namespace": "busybox-test",
  "deployment_name": "busybox-deployment0",
  "current_replicas": 5,
  "desired_replicas": 5,
  "state_drift": false,
  "http_status_code": 200
}
```

**POST**

Sets the replicas of the specified deployment.

**Request**

```json
{
    "replica_size":4
}
```

**Response**

```json
{
  "namespace": "busybox-test",
  "deployment_name": "busybox-deployment0",
  "current_replicas": 5,
  "desired_replicas": 5,
  "requested_replicas": 4,
  "state_drift": false,
  "http_status_code": 200
}
```