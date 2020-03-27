#!/bin/sh

set -euo pipefail

aws s3 cp s3://${KUBECONFIG_S3_BUCKET}/${KUBECONFIG_S3_KEY} /tmp/kubeconfig
kubectl config use-context ${KUBE_CLUSTER}

helm repo update

curl -H "X-API-KEY: $(kubectl -n how-out-of-date-are-we get secrets how-out-of-date-are-we-api-key -o jsonpath='{.data.token}' | base64 -d)" -d "$(helm whatup -o json)" ${HTTP_ENDPOINT}
