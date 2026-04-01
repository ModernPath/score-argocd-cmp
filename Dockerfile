FROM golang:1.25-alpine AS builder
ARG SCORE_K8S_VERSION=0.10.3

RUN apk add --no-cache git && \
    go install github.com/score-spec/score-k8s/cmd/score-k8s@${SCORE_K8S_VERSION}

FROM alpine:3.23
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/bin/score-k8s /usr/local/bin/score-k8s
COPY provisioners/ /opt/provisioners/
USER 999
