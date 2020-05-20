#!/bin/sh

NAMESPACE="how-out-of-date-are-we"
API_KEY_SECRET="how-out-of-date-are-we-api-key"

main() {
  set_api_key
  helm_releases
  terraform_modules
  documentation
  repositories
}

set_kube_context() {
  aws s3 cp s3://${KUBECONFIG_S3_BUCKET}/${KUBECONFIG_S3_KEY} /tmp/kubeconfig
  kubectl config use-context ${KUBE_CLUSTER}
  export KUBE_CONTEXT=${KUBE_CLUSTER} # So that we can tell if this function has been called
}

# Fetch the API key from a kubernetes secret in the live-1 cluster, if no API_KEY environment variable is set.
# This allows us to bypass the kubernetes secret lookup in development, but setting an API_KEY env. var.
set_api_key() {
  if [[ -z "${API_KEY}" ]]; then
    echo "Fetching API_KEY from kubernetes secret"
    set_kube_context
    export API_KEY=$(kubectl -n ${NAMESPACE} get secrets ${API_KEY_SECRET} -o jsonpath='{.data.token}' | base64 -d)
  fi
}

helm_releases() {
  if [[ -z "${KUBE_CONTEXT}" ]]; then
    set_kube_context
  fi
  helm repo update
  curl -H "X-API-KEY: ${API_KEY}" -d "$(/app/helm-releases.rb)" ${DATA_URL}/helm_whatup
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

repositories() {
  curl -H "X-API-KEY: ${API_KEY}" -d "$(cloud-platform-repository-checker)" ${DATA_URL}/repositories
}

main
