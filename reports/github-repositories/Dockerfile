FROM ruby:2.7-alpine

ENV \
  REPO_CHECKER_VERSION=1.4.1

RUN apk update && apk add --no-cache curl
RUN gem install json_pure \
  && gem install cloud-platform-repository-checker -v ${REPO_CHECKER_VERSION}

WORKDIR /app

COPY bin/ /app/bin

RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

RUN chown 1000:1000 /app
USER 1000
