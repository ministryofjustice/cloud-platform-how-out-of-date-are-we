FROM golang:1.23.6-alpine3.21

ARG COMMAND

ENV binary $COMMAND

WORKDIR /go/bin

COPY ${COMMAND} /go/bin

RUN addgroup -g 1000 -S appgroup \
  && adduser -u 1000 -S appuser -G appgroup

RUN chown -R appuser:appgroup /go/bin

USER 1000