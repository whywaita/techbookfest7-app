# syntax = docker/dockerfile:experimental
FROM golang:1.12.5-alpine AS builder
WORKDIR /go/src/github.com/whywaita/techbookfest7-app/cluster-generator/apps/kustomization
COPY ./ /go/src/github.com/whywaita/techbookfest7-app/cluster-generator/apps/kustomization
ENV GO111MODULE=on
RUN apk update \
    && apk add --no-cache git openssh libc-dev gcc \
    && cd /go/src/github.com/whywaita/techbookfest7-app/cluster-generator/apps/kustomization \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o kustomization main.go 

FROM debian AS second_builder
RUN apt update \
    && apt install -y ca-certificates curl apt-transport-https gnupg wget
RUN curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
RUN echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | tee -a /etc/apt/sources.list.d/kubernetes.list \
    && apt update -y \
    && apt-get install -y kubectl \
    && wget -O kustomize https://github.com/kubernetes-sigs/kustomize/releases/download/v3.1.0/kustomize_3.1.0_linux_amd64 \
    && chmod +x kustomize \
    && cp kustomize /usr/local/bin

FROM debian
COPY --from=builder /go/src/github.com/whywaita/techbookfest7-app/cluster-generator/apps/kustomization/kustomization /usr/local/bin/kustomization
COPY --from=second_builder /usr/bin/kubectl /usr/local/bin/kubectl
COPY --from=second_builder /usr/local/bin/kustomize /usr/local/bin/kustomize
RUN apt update \
    && apt install -y ca-certificates lsb-release curl gnupg
RUN export CLOUD_SDK_REPO="cloud-sdk-$(lsb_release -c -s)" && \
    echo "deb http://packages.cloud.google.com/apt $CLOUD_SDK_REPO main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - && \
    apt-get update -y && apt-get install google-cloud-sdk -y
CMD ["/usr/local/bin/kustomization"]