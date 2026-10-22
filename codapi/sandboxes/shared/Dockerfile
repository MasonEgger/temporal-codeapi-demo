# ABOUTME: Base layer holding the Temporal CLI and the temporal-serve and temporal-run harness scripts.
# ABOUTME: Each language sandbox copies /usr/local/bin from this image instead of downloading the CLI itself.
FROM debian:bookworm-slim

ARG TEMPORAL_CLI_VERSION=v1.9.1
ARG TARGETARCH

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl \
    && curl -sSfL "https://temporal.download/cli/archive/${TEMPORAL_CLI_VERSION}?platform=linux&arch=${TARGETARCH}" \
        | tar -xz -C /usr/local/bin temporal \
    && rm -rf /var/lib/apt/lists/*

COPY codapi/sandboxes/shared/temporal-serve codapi/sandboxes/shared/temporal-run /usr/local/bin/
RUN chmod 0755 /usr/local/bin/temporal /usr/local/bin/temporal-serve /usr/local/bin/temporal-run
