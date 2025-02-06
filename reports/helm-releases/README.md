# Outdated Helm Releases

Uses [Helm Whatup] to compare running vs. latest Helm charts in our clusters.

[Helm Whatup]: https://github.com/fabmation-gmbh/helm-whatup

The main package in this report will perform the following steps:

- kube context switch to Live/manager clusters and output the results of `helm whatup`.
- post them as json to our hoodaw s3 bucket.

## Environment variables

You can see from the codebase, a number of environment variables are required to run the program. These are:

- AWS creds (region, aws_role_arn & the long lived secret keys are pulled at runtime)

- hoodaw bucket - The bucket name that hosts the json we will display.

- kubeconfig bucket - The bucket name to pullthe kubeconfig from

- ctxLive/ctxManager - Context used to switch clusters

## How to test locally

Run `go run main.go` with arguments specified and navigate to `/helm_whatup`
