# ABOUTME: Codapi box for running Temporal TypeScript SDK programs against a per-run dev server.
# ABOUTME: Node runs the learner's .ts files directly via built-in type stripping; no compile step.
ARG BASE_IMAGE=temporal-codapi/base:latest
FROM ${BASE_IMAGE} AS base

FROM node:24-bookworm-slim

COPY --from=base /usr/local/bin/temporal /usr/local/bin/temporal-run /usr/local/bin/

RUN apt-get update \
    && apt-get install -y --no-install-recommends jq \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home sandbox

# Preinstall the SDK where the learner's files will be copied.
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --omit=dev --no-audit --no-fund \
    && chown -R sandbox:sandbox /app

ENV TEMPORAL_APP_DIR=/app

USER sandbox
