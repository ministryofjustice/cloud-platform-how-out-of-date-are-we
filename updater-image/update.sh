#!/bin/sh

NAMESPACE="how-out-of-date-are-we"
API_KEY_SECRET="how-out-of-date-are-we-api-key"

main() {
  set_api_key
  helm_releases
  terraform_modules
  documentation
}

set_api_key() {
  aws s3 cp s3://${KUBECONFIG_S3_BUCKET}/${KUBECONFIG_S3_KEY} /tmp/kubeconfig
  kubectl config use-context ${KUBE_CLUSTER}
  export API_KEY=$(kubectl -n ${NAMESPACE} get secrets ${API_KEY_SECRET} -o jsonpath='{.data.token}' | base64 -d)
}

helm_releases() {
  helm repo update
  curl -H "X-API-KEY: ${API_KEY}" -d "$(helm whatup -o json)" ${DATA_URL}/helm_whatup
}

terraform_modules() {
  git clone --depth 1 https://github.com/ministryofjustice/cloud-platform-environments.git
  (
    cd cloud-platform-environments
    curl -H "X-API-KEY: ${API_KEY}" -d "$(/app/module-versions.rb)" ${DATA_URL}/terraform_modules
  )
}

documentation() {
  curl -H "X-API-KEY: ${API_KEY}" -d "$(/app/documentation-pages-to-review.rb)" ${DATA_URL}/documentation
}

main
