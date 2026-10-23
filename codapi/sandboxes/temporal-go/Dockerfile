# ABOUTME: Codapi box for the Go IP geolocation tutorial: a Temporal dev server, Worker, and client per run.
# ABOUTME: Build from the repo root after the shared base image; caches are primed with examples/go.
ARG BASE_IMAGE=temporal-codapi/base:latest
FROM ${BASE_IMAGE} AS base

FROM golang:1.26-bookworm

COPY --from=base /usr/local/bin/temporal /usr/local/bin/temporal-serve /usr/local/bin/temporal-run /usr/local/bin/

RUN apt-get update \
    && apt-get install -y --no-install-recommends jq \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home sandbox

ENV CGO_ENABLED=0 \
    GOTOOLCHAIN=local \
    GOFLAGS=-mod=readonly \
    GOCACHE=/home/sandbox/.cache/go-build \
    TEMPORAL_APP_DIR=/app

# Download the SDK, then build and test the tutorial code once as the sandbox user to fill the build cache.
WORKDIR /app
COPY codapi/sandboxes/temporal-go/go.mod codapi/sandboxes/temporal-go/go.sum ./
RUN go mod download && chown -R sandbox:sandbox /app
USER sandbox
COPY --chown=sandbox:sandbox examples/go/ /tmp/prewarm/
RUN cp /app/go.mod /app/go.sum /tmp/prewarm/ \
    && cd /tmp/prewarm \
    && go build -o /dev/null ./... \
    && go test -count=1 ./... \
    && rm -rf /tmp/prewarm

# At run time the only route out is the gateway's allowlisting proxy; module lookups never leave the image.
ENV GOPROXY=off \
    HTTP_PROXY=http://gateway:3128 \
    HTTPS_PROXY=http://gateway:3128 \
    http_proxy=http://gateway:3128 \
    https_proxy=http://gateway:3128 \
    NO_PROXY=localhost,127.0.0.1 \
    no_proxy=localhost,127.0.0.1
