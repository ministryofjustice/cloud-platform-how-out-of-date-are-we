FROM golang:1.17-stretch AS helm_release_builder

ENV CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build .

FROM alpine:3.11.0

ENV \
  HELM_VERSION=3.4.0

RUN \
  apk add \
    --no-cache \
    --no-progress \
    --update \
    bash \
    curl \
    git

WORKDIR /app

COPY --from=helm_release_builder /app/helm-releases ./

# Install helm
RUN curl -sL https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz \
    | tar -xzC /usr/local/bin --strip-components 1 linux-amd64/helm \
    && chmod 755 /usr/local/bin/helm

RUN addgroup -g 1000 -S appgroup \
  && adduser -u 1000 -S appuser -G appgroup

RUN chown -R appuser:appgroup /app

USER 1000

# Install helm-whatup
RUN helm plugin install https://github.com/fabmation-gmbh/helm-whatup

# Add helm repos
RUN helm repo add jetstack https://charts.jetstack.io/; \
  helm repo add concourse https://concourse-charts.storage.googleapis.com/; \
  helm repo add cloud-platform https://ministryofjustice.github.io/cloud-platform-helm-charts; \
  helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx; \
  helm repo add prometheus https://prometheus-community.github.io/helm-charts; \
  helm repo add autoscaler https://kubernetes.github.io/autoscaler; \
  helm repo add vmware-tanzu https://vmware-tanzu.github.io/helm-charts; \
  helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com; \
  helm repo add kiam https://uswitch.github.io/kiam-helm-charts/charts; \
  helm repo add bitnami https://charts.bitnami.com/bitnami; \
  helm repo add kube-state-metrics https://kubernetes.github.io/kube-state-metrics; \
  helm repo add grafana https://grafana.github.io/helm-charts; \
  helm repo add eks https://aws.github.io/eks-charts; \
  helm repo add hashicorp https://helm.releases.hashicorp.com; \
  helm repo add stable https://charts.helm.sh/stable; \
  helm repo add oteemocharts https://oteemo.github.io/charts;
