# ABOUTME: Codapi box for running Temporal Go SDK programs against a per-run dev server.
# ABOUTME: Dependencies and the build cache are primed at build time so `go run` only compiles learner code.
ARG BASE_IMAGE=temporal-codapi/base:latest
FROM ${BASE_IMAGE} AS base

FROM golang:1.26-bookworm

COPY --from=base /usr/local/bin/temporal /usr/local/bin/temporal-run /usr/local/bin/

RUN apt-get update \
    && apt-get install -y --no-install-recommends jq \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home sandbox

ENV CGO_ENABLED=0 \
    GOTOOLCHAIN=local \
    GOFLAGS=-mod=readonly \
    GOCACHE=/home/sandbox/.cache/go-build \
    TEMPORAL_APP_DIR=/app

# Download the SDK, then compile it once as the sandbox user to fill the build cache.
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && chown -R sandbox:sandbox /app
USER sandbox
COPY --chown=sandbox:sandbox warmup/main.go ./main.go
RUN go build -o /dev/null . && rm main.go

# Learner runs have no network, so module lookups must never leave the image.
ENV GOPROXY=off
