ARG GO_VERSION=1.18
ARG ALPINE_VERSION=3.16

FROM golang:${GO_VERSION}-alpine as builder
LABEL stage=gobuilder
RUN apk add --no-cache make gcc musl-dev git

WORKDIR $GOPATH/src/github.com/kaonone/eth-rpc-gate
COPY go.mod go.sum $GOPATH/src/github.com/kaonone/eth-rpc-gate/

# Cache go modules
RUN go mod download -x

ARG GIT_SHA
ENV CGO_ENABLED=0
ENV GOOS linux

COPY ./ $GOPATH/src/github.com/kaonone/eth-rpc-gate

ENV GIT_SHA=$GIT_SH

RUN go build \
        -ldflags \
            "-s -w -X 'github.com/kaonone/eth-rpc-gate/pkg/params.GitSha=`./sha.sh`'" \
        -o $GOPATH/bin $GOPATH/src/github.com/kaonone/eth-rpc-gate/... && \
    rm -fr $GOPATH/src/github.com/kaonone/eth-rpc-gate/.git

# Final stage
FROM alpine:${ALPINE_VERSION} as base

ARG GATE_PORT
ENV GATE_PORT=${GATE_PORT}

ARG GATE_BIND
ENV GATE_BIND=${GATE_BIND}

ARG KAON_RPC
ENV KAON_RPC=${KAON_RPC}

WORKDIR /app

COPY --from=builder /go/bin/eth-rpc-gate /app/eth-rpc-gate
COPY --from=builder /go/src/github.com/kaonone/eth-rpc-gate/docker/standalone/myaccounts.txt /app/myaccounts.txt

# Makefile supports generating ssl files from docker
# Install curl and openssl
RUN apk add --no-cache curl openssl

ENTRYPOINT /app/eth-rpc-gate \
                --kaon-rpc $KAON_RPC \
                --kaon-network "auto" \
                --bind $GATE_BIND \
                --port $GATE_PORT \
                --accounts /app/myaccounts.txt \
                --log-file /app/logs/gateLogs.txt \
                --https-cert /certs/fullchain.pem \
                --https-key /certs/privkey.pem \
                --dbstring ""