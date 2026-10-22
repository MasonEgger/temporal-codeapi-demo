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
    gitleaks dir --no-banner --redact .
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
# The DigitalOcean token lives sops-encrypted in secrets/terraform.sops.env and Terraform state in
# deploy/terraform/state.sops.json (see .sops.yaml); `_tf` decrypts both around each terraform run.

tf_dir := "deploy/terraform"
tf_secrets := "secrets/terraform.sops.env"
inventory := "deploy/ansible/inventory/hosts.yml"

# Download the Terraform providers.
infra-init: (_tf "init -input=false")

# Show what Terraform would create or change. Extra arguments pass through, e.g. `-var git_ref=my-branch`.
infra-plan *args: (_tf "plan " + args)

# Create the droplet, firewall, and DNS record, and write the Ansible inventory. Commit state.sops.json afterward.
infra-apply *args: (_tf "apply -auto-approve " + args)

# Destroy everything Terraform created. The droplet bills hourly until this runs. Commit state.sops.json afterward.
infra-destroy *args: (_tf "destroy -auto-approve " + args)

# Read a Terraform output, e.g. `just infra-output -raw ipv4`.
infra-output *args: (_tf "output " + args)

# Configure the droplet (or bring it up to date) with Ansible. Extra arguments pass through.
deploy *args:
    ansible-playbook -i {{inventory}} deploy/ansible/site.yml {{args}}

# Open a shell on the droplet.
deploy-ssh:
    ssh "root@$(just infra-output -raw ipv4)"

# Run the smoke tests against a deployed playground, for example `just smoke https://temporal.egger.codes`.
smoke url:
    PLAYGROUND_URL="{{url}}" uv run pytest

# Create or edit a sops-encrypted file, for example `just sops secrets/terraform.sops.env`.
sops file:
    sops {{file}}

# Re-encrypt every secrets file after changing the recipient list in .sops.yaml.
rekey:
    #!/usr/bin/env bash
    set -euo pipefail
    shopt -s nullglob
    for f in secrets/*.sops.env secrets/*.sops.yaml {{tf_dir}}/*.sops.json; do
        echo "rekey: $f"
        sops updatekeys -y "$f"
    done

# Run a terraform command with the token in its environment and the state decrypted from state.sops.json.
# The state is re-encrypted in place only when it changed (so plans do not churn git), including after a failed apply.
_tf cmd:
    #!/usr/bin/env bash
    set -euo pipefail
    enc="{{tf_dir}}/state.sops.json"
    plain="{{tf_dir}}/terraform.tfstate"
    pre="${plain}.pre"
    if [[ -f "${enc}" ]]; then
        sops --decrypt --output-type binary "${enc}" > "${plain}"
        cp -f "${plain}" "${pre}"
    else
        : > "${pre}"
    fi
    rc=0
    sops exec-env {{tf_secrets}} "terraform -chdir={{tf_dir}} {{cmd}}" || rc=$?
    if [[ -f "${plain}" ]] && ! cmp -s "${plain}" "${pre}"; then
        cp -f "${plain}" "${enc}"
        sops --encrypt --input-type binary --output-type json --in-place "${enc}"
        echo ">> Terraform state changed; re-encrypted ${enc}. Commit it." >&2
    fi
    rm -f "${pre}"
    exit "${rc}"
