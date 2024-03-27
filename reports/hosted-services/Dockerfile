FROM golang:1.22.1
WORKDIR /app

ENV XDG_CACHE_HOME=/tmp/.cache

RUN useradd nonroot --uid 1001 -U -M

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build .

RUN chown -R nonroot:nonroot /app

USER 1001
