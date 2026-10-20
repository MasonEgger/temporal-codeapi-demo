# Session Summary: Codapi Playground Prototype for Four Temporal SDKs

**Date**: 2026-10-20
**Duration**: ~2.5 hours
**Conversation Turns**: 4
**Estimated Cost**: ~$15 (long agentic build with many tool calls)
**Model**: Claude Opus 5.5 (build), Claude Sonnet 5.5 (wrap-up)

## Key Actions

- Read the #C02F5C9DC9W Slack thread on a no-install "sandbox" for Temporal and researched Codapi (server source, config layout, docker flags, defaults).
- Assessed feasibility: a per-run headless dev server inside each Codapi container works because `--network none` keeps loopback.
- Built a shared base image (Temporal CLI v1.9.1 + `temporal-run` harness) and Codapi sandboxes for Python, Go, Java, and TypeScript, each with `run` and `history` commands.
- Primed per-SDK images: uv venv (Python), module + build cache (Go), SDK jars + JDK 25 source launcher with an argfile (Java), npm deps + Node 24 type stripping (TypeScript).
- Found and worked around temporalio/sdk-python#300 (exit hang under CPU contention) with `run_program.py`, which exits via `os._exit`; added a concurrency regression test.
- Added a demo page that loads `examples/` at runtime, a justfile, and 12 end-to-end pytest checks; verified the page in the browser, including a multi-file TypeScript edit.
- Measured warm run times (1.4–2.7 s), peak memory (102–293 MiB), and image sizes; documented in README.

## Prompt Inventory

| Prompt/Command | Action Taken | Outcome |
|---|---|---|
| Read the Slack thread, explore codapi.org, assess with Temporal MCP | Read thread, codapi docs, Temporal docs | Feasibility answer: easy, with three architecture options and a Cloud caveat |
| Support the core 4 SDKs | Built the full prototype on a feature branch | Working sandboxes, tests, demo page, README |
| "Wait you already built the whole thing?" | Clarified scope and offered keep/step back/discard | User chose to keep it |
| What was the first workflow? / keep it | Explained Hello World; ran session summary and commit flow | This summary and the commit |

## Efficiency Insights

**What went well:**
- Validating one SDK end to end (base + Python) before templating the other three caught the Colima mount issue early.
- Reading Codapi's Go source directly (docker.go, load.go) answered config questions faster than its docs.
- Stress tests with a stuck-container watcher turned an intermittent timeout into a diagnosis within a few iterations.

**What could improve:**
- "Start with the core 4" was read as a go-ahead to build; the user was surprised. Confirm before turning a feasibility question into a full build.
- A py-spy debug image failed to build and the script kept going; check build success before running dependent steps.

**Course corrections:**
- Container cleanup with `docker rm -f` tripped a safety check; switched to `docker run --rm` plus `docker kill`.
- Java launcher noise moved from `JDK_JAVA_OPTIONS` to an `@argfile` to silence the "Picked up" note.

## Process Improvements

- Ask "build it now, or plan first?" when a reply could be either a requirement or a go-ahead.
- Restart Codapi after any `commands.json` or `box.json` change; it reads config only at startup.

## Observations

- Python's per-run dev server was faster than expected: 1.4 s including server startup.
- Go's primed build cache made `go run` nearly as fast as Python despite compiling.
- The Python SDK hang only appeared under contention, matching reports in sdk-python#300.

## Suggested Skills for Next Session

- `temporal:temporal-developer`: likely next step is richer examples (retries, timers, signals) across four SDKs.
- `python:python`: tests and the Python launcher follow the Python standards.
- `content-design:tutorial-writing`: if the playground grows into a "Tour of Temporal" with lesson text.
