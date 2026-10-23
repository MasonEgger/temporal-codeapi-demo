# Session Summary: Session History Handoff Between Run Containers

**Date**: 2026-10-23
**Duration**: ~1.5 hours (two local image builds included)
**Conversation Turns**: 1 user turn
**Estimated Cost**: ~$8
**Model**: Claude Fable 5.1

## Key Actions

- Mason reported that repeated runs showed "the same Workflow ID" in the Web UI and asked for a real UUID in the ID. A two-run probe against the live site showed the IDs already differ (all four clients generate one per run); the Web UI only ever listed the latest because every run starts a fresh Temporal Service in a fresh container. Reported the real cause and built the fix for the intent: a session's runs accumulate in one Web UI.
- Gateway: `/register/{session}` now answers `handoff` when a previous container exists (or `none`); `POST /handoff/{session}` stores the superseded container's SQLite database (32 MiB cap, 15 minute life); `GET /handoff/{session}` returns it only if it came from the container the newest one replaced, so a stale upload from two runs ago is never used. 5 new tests; existing registration test updated for the 200 with a body.
- `temporal-serve`: registers before starting the server, waits up to 12 s for a database when told to, starts `start-dev --db-filename`, polls `/current` every second, and on supersede or linger timeout stops the server cleanly (SIGTERM, wait) and uploads the database. `temporal-run`: waits up to 30 s for the server and filters `--history` output by `StartTime` so each run prints only its own Workflows.
- The four tutorial pages now say the panel shows this session's Temporal Service and every run appears there. `TestSession` runs twice in one session and asserts both Workflow IDs appear in the Web UI list and each run printed one history.
- Local build failed: `TARGETARCH` is BuildKit-only and this machine's Docker 29 lacks buildx, so the Temporal CLI URL had an empty arch (HTTP 400). The shared Dockerfile falls back to `dpkg --print-architecture`.
- Verified locally end to end (gateway plus Codapi on this machine): handoff of a 676 KB database arrived one second after the second container registered, about 1.7 s added to the second run; 32 smoke tests pass.

## Prompt Inventory

| Prompt/Command | Action Taken | Outcome |
|---|---|---|
| Found a bug, the uuid on the Workflow ID is the same... add an actual UUID | Probe showed IDs differ; root cause is one dev server per run; built the database handoff through the gateway, fixed the legacy-builder Dockerfile bug, tested locally | PR opened; 32 passed locally |

## Efficiency Insights

**What went well:**
- A 20-line probe through the live API settled the premise (IDs differ, UI lists one) before any code was written.
- Measuring the dev server database (660 KB, clean shutdown, restart keeps the Workflow) on the droplet in one command settled the design.

**What could improve:**
- Starting Codapi by hand without `mkdir -p .tmp/codapi` cost one failed test run; `just serve` exists for a reason.

**Course corrections:**
- Rejected per-session networks, container reuse, and history import (Temporal has no import) before settling on handing the SQLite file over through the gateway, which already sits between sandboxes.

## Process Improvements

- When a bug report names a cause, verify the cause with a probe before fixing; the reported symptom was real, the diagnosis was not.
- Local image builds need the `TARGETARCH` fallback whenever a Dockerfile uses BuildKit build args; check `docker buildx version` on a new machine.

## Observations

- Docker 29 without the buildx plugin silently uses the legacy builder, which leaves `TARGETARCH` empty.
- The Temporal dev server shuts down cleanly on SIGTERM and leaves a single SQLite file; a restart from that file keeps namespaces and Workflows.

## Suggested Skills for Next Session

- `temporal:temporal-developer`: if the Web UI's handling of a reopened database (search attributes, visibility) shows anything odd.
