FROM python:alpine3.10

ENV \
  HELM_VERSION=2.14.3 \
  KUBECTL_VERSION=1.15.0

RUN pip install awscli
RUN apk add curl libc6-compat git

# Install helm
RUN curl -sL https://storage.googleapis.com/kubernetes-helm/helm-v${HELM_VERSION}-linux-amd64.tar.gz | tar -xzC /usr/local/bin --strip-components 1 linux-amd64/helm

RUN apk add bash

# Install kubectl
RUN curl -sLo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl
RUN chmod 755 /usr/local/bin/kubectl

RUN helm init --client-only
RUN helm repo add jetstack https://charts.jetstack.io
RUN helm plugin install https://github.com/bacongobbler/helm-whatup

RUN apk add ruby ruby-nokogiri
RUN gem install json_pure --no-rdoc --no-ri

WORKDIR /app
COPY update.sh /app/
COPY module-versions.rb documentation-pages-to-review.rb /app/

COPY module-versions.rb /app/

