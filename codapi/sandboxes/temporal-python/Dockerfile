# ABOUTME: Codapi box for running Temporal Python SDK programs against a per-run dev server.
# ABOUTME: Build after the shared base image; see the justfile's build recipe.
ARG BASE_IMAGE=temporal-codapi/base:latest
FROM ${BASE_IMAGE} AS base

FROM python:3.13-slim-bookworm

COPY --from=ghcr.io/astral-sh/uv:0.12.19 /uv /usr/local/bin/uv
COPY --from=base /usr/local/bin/temporal /usr/local/bin/temporal-run /usr/local/bin/

RUN apt-get update \
    && apt-get install -y --no-install-recommends jq \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home sandbox

# Preinstall the SDK into a virtual environment the learner's code runs in.
WORKDIR /app
COPY pyproject.toml uv.lock ./
RUN UV_COMPILE_BYTECODE=1 uv sync --frozen --no-dev \
    && chown -R sandbox:sandbox /app

# Launcher that exits without interpreter finalization (temporalio/sdk-python#300).
COPY run_program.py /opt/temporal/run_program.py

ENV PATH=/app/.venv/bin:$PATH \
    TEMPORAL_APP_DIR=/app \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

USER sandbox
