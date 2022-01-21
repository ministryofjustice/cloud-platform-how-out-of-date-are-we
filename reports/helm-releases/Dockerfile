FROM golang:1.17-alpine3.15

RUN \
  apk add \
    --no-cache \
    --no-progress \
    --update \
    bash \
    curl \
    git

ENV \
  HELM_VERSION=3.4.0 \
  KUBECTL_VERSION=1.20.7

# kubectl
RUN curl -sLo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl \
  && chmod 755 /usr/local/bin/kubectl

# Install helm
RUN curl -sL https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz \
    | tar -xzC /usr/local/bin --strip-components 1 linux-amd64/helm \
    && chmod 755 /usr/local/bin/helm

RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build .
RUN pwd && ls

RUN chown -R 1000:1000 /app
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
  helm repo add vmware https://kubernetes-charts.banzaicloud.com; \
  helm repo add banzaicloud https://kubernetes-charts.banzaicloud.com; \
  helm repo add kiam https://uswitch.github.io/kiam-helm-charts/charts; \
  helm repo add bitnami https://charts.bitnami.com/bitnami; \
  helm repo add kube-state-metrics https://kubernetes.github.io/kube-state-metrics; \
  helm repo add grafana https://grafana.github.io/helm-charts; \
  helm repo add eks https://aws.github.io/eks-charts; \
  helm repo add hashicorp https://helm.releases.hashicorp.com
