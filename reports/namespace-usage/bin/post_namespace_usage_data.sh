#!/bin/sh

set -euo pipefail

# Authenticate to the cluster
aws s3 cp s3://${KUBECONFIG_S3_BUCKET}/${KUBECONFIG_S3_KEY} /tmp/kubeconfig
kubectl config use-context ${KUBE_CLUSTER}

/app/bin/namespace-usage-reporter.rb > namespace-usage.json

curl \
  --http1.1 \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: ${HOODAW_API_KEY}" \
  -d @namespace-usage.json \
  ${HOODAW_HOST}/namespace_usage
