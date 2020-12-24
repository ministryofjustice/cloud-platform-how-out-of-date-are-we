FROM python:alpine3.12

ENV \
  KUBECTL_VERSION=1.17.12

RUN pip install awscli \
  && apk add curl ruby
RUN gem install json_pure

# kubectl
RUN curl -sLo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl \
  && chmod 755 /usr/local/bin/kubectl


RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

COPY bin ./bin

RUN chown 1000:1000 /app
USER 1000
