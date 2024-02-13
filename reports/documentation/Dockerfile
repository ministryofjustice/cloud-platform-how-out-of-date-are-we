FROM ruby:3.2.2-alpine3.18

RUN apk update && apk add --no-cache build-base curl
RUN gem install nokogiri json_pure

WORKDIR /app

COPY bin/ /app/bin

RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

RUN chown 1000:1000 /app
USER 1000
