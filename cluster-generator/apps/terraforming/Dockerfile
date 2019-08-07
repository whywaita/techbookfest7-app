# syntax = docker/dockerfile:experimental
FROM golang:1.12.5-alpine AS builder
WORKDIR /go/src/github.com/whywaita/techbookfest7-app/cluster-generator/apps/terraforming
COPY ./ /go/src/github.com/whywaita/techbookfest7-app/cluster-generator/apps/terraforming
ENV GO111MODULE=on
RUN apk update \
    && apk add --no-cache git openssh libc-dev \
    && cd /go/src/github.com/whywaita/techbookfest7-app/cluster-generator/apps/terraforming \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o terraforming main.go

FROM debian AS second_builder
RUN apt update \
    && apt install -y ca-certificates wget unzip \
    && wget https://releases.hashicorp.com/terraform/0.12.7/terraform_0.12.7_linux_amd64.zip \
    && unzip terraform_0.12.7_linux_amd64.zip\
    && cp terraform /usr/local/bin
#CMD ["/usr/local/bin/terraforming"]

FROM debian
COPY --from=builder /go/src/github.com/whywaita/techbookfest7-app/cluster-generator/apps/terraforming/terraforming /usr/local/bin/terraforming
COPY --from=second_builder /usr/local/bin/terraform /usr/local/bin/terraform
RUN apt update \
    && apt install -y ca-certificates
CMD ["/usr/local/bin/terraforming"]
