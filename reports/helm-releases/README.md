# Outdated Helm Releases

Uses [Helm Whatup] to compare running vs. latest Helm charts in our clusters.

[Helm Whatup]: https://github.com/fabmation-gmbh/helm-whatup

The main package in this report will perform the following steps:

- Gain access to a Kubernetes cluster using a config file stored in an S3 bucket.
- kube context switch to Live/manager/live-1 clusters and output the results of `helm whatup`.
- post them as json to the `helm_whatup` endpoint.

The only pre-requisite/requirement for this report to run is the existence of a kubeconfig file in an s3 bucket.

## Environment variables

You can see from the codebase, a number of environment variables are required to run the program. These are:

- hoodawAPIKey: The API key of the "How out of date are we application" (HOODAW)

- hoodawHost: The hostname of the HOODAW hostname i.e. https://reports.cloud-platform.service.justice.gov.uk

- bucket - The bucket name that hosts a kubeconfig file, commonly used in cloud-platform.

- ctxLive/ctxManager/ctxLive_1 - Context used to switch clusters

## How to test locally

From the root of the HOODAW directory, run `make dev-server`. Ensure your environment variables are set i.e. hoodawHost=http://localhost:4567.

Either run `go run main.go` with arguments specified, or simply run `go test -v ./...`.
