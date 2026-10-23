# ABOUTME: Codapi box for the Python IP geolocation tutorial: a Temporal dev server, Worker, and client per run.
# ABOUTME: Build from the repo root after the shared base image; see the justfile's build recipe.
ARG BASE_IMAGE=temporal-codapi/base:latest
FROM ${BASE_IMAGE} AS base

FROM python:3.13-slim-bookworm

COPY --from=ghcr.io/astral-sh/uv:0.12.19 /uv /usr/local/bin/uv
COPY --from=base /usr/local/bin/temporal /usr/local/bin/temporal-serve /usr/local/bin/temporal-run /usr/local/bin/

RUN apt-get update \
    && apt-get install -y --no-install-recommends curl jq \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home sandbox

# Preinstall the SDK, requests, and pytest into the virtual environment learner code runs in.
WORKDIR /app
COPY codapi/sandboxes/temporal-python/pyproject.toml codapi/sandboxes/temporal-python/uv.lock ./
RUN UV_COMPILE_BYTECODE=1 uv sync --frozen --no-dev \
    && chown -R sandbox:sandbox /app

# Launcher that exits without interpreter finalization (temporalio/sdk-python#300).
COPY codapi/sandboxes/temporal-python/run_program.py /opt/temporal/run_program.py

ENV PATH=/app/.venv/bin:$PATH \
    TEMPORAL_APP_DIR=/app \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

USER sandbox

# Run the example tests once so the SDK downloads its time-skipping test server into /tmp now,
# while the build still has internet access.
COPY --chown=sandbox:sandbox examples/python/ /tmp/prewarm/
RUN cd /tmp/prewarm && python -m pytest -q -p no:cacheprovider tests && rm -rf /tmp/prewarm

# At run time the only route out is the gateway's allowlisting proxy; the dev server stays on loopback.
ENV HTTP_PROXY=http://gateway:3128 \
    HTTPS_PROXY=http://gateway:3128 \
    http_proxy=http://gateway:3128 \
    https_proxy=http://gateway:3128 \
    NO_PROXY=localhost,127.0.0.1 \
    no_proxy=localhost,127.0.0.1
