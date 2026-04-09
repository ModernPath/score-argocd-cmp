FROM golang:1.26-alpine AS builder
ARG SCORE_K8S_VERSION=0.10.3

RUN apk add --no-cache git && \
    go install github.com/score-spec/score-k8s/cmd/score-k8s@${SCORE_K8S_VERSION} && \
    go install github.com/GoogleCloudPlatform/docker-credential-gcr/v2@latest

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -o /go/bin/score-argocd-cmp ./cmd/score-argocd-cmp

FROM alpine:3.23
RUN apk add --no-cache ca-certificates bash
COPY --from=builder /go/bin/docker-credential-gcr /usr/local/bin/docker-credential-gcr
COPY --from=builder /go/bin/score-k8s /usr/local/bin/score-k8s
COPY --from=builder /go/bin/score-argocd-cmp /usr/local/bin/score-argocd-cmp
COPY plugin.yaml /home/argocd/cmp-server/config/plugin.yaml

# Patch templates baked into the image and applied unconditionally by `score-argocd-cmp init`.
# See internal/initialize/initialize.go for the wiring and patches/*.tpl for the rationale.
COPY patches/ /usr/local/share/score-argocd-cmp/patches/

RUN mkdir -p /.docker && chown 999:999 /.docker && chmod 700 /.docker
COPY --chown=999:999 --chmod=600 docker-config.json /.docker/config.json
RUN mkdir /work && chown 999:999 /work
USER 999
