# Session Summary: DigitalOcean Deployment (Spec, Gateway One-Origin, Terraform, Ansible)

**Date**: 2026-10-22
**Duration**: ~2 hours
**Conversation Turns**: ~6 user turns
**Estimated Cost**: ~$20 (repo read, Go TDD, Terraform init, Ansible authoring and lint)
**Model**: Claude Fable 5.1 (planning, gateway, infra), Claude Sonnet 5.5 (wrap-up)

## Key Actions

- Read the whole stack (gateway, sandboxes, Codapi config, session notes) and answered the per-user-container question: every run gets a fresh container; sessions supersede older containers; the shared internal network was the open exposure.
- Checked DigitalOcean spend with doctl: September invoice $133 fully discounted ("DO Alumni Discount (100%)"), standing resources ~$171/mo, so `s-4vcpu-8gb` fits for one or two weeks only.
- Asked four batched decisions (hostname, size, image build path, Codapi placement); Mason chose `temporal.egger.codes`, `s-4vcpu-8gb` in `sfo3`, build on the droplet from a git clone, Codapi as a host systemd service, Ansible defaulting to `main`.
- Wrote `spec.md` in the `/bpe:brainstorm` canonical section order with goals G1 to G7, deferred items, invariants, and success criteria.
- Gateway (TDD, Go): `-codapi-addr` forwards `/v1/` to Codapi so the page has one origin; `-run-rate`/`-run-burst` per-address token bucket on `/v1/exec` returning 429 with a JSON `stderr`; `-trust-proxy` keys on `X-Forwarded-For`. 11 new tests pass under `golang:1.26-bookworm`.
- Page and tests: `tutorial.js` defaults to same-origin `/v1` and shows the 429 message; smoke tests read `PLAYGROUND_URL` (default `http://localhost:3129`).
- `deploy/terraform/`: droplet, cloud firewall (22/80/443), A record in `egger.codes`, `local_file` Ansible inventory; `terraform validate` passes with provider 2.106.0.
- `deploy/ansible/`: roles base, docker, playground (just, checkout, conditional `just build`, fixed-subnet sandbox network, iptables isolation unit, compose gateway, housekeeping timer), codapi (user, pinned 0.14.2 with sha256, systemd unit), caddy (apt repo, Caddyfile). `ansible-lint` passes the production profile.
- `deploy/compose.yaml`, justfile recipes (`infra-*`, `deploy`, `deploy-ssh`, `smoke`), `deploy/README.md`, README and CLAUDE.md updates, `.gitignore` for state and inventory.
- Verified the gateway image locally: pages serve and `/v1/exec` reaches `host.docker.internal:1313` (502 with a clear message when Codapi is down).

## Prompt Inventory

| Prompt/Command | Action Taken | Outcome |
|---|---|---|
| Deploy to a DO droplet via Docker; per-user containers?; plan Ansible + Terraform | Read repo, checked tooling and doctl, asked 4 decisions, wrote spec.md and 2 memories | Plan delivered with budget math |
| Hostname temporal.egger.codes, sfo3, Ansible defaults to main; set up, merge, deploy; test deploys OK | Built gateway changes (TDD), Terraform, Ansible, compose, docs on `feature/digitalocean-deploy` | Everything lints and validates; test deploy pending |
| (commit process) `/bpe:session-summary` | This summary and lessons | Written |

## Efficiency Insights

**What went well:**
- Reading Codapi's upstream `codapi.service`, install guide, and `cmd/main.go` flags settled the systemd layout and the `-port`-only CLI before writing Ansible.
- Running the gateway tests in the Go 1.26 Docker image sidestepped the local toolchain download failure in one step.
- `ansible-lint --offline` at the production profile caught the cross-role handler shape before any droplet run.

**What could improve:**
- The Codapi tarball layout and just tarball layout had to be downloaded to learn (`include:` lists); release notes would have said.
- The Python skill's profile was not read before the three-line test edit; the edit is trivial but the skill asks for it on every invocation.

**Course corrections:**
- Replaced a `debug` task with `notify` as a Codapi-restart trigger with a real handler defined in the playground role, notified by the git checkout task.

## Process Improvements

- Define a handler in the earliest role whose task notifies it; handlers are play-wide and run at the end, so a later role's service can still be restarted by an earlier role's change.
- Use `doctl invoice summary <uuid>` for real spend; `doctl balance get` reads zero under a 100% discount.

## Observations

- Mason's DO account carries ~$171/mo of standing droplets under a $200 credit; only short-lived additions fit.
- The gateway's run limit and Codapi's pool-full response both return 429; the page distinguishes them by the JSON `stderr` the gateway includes.
- Sandbox isolation relies on br_netfilter sending bridged traffic through `DOCKER-USER`; this is unverified until the droplet test (next step).

## Suggested Skills for Next Session

- `python:python`: if the smoke tests gain a prod-only marker or the concurrency test needs a rate-limit-aware mode.
- `temporal:temporal-developer`: if the dev server's Web UI misbehaves behind Caddy's TLS (mixed content, websocket).
