# Get all ingress resources that don't have an annotation

This report was created out of the requirement to capture all ingress resources that don't have a required annotation for migration to a new cluster. This annotation is very specific, so the report checks for the existence of the key: "external-dns.alpha.kubernetes.io/aws-weight", this annotation can be changed by setting the flag `annotation="<insert-annotation>"`.

The main package in this report will perform the following steps:

- authenticate to the "live-1" Kubernetes cluster.
- grab all ingress resources that don't have the required annotation.
- post them as json to the `ingress_weighting` endpoint, again this can be changed by a flag.

The only pre-requisite/requirement for this report to run is the existence of a kubeconfig file in an s3 bucket, and for the `current-context` to be set.

## Environment variables

You can see from the codebase, a number of environment variables are required to run the program. These are:

- hoodawAPIKey: The API key of the "How out of date are we application" (HOODAW)

- hoodawHost: The hostname of the HOODAW hostname i.e. https://reports.cloud-platform.service.justice.gov.uk

- bucket - The bucket name that hosts a kubeconfig file, commonly used in cloud-platform.

- configFile - The kubeconfig name in the variable bucket. This should have the current-context set or this report will fail.

## How to test locally

From the root of the HOODAW directory, run `make dev-server`. Ensure your environment variables are set i.e. hoodawHost=http://localhost:4567.

Either run `go run main.go` with arguments specified, or simply run `go test -v ./...`.
