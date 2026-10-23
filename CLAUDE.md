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
just infra-init        # Terraform providers (once)
just infra-apply       # Terraform: droplet, firewall, DNS record, Ansible inventory; then commit deploy/terraform/state.sops.json
just deploy            # Ansible: configure the droplet; `just deploy -e playground_git_ref=<branch>` deploys a branch
just deploy-ssh        # root shell on the droplet
just smoke <url>       # the smoke tests against a deployed playground
just infra-destroy     # tear the droplet down (it bills hourly); then commit state.sops.json
just sops <file>       # edit a sops-encrypted secret (secrets/terraform.sops.env holds TF_VAR_do_token)
just rekey             # re-encrypt secrets and state after changing .sops.yaml recipients
```

Secrets follow the PyTexas/Labradillo convention: sops with age, recipients in `.sops.yaml`, `secrets/*.sops.env` consumed with `sops exec-env`, Terraform state encrypted whole-file in `deploy/terraform/state.sops.json` by the justfile's `_tf` wrapper. Never run `terraform` directly in `deploy/terraform`; go through `just infra-*` so the state is decrypted and re-encrypted. Nothing secret reaches the droplet. `.gitleaks.toml` allowlists the encrypted files.

Deploy-side checks: `terraform -chdir=deploy/terraform fmt -check && terraform -chdir=deploy/terraform validate`, and `ansible-lint --offline` from `deploy/ansible/` (passes the production profile; registered variables in a role must carry the role's name as a prefix).
Ansible clones the repository from GitHub, so a change must be pushed before `just deploy` can pick it up.

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
- `deploy/` is the hosted setup: `terraform/` (droplet, firewall, DNS, writes `ansible/inventory/hosts.yml`), `ansible/` (roles base, docker, playground, codapi, caddy), `compose.yaml` (the gateway on the droplet). `spec.md` at the repo root is the deployment spec. Codapi runs as a systemd service on the droplet with `WorkingDirectory` set to the checkout's `codapi/` directory; `sandbox-isolation.service` installs nftables rules (bridge family, since Docker 28+ does not load `br_netfilter`) that keep sandboxes from reaching each other or the host.
- Sandboxes reach the internet only through `HTTP(S)_PROXY=http://gateway:3128` (Java uses the same settings via `java.args`, Node via `NODE_USE_ENV_PROXY=1`). `NO_PROXY` keeps the local dev server direct.

## Tutorial content

- `web/tutorials/source/*.md` are untouched copies of the live tutorials from `temporalio/temporal-learning`; ruff excludes them. Never edit them.
- `web/tutorials/<language>.md` are copy-and-edit versions: Go and TypeScript from their own originals with playground-only edits; Python and Java from the Go original with language-specific edits. Keep the original wording wherever it still applies.
- `<!--SNIPSTART id-->…<!--SNIPEND-->` blocks are replaced at render time by the region between `@@@SNIPSTART id` and `@@@SNIPEND` comments in `examples/<language>/`. The code shown and the code that runs are the same file.
- Page markers: `<div class="run-block" data-command="run|history|test" …>` (exercise buttons add `data-variant-file/find/replace`; `data-omit-files` leaves listed files out of the request, so the first "Run the Workflow test" button runs before the Activity tests exist), `<div class="network-switch">`, `<div class="webui">`, `<div class="quiz">`, and `:::setup Title` blocks for local-only setup steps.
- Anything a run would show live is backup and folds into a closed `<details class="sample-output">`: every ```output fence and every screenshot (an image inside a mixed paragraph is lifted out into its own accordion). No Markdown change is needed for that.
- To check rendered pages without a local browser: `just up`, then `docker run --rm --network host zenika/alpine-chrome --no-sandbox --headless --disable-gpu --virtual-time-budget=15000 --dump-dom http://localhost:3129/python.html`.
- Each page's `data-files` attribute must list every example file the run needs.

## Gotchas

- On macOS with Colima, Docker only mounts paths under `$HOME`; `just serve` points Codapi's `TMPDIR` at `.tmp/codapi`. If Colima stops, `colima start`, `docker context use colima`, `just up`.
- Python programs run through `codapi/sandboxes/temporal-python/run_program.py`, which exits with `os._exit` to avoid temporalio/sdk-python#300.
- Java Activities must rethrow with `Activity.wrap(e)`; extra exception layers push the failure past the server's size limit and the Web UI shows "Failure exceeds size limit."
- codapi-js is not used: its 10-second fetch timeout breaks outage runs, so `tutorial.js` posts to `/v1/exec` itself.
- Code files start with a two-line `ABOUTME:` header; `tutorial.js` never shows it because snippets only include marked regions.
- Commits must be signed (`git commit -S`, SSH signing key). Local `git log --show-signature` reports "No signature" without an allowed-signers file; GitHub still verifies it.
