#!/bin/sh

main() {
  set_api_key
  helm_releases
}

set_api_key() {
  aws s3 cp s3://${KUBECONFIG_S3_BUCKET}/${KUBECONFIG_S3_KEY} /tmp/kubeconfig
  kubectl config use-context ${KUBE_CLUSTER}
  export API_KEY=$(kubectl -n how-out-of-date-are-we get secrets how-out-of-date-are-we-api-key -o jsonpath='{.data.token}' | base64 -d)
}


helm_releases() {
  helm repo update
  curl -H "X-API-KEY: ${API_KEY}" -d "$(helm whatup -o json)" ${HTTP_ENDPOINT}
}
