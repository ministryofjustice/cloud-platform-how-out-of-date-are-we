#!/bin/sh

set -euo pipefail

# Get the kubeconfig to access the clusters
aws s3 cp s3://${KUBECONFIG_S3_BUCKET}/${KUBECONFIG_S3_KEY} ${KUBECONFIG}
chmod 600 ${KUBECONFIG} # This suppresses "group readable" warnings from helm

# set context to the first cluster
kubectl config use-context ${KUBE_CLUSTER}

helm repo update
/app/bin/helm-releases.rb > helm-releases.json

curl \
  --http1.1 \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: ${HOODAW_API_KEY}" \
  -d @helm-releases.json \
  ${HOODAW_HOST}/helm_whatup
