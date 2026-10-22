# ABOUTME: Codapi box for the TypeScript IP geolocation tutorial: a Temporal dev server, Worker, and client per run.
# ABOUTME: tsx runs the learner's .ts files directly; Mocha runs the tests, with the test server downloaded at build time.
ARG BASE_IMAGE=temporal-codapi/base:latest
FROM ${BASE_IMAGE} AS base

FROM node:24-bookworm-slim

COPY --from=base /usr/local/bin/temporal /usr/local/bin/temporal-serve /usr/local/bin/temporal-run /usr/local/bin/

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl jq \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home sandbox

# Preinstall the SDK, its testing package, and tsx where the learner's files will be copied.
WORKDIR /app
COPY codapi/sandboxes/temporal-typescript/package.json codapi/sandboxes/temporal-typescript/package-lock.json codapi/sandboxes/temporal-typescript/tsconfig.json ./
RUN npm ci --omit=dev --no-audit --no-fund \
    && chown -R sandbox:sandbox /app

ENV TEMPORAL_APP_DIR=/app

USER sandbox

# Run the tutorial tests once so the SDK downloads the dev server that
# TestWorkflowEnvironment.createLocal() uses, while the build still has internet access.
COPY --chown=sandbox:sandbox examples/typescript/src/ /app/src/
RUN node_modules/.bin/mocha --exit --timeout 20000 --require tsx/cjs 'src/mocha/*.test.ts' \
    && rm -rf /app/src

# At run time the only route out is the gateway's allowlisting proxy; the dev server stays on loopback.
ENV NODE_USE_ENV_PROXY=1 \
    HTTP_PROXY=http://gateway:3128 \
    HTTPS_PROXY=http://gateway:3128 \
    http_proxy=http://gateway:3128 \
    https_proxy=http://gateway:3128 \
    NO_PROXY=localhost,127.0.0.1 \
    no_proxy=localhost,127.0.0.1
