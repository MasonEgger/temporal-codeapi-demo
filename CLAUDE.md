# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A browser playground edition of the learn.temporal.io tutorial "Build a Temporal Application from scratch" (the ip-geolocation app) for Python, Go, Java, and TypeScript.
Each run gets a throwaway Temporal dev server, Worker, and client in a Codapi sandbox container; a Go gateway handles the egress allowlist, a per-learner network switch, and per-learner Web UI routing.

## Commands

```bash
just build             # gateway image, shared base image, then temporal-codapi/{python,go,java,typescript}
just install-codapi    # Codapi server binary into bin/ (gitignored)
just up                # internal Docker network + gateway on :3129 (serves the site)
just serve             # Codapi on :1313 (foreground); restart it after editing any codapi/*.json
just check             # ruff, ruff format --check, mypy --strict, gateway go vet/test, pytest smoke tests
just infra-apply       # Terraform: droplet, firewall, DNS record, Ansible inventory (needs DIGITALOCEAN_TOKEN)
just deploy            # Ansible: configure the droplet; `just deploy -e playground_git_ref=<branch>` deploys a branch
just smoke <url>       # the smoke tests against a deployed playground
just infra-destroy     # tear the droplet down (it bills hourly)
```

The gateway needs Go 1.25 or newer; if the local toolchain is older and cannot download one, run the gateway checks in Docker: `docker run --rm -v "$PWD/gateway":/src -w /src -e GOFLAGS=-buildvcs=false golang:1.26-bookworm sh -c 'gofmt -l . && go vet ./... && go test ./...'`.

Single smoke test (needs `just up` and `just serve` running):

```bash
uv run pytest tests/test_sandboxes.py -k "TestTutorials and go"
```

The smoke tests post to `PLAYGROUND_URL` (default `http://localhost:3129`), always through the gateway.

Gateway unit tests alone: `cd gateway && go test ./...`.
After changing a Dockerfile, `codapi/sandboxes/shared/temporal-*`, or anything a Dockerfile copies (including `examples/` used for pre-warming), rebuild with `just build`.
After rebuilding the gateway, `docker stop temporal-codapi-gateway && just up`.

## Architecture

Request path: page (`web/<language>.html` + `web/tutorial.js`) → gateway `POST /v1/exec` (same origin) with every example file plus `session.id` and `name.txt` → forwarded to Codapi → `docker run --detach` (`temporal-serve`) then `docker exec` (`temporal-run`) in the same container on the internal `temporal-codapi-sandbox` network.

- `temporal-serve` starts `temporal server start-dev` with the Web UI at `/ui/<session>`, registers the container's IP with the gateway, and lingers up to 10 minutes or until the gateway reports a newer container for that session.
- `temporal-run` waits for the server, copies `/sandbox` into `$TEMPORAL_APP_DIR`, runs an optional `--prepare` step, starts `--worker` in the background, runs the client (appending the learner's name with `--pass-name`), then prints the Worker log and, with `--history`, each Workflow's Event History. `--no-server` is for the `test` command.
- `gateway/` has two listeners: `:3128` on the sandbox network (forward proxy, `/register/{session}`, `/current/{session}`) and `:3129` for browsers (site files with `Cache-Control: no-cache`, `/examples/`, `/api/sessions/{id}`, `/ui/{id}/`, and `/v1/` forwarded to `-codapi-addr`). `-run-rate`/`-run-burst` cap `/v1/exec` per client address (`-trust-proxy` keys on `X-Forwarded-For` behind Caddy). Sessions, network state, blocked counts, and rate buckets live in memory, so restarting the gateway resets them.
- `deploy/` is the hosted setup: `terraform/` (droplet, firewall, DNS, writes `ansible/inventory/hosts.yml`), `ansible/` (roles base, docker, playground, codapi, caddy), `compose.yaml` (the gateway on the droplet). `spec.md` at the repo root is the deployment spec. Codapi runs as a systemd service on the droplet with `WorkingDirectory` set to the checkout's `codapi/` directory; `sandbox-isolation.service` installs the iptables rules that keep sandboxes from reaching each other or the host.
- Sandboxes reach the internet only through `HTTP(S)_PROXY=http://gateway:3128` (Java uses the same settings via `java.args`, Node via `NODE_USE_ENV_PROXY=1`). `NO_PROXY` keeps the local dev server direct.

## Tutorial content

- `web/tutorials/source/*.md` are untouched copies of the live tutorials from `temporalio/temporal-learning`; ruff excludes them. Never edit them.
- `web/tutorials/<language>.md` are copy-and-edit versions: Go and TypeScript from their own originals with playground-only edits; Python and Java from the Go original with language-specific edits. Keep the original wording wherever it still applies.
- `<!--SNIPSTART id-->…<!--SNIPEND-->` blocks are replaced at render time by the region between `@@@SNIPSTART id` and `@@@SNIPEND` comments in `examples/<language>/`. The code shown and the code that runs are the same file.
- Page markers: `<div class="run-block" data-command="run|history|test" …>` (exercise buttons add `data-variant-file/find/replace`), `<div class="network-switch">`, `<div class="webui">`, `<div class="quiz">`, and `:::setup Title` blocks for local-only setup steps.
- Each page's `data-files` attribute must list every example file the run needs.

## Gotchas

- On macOS with Colima, Docker only mounts paths under `$HOME`; `just serve` points Codapi's `TMPDIR` at `.tmp/codapi`. If Colima stops, `colima start`, `docker context use colima`, `just up`.
- Python programs run through `codapi/sandboxes/temporal-python/run_program.py`, which exits with `os._exit` to avoid temporalio/sdk-python#300.
- Java Activities must rethrow with `Activity.wrap(e)`; extra exception layers push the failure past the server's size limit and the Web UI shows "Failure exceeds size limit."
- codapi-js is not used: its 10-second fetch timeout breaks outage runs, so `tutorial.js` posts to `/v1/exec` itself.
- Code files start with a two-line `ABOUTME:` header; `tutorial.js` never shows it because snippets only include marked regions.
- Commits must be GPG-signed (`git commit -S`); if gpg-agent's cache expired, signing fails without a terminal until the passphrase is re-entered.
