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
    # The gateway forwards /v1/ to Codapi on the host, so the browser talks to one origin.
    docker run --detach --rm --name {{gateway_container}} \
        --publish 127.0.0.1:3129:3129 \
        --add-host host.docker.internal:host-gateway \
        --volume "{{justfile_directory()}}/web:/srv/site:ro" \
        --volume "{{justfile_directory()}}/examples:/srv/examples:ro" \
        temporal-codapi/gateway:latest -codapi-addr host.docker.internal:1313
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

# ---- Hosting on DigitalOcean (see deploy/README.md) --------------------------------------------

terraform := "terraform -chdir=deploy/terraform"
inventory := "deploy/ansible/inventory/hosts.yml"

# Fail early when the DigitalOcean token is missing from the environment.
_need-token:
    #!/usr/bin/env bash
    if [[ -z "${DIGITALOCEAN_TOKEN:-}" ]]; then
        echo "Set DIGITALOCEAN_TOKEN first, for example from doctl's config:" >&2
        echo "  export DIGITALOCEAN_TOKEN=\$(awk '/^access-token:/ {print \$2}' ~/.config/doctl/config.yaml)" >&2
        exit 1
    fi

# Download the Terraform providers.
infra-init:
    {{terraform}} init

# Show what Terraform would create or change.
infra-plan: _need-token
    {{terraform}} plan

# Create the droplet, firewall, and DNS record, and write the Ansible inventory.
infra-apply: _need-token
    {{terraform}} apply

# Destroy everything Terraform created. The droplet bills hourly until this runs.
infra-destroy: _need-token
    {{terraform}} destroy

# Configure the droplet (or bring it up to date) with Ansible. Extra arguments pass through.
deploy *args:
    ansible-playbook -i {{inventory}} deploy/ansible/site.yml {{args}}

# Open a shell on the droplet.
deploy-ssh:
    ssh "root@$({{terraform}} output -raw ipv4)"

# Run the smoke tests against a deployed playground, for example `just smoke https://temporal.egger.codes`.
smoke url:
    PLAYGROUND_URL="{{url}}" uv run pytest
