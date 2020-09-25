#!/bin/sh

main() {
  helm_releases
  terraform_modules
  documentation
  repositories
  hosted_services
}

set_kube_context() {
  aws s3 cp s3://${KUBECONFIG_S3_BUCKET}/${KUBECONFIG_S3_KEY} /tmp/kubeconfig
  kubectl config use-context ${KUBE_CLUSTER}
  export KUBE_CONTEXT=${KUBE_CLUSTER} # So that we can tell if this function has been called
}

helm_releases() {
  if [[ -z "${KUBE_CONTEXT}" ]]; then
    set_kube_context
  fi
  helm repo update
  curl -H "X-API-KEY: ${HOODAW_API_KEY}" -d "$(/app/helm-releases.rb)" ${DATA_URL}/helm_whatup
}

terraform_modules() {
  git clone --depth 1 https://github.com/ministryofjustice/cloud-platform-environments.git
  (
    cd cloud-platform-environments
    curl -H "X-API-KEY: ${HOODAW_API_KEY}" -d "$(/app/module-versions.rb)" ${DATA_URL}/terraform_modules
  )
}

documentation() {
  curl -H "X-API-KEY: ${HOODAW_API_KEY}" -d "$(/app/documentation-pages-to-review.rb)" ${DATA_URL}/documentation
}

repositories() {
  curl -H "X-API-KEY: ${HOODAW_API_KEY}" -d "$(cloud-platform-repository-checker)" ${DATA_URL}/repositories
}

main
