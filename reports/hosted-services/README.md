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

* get all namespaces and build a map of namespaces
* get all ingresses and build a map of ingresses
* Add slice of domains to the corresponding namespace
* post them as json to the `hosted_services` hoodaw s3 bucket

## Environment variables

You can see from the codebase, a number of environment variables are required to run the program. These are:

* bucket - The bucket name that we will post the data to.

* context - The kubernetes cluster to which the report

## How to test locally

Run `go run main.go` with arguments specified and navigate to `/hosted_services`
