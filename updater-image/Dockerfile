FROM python:alpine3.10

ENV \
  HELM_VERSION=2.14.3 \
  KUBECTL_VERSION=1.15.0

RUN pip install awscli \
  && apk add curl libc6-compat git bash ruby ruby-nokogiri \
  && gem install json_pure --no-rdoc --no-ri \
  && curl -sL https://storage.googleapis.com/kubernetes-helm/helm-v${HELM_VERSION}-linux-amd64.tar.gz | tar -xzC /usr/local/bin --strip-components 1 linux-amd64/helm \
  && curl -sLo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl \
  && chmod 755 /usr/local/bin/kubectl \
  && helm init --client-only \
  && helm repo add jetstack https://charts.jetstack.io \
  && helm plugin install https://github.com/bacongobbler/helm-whatup \
  && gem install cloud-platform-repository-checker -v 1.0.4 --no-rdoc --no-doc

WORKDIR /app
COPY *.sh *.rb /app/
