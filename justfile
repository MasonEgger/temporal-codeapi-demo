# ABOUTME: Task runner recipes; `just check` is the local gate (no pre-commit).
# ABOUTME: Also builds the sandbox and gateway images and runs the gateway and Codapi server.

codapi_version := "0.14.2"
sdks := "python go java typescript"
sandbox_network := "temporal-codapi-sandbox"
gateway_container := "temporal-codapi-gateway"

default:
    @just --list

check:
    uv run ruff check .
    uv run ruff format --check .
    uv run mypy .
    cd gateway && go vet ./... && go test ./...
    uv run pytest

fmt:
    uv run ruff format .
    uv run ruff check --fix .
    cd gateway && gofmt -w .

test:
    uv run pytest

# Build the gateway, the shared Temporal CLI base image, then one sandbox image per SDK.
build:
    docker build --tag temporal-codapi/gateway:latest gateway
    docker build --tag temporal-codapi/base:latest --file codapi/sandboxes/shared/Dockerfile .
    for sdk in {{sdks}}; do \
        docker build --tag "temporal-codapi/${sdk}:latest" --file "codapi/sandboxes/temporal-${sdk}/Dockerfile" .; \
    done

# Create the internal sandbox network and start the gateway; the site is at http://localhost:3129/
up:
    #!/usr/bin/env bash
    set -euo pipefail
    # An internal network has no route out; the gateway is the only member that also has internet access.
    docker network inspect {{sandbox_network}} >/dev/null 2>&1 \
        || docker network create --internal {{sandbox_network}}
    docker run --detach --rm --name {{gateway_container}} \
        --publish 127.0.0.1:3129:3129 \
        --volume "{{justfile_directory()}}/web:/srv/site:ro" \
        --volume "{{justfile_directory()}}/examples:/srv/examples:ro" \
        temporal-codapi/gateway:latest
    docker network connect --alias gateway {{sandbox_network}} {{gateway_container}}
    echo "Gateway up: open http://localhost:3129/"

# Stop the gateway.
down:
    docker stop {{gateway_container}}

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
