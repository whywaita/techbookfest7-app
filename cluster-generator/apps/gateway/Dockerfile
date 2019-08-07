# syntax = docker/dockerfile:experimental
FROM golang:1.12.5-alpine AS builder
WORKDIR /go/src/github.com/whywaita/techbookfest7-app/cluster-generator/apps/gateway
COPY ./ /go/src/github.com/whywaita/techbookfest7-app/cluster-generator/apps/gateway
ENV GO111MODULE=on
RUN apk update \
    && apk add --no-cache git openssh libc-dev \
    && cd /go/src/github.com/whywaita/techbookfest7-app/cluster-generator/apps/gateway \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gateway main.go

FROM alpine
COPY --from=builder /go/src/github.com/whywaita/techbookfest7-app/cluster-generator/apps/gateway/gateway /usr/local/bin/gateway
RUN apk update \
    && apk add ca-certificates --no-cache
CMD ["/usr/local/bin/gateway"]