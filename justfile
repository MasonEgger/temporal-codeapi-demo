# ABOUTME: Task runner recipes; `just check` is the local gate (no pre-commit).
# ABOUTME: Also builds the Temporal sandbox images and runs the Codapi server and demo page.

codapi_version := "0.14.2"
sdks := "python go java typescript"

default:
    @just --list

check:
    uv run ruff check .
    uv run ruff format --check .
    uv run mypy .
    uv run pytest

fmt:
    uv run ruff format .
    uv run ruff check --fix .

test:
    uv run pytest

# Build the shared Temporal CLI base image, then one sandbox image per SDK.
build:
    docker build --tag temporal-codapi/base:latest codapi/sandboxes/shared
    for sdk in {{sdks}}; do \
        docker build --tag "temporal-codapi/${sdk}:latest" "codapi/sandboxes/temporal-${sdk}"; \
    done

# Download the Codapi server binary for this machine into bin/.
install-codapi:
    #!/usr/bin/env bash
    set -euo pipefail
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    arch="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')"
    mkdir -p bin
    curl -sSfL "https://github.com/nalgeon/codapi/releases/download/v{{codapi_version}}/codapi_{{codapi_version}}_${os}_${arch}.tar.gz" \
        | tar -xz -C bin codapi
    echo "Installed bin/codapi {{codapi_version}}"

# Run the Codapi server on :1313. Temp files live under the repo because Colima only mounts $HOME.
serve:
    mkdir -p .tmp/codapi
    cd codapi && TMPDIR="{{justfile_directory()}}/.tmp/codapi" ../bin/codapi

# Serve the demo page at http://localhost:8080/web/
web:
    uv run python -m http.server 8080 --bind 127.0.0.1
