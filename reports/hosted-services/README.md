# List 'hosted services' in the cluster

Output service information for all namespaces in the cluster.

This information includes:

* Namespace name
* Application name
* Business unit
* Service team
* Team slack channel
* Github repo(s)
* Domain name(s)

The main package in this report will perform the following steps:

- Fetch the kubeconfig from the s3 bucket 
- authenticate to the kubernetes cluster and set the current context to `ctx` env variable
- get all namespaces and build a map of namespaces
- get all ingresses and build a map of ingresses
- Add slice of domains to the corresponding namespace
- post them as json to the `hosted_services` endpoint

## Environment variables

You can see from the codebase, a number of environment variables are required to run the program. These are:

- bucket - The bucket name that hosts a kubeconfig file, commonly used in cloud-platform.

- context - The kubernetes cluster to which the report  

- kubeconfig - The kubeconfig name in the variable bucket

- region - AWS Region to get the s3 bucket

- hoodawBucket - The bucket name that hosts the hoodaw reports

## How to test locally

From the root of the HOODAW directory, run `make dev-server`. Ensure your environment variables are set i.e. hoodawHost=http://localhost:4567.

Either run `go run main.go` with arguments specified, or simply run `go test -v ./...`.
