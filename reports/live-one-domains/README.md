# Cloud Platform Remaining Live-1 Domains

Ouputs a JSON report showing
- A list of remaining live1 domains within the cloud platform live cluster, listing Namespace, Ingress Name and the Host which is still using live-1 DNS.

Output service information for all namespaces in the cluster.

This information includes:

* Namespace
* Ingress Name
* Domain URL
* Created Date

The main package in this report will perform the following steps:

- fetch the kubeconfig from the s3 bucket
- authenticate to the kubernetes cluster and set the current context to `ctx` env variable
- get all ingresses and build a map of ingresses
- search the ingress map for the remaining live1 domains within the cluster
- post them as json to the `live_one_domains` endpoint

## Environment variables

You can see from the codebase, a number of environment variables are required to run the program. These are:

- hoodawBucket  = The bucket name where json reports are stored"

- bucket - The bucket name that hosts a kubeconfig file, commonly used in cloud-platform.

- context - The kubernetes cluster to which the report  

- kubeconfig - The kubeconfig name in the variable bucket

- region - AWS Region to get the s3 bucket

## How to test locally

From the root of the HOODAW directory, run `make dev-server`. Ensure your environment variables are set i.e. hoodawHost=http://localhost:4567.

Either run `go run main.go` with arguments specified, or simply run `go test -v ./...`.
