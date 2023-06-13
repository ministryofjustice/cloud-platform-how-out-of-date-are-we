# Namespace Usage

Ouputs a JSON report showing 
- the CPU requested vs. used, for all namespaces
- the Memory requested vs. used for all namespaces
- the number of Hard limits set for pods vs current number of running pods for all namespaces
- number of containers for all namespaces

The main package in this report will perform the following steps:

- fetch the kubeconfig from the s3 bucket 
- authenticate to the kubernetes cluster and set the current context to `ctx` env variable
- get all namespaces 
- get all pods and create a resource requests Map of NamespaceResource type
- get all pod metrics and create a resource usage map of NamespaceResource type 
- get all resourcequota from cluster and create a quota map
- build a usageReport with all the data required i.e cpu, memory and pods
- post them as json to the `namespace_usage` endpoint

## Environment variables

You can see from the codebase, a number of environment variables are required to run the program. These are:

- bucket - The bucket name that hosts a kubeconfig file, commonly used in cloud-platform.

- context - The kubernetes cluster to which the report  

- hoodawAPIKey: The API key of the "How out of date are we application" (HOODAW)

- hoodawEndpoint: The endpoint of the HOODAW Application "namespace_usage"

- hoodawHost: The hostname of the HOODAW hostname i.e. https://reports.cloud-platform.service.justice.gov.uk

- kubeconfig - The kubeconfig name in the variable bucket

- region - AWS Region to get the s3 bucket

- kubeCfgPath - Path in which the kubeconfig has to be stored

## How to test locally

From the root of the HOODAW directory, run `make dev-server`. Ensure your environment variables are set i.e. hoodawHost=http://localhost:4567.

Either run `go run main.go` with arguments specified, or simply run `go test -v .`.
