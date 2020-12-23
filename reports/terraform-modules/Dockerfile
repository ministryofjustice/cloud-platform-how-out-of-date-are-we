FROM ruby:2.7-alpine

RUN apk update && apk add git curl

WORKDIR /app

COPY bin/ /app/bin

RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

RUN chown 1000:1000 /app
USER 1000
