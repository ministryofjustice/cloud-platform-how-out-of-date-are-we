FROM golang:1.23.5 AS namespace_usage_builder

ENV CGO_ENABLED=0 \
  GOOS=linux

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build .

FROM alpine:3.11.0

WORKDIR /app

COPY --from=namespace_usage_builder /app/namespace-usage ./

RUN addgroup -g 1000 -S appgroup \
  && adduser -u 1000 -S appuser -G appgroup

RUN chown -R appuser:appgroup /app

USER 1000
