FROM hashicorp/terraform:0.12.17 AS terraform

FROM ruby:2.5-alpine

ENV \
  KUBECTL_VERSION=1.16.3

# Install pre-requisites for building unf_ext gem
RUN apk --update add --virtual build_deps \
    build-base ruby-dev libc-dev linux-headers \
    git \
    curl

# Install kubectl
RUN curl -sLo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl

# Ensure everything is executable
RUN chmod +x /usr/local/bin/*

WORKDIR /app

COPY Gemfile Gemfile.lock ./

RUN bundle install --without development test

COPY lib ./lib
COPY bin ./bin

COPY --from=terraform bin/terraform /app/bin/terraform

RUN addgroup -g 1000 -S appgroup \
  && adduser -u 1000 -S appuser -G appgroup

RUN chown -R appuser:appgroup /app

USER 1000
