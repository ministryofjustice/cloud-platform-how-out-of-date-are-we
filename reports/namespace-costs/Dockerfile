FROM alpine:3.12

RUN apk --no-cache add git curl ruby ruby-json

RUN gem install aws-sdk-costexplorer

RUN addgroup -g 1000 -S appgroup \
  && adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

RUN mkdir lib
COPY post-namespace-costs.rb .
COPY lib/* lib/

RUN chown -R appuser:appgroup /app

USER 1000
