# Lessons Learned

## Recent
<!-- 10 most recent lessons, newest first -->
- A bind mount follows the inode, not the path: when a deploy deletes and recreates a mounted directory (git checkout through a tree without it), recreate the container or every read under the mount fails with ENOENT (2026-10-23)
- Smoke tests must fetch what users fetch: the suite only exercised /v1/exec, so a site where every page 404'd passed 16 of 16 (2026-10-23)
- Docker 28+ does not load `br_netfilter`, so iptables `DOCKER-USER` rules never see traffic between containers on the same bridge; isolate same-bridge containers with an nftables `bridge`-family forward chain keyed on `meta ibrname`/`obrname` (allow ARP to the permitted peer too) (2026-10-22)
- Before reusing a DigitalOcean token copied from another repo's sops file, run `doctl account get --access-token "$TF_VAR_do_token"`; PyTexas's token is the PyTexas Developers team, not Mason's personal team (2026-10-22)
- Verify a firewall rule by exercising the path it is meant to block (and reading rule counters), not by listing the rule; a zero-packet chain is a rule that is not in the path (2026-10-22)
- Ansible handlers are play-wide and run at the end of the play, so define a handler in the earliest role whose task notifies it even if a later role installs the service (2026-10-22)
- ansible-lint's production profile wants registered variables prefixed with the role name (`playground_images`, not `image_inspect`) (2026-10-22)
- `doctl balance get` reads $0.00 under a 100% account discount; use `doctl invoice summary <uuid>` and `doctl compute droplet list` for real spend (2026-10-22)
- When the local Go toolchain cannot download the go.mod version, run `gofmt`, `go vet`, and `go test` inside `golang:<version>-bookworm` with `GOFLAGS=-buildvcs=false` (2026-10-22)
- When adapting existing documentation, ask whether to keep the original text and edit it before writing anything; fresh prose cost a full redo (2026-10-22)

## Workflow
- When adapting existing documentation, ask whether to keep the original text and edit it before writing anything; fresh prose cost a full redo (2026-10-22)
- Confirm "build now or plan first?" before turning a feasibility discussion into a full implementation (2026-10-20)

## Infrastructure
- A bind mount follows the inode, not the path: when a deploy deletes and recreates a mounted directory (git checkout through a tree without it), recreate the container or every read under the mount fails with ENOENT (2026-10-23)
- Docker 28+ does not load `br_netfilter`, so iptables `DOCKER-USER` rules never see traffic between containers on the same bridge; isolate same-bridge containers with an nftables `bridge`-family forward chain keyed on `meta ibrname`/`obrname` (allow ARP to the permitted peer too) (2026-10-22)
- Before reusing a DigitalOcean token copied from another repo's sops file, run `doctl account get --access-token "$TF_VAR_do_token"`; PyTexas's token is the PyTexas Developers team, not Mason's personal team (2026-10-22)
- Verify a firewall rule by exercising the path it is meant to block (and reading rule counters), not by listing the rule; a zero-packet chain is a rule that is not in the path (2026-10-22)
- Ansible handlers are play-wide and run at the end of the play, so define a handler in the earliest role whose task notifies it even if a later role installs the service (2026-10-22)
- ansible-lint's production profile wants registered variables prefixed with the role name (`playground_images`, not `image_inspect`) (2026-10-22)
- `doctl balance get` reads $0.00 under a 100% account discount; use `doctl invoice summary <uuid>` and `doctl compute droplet list` for real spend (2026-10-22)
- Colima only mounts paths under `$HOME`; run Codapi with `TMPDIR` inside the repo so its per-run temp dirs reach the containers (2026-10-20)
- Codapi reads `codapi.json`, `box.json`, and `commands.json` only at startup; restart the server after editing them (2026-10-20)

## Temporal
- `temporal server start-dev --headless` works inside a `--network none` container because loopback still exists, so each run can have its own dev server (2026-10-20)
- The Temporal Python SDK can hang at interpreter exit under CPU contention (sdk-python#300); in sandboxes, run the program via a launcher that flushes output and calls `os._exit` (2026-10-20)

## Java
- On JDK 25, pass JVM flags via an `@argfile` instead of `JDK_JAVA_OPTIONS` to avoid the launcher's "Picked up" notice; add `--sun-misc-unsafe-memory-access=allow` and `--enable-native-access=ALL-UNNAMED` for the Temporal SDK's netty (2026-10-20)

## Tooling
- When the local Go toolchain cannot download the go.mod version, run `gofmt`, `go vet`, and `go test` inside `golang:<version>-bookworm` with `GOFLAGS=-buildvcs=false` (2026-10-22)
- codapi-js aborts exec requests after a hard-coded 10s; register a custom engine on `window.codapi.engines` (or drop the widget) for long runs (2026-10-22)
- Temporal's dev server Web UI accepts `--ui-public-path` and stores its sidebar state in localStorage `navOpen`, so a same-origin iframe can collapse it (2026-10-22)

## Testing
- Smoke tests must fetch what users fetch: the suite only exercised /v1/exec, so a site where every page 404'd passed 16 of 16 (2026-10-23)
