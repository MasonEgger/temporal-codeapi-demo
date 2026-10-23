# Session Summary: ip-geo Tutorial Playground (Gateway, Sandboxes, Python Page)

**Date**: 2026-10-20 to 2026-10-22
**Duration**: ~2 days of intermittent work
**Conversation Turns**: ~40
**Estimated Cost**: ~$60 (long agentic build, many Docker builds and browser checks)
**Model**: Claude Opus 5.5, Claude Sonnet 5.5

## Key Actions

- Replaced the Hello World examples with the learn.temporal.io ip-geolocation app: official Go/TS tutorial code (temporal-tutorial-ipgeo-go/-ts, kept with @@@SNIP markers) and new Python/Java code mirroring it.
- Built `gateway/` (Go, TDD): allowlisting forward proxy (icanhazip.com, ip-api.com), per-session network switch API, per-session reverse proxy to each run's Temporal Web UI, and the tutorial site with no-cache headers.
- Reworked the sandboxes: internal Docker network, Codapi detach+exec steps (`temporal-serve` keeps the dev server and UI up 10 minutes or until superseded; `temporal-run` runs Worker in background and client in foreground), `test` command per SDK, images pre-warmed with the tutorial tests.
- Changed the greeting in all four languages to "The machine running this code has the IP address X and is located in Y" so playground output is accurate.
- Built the page framework (`web/tutorial.js`/`.css`): renders learn-site Markdown with marked, swaps `<!--SNIPSTART-->` blocks for live snippets from `examples/`, `:::note`/`:::setup` blocks, run consoles, network switch (in-text only), Web UI iframe (sidebar collapsed), quizzes, live retry lines from the Web UI API, Worker log folded.
- Copied the live tutorial Markdown from the local temporal-learning checkout into `web/tutorials/source/`; `web/tutorials/python.md` is the Go text with surgical language- and playground-specific edits (58% of prose lines kept verbatim).
- Built the Java page the same way (two extra snippet blocks for the interfaces, Maven commands verified with a real `mvn test`) and the Go and TypeScript pages from their own originals with playground-only edits.
- Fixed Java Activity failures exceeding the server's size limit (rethrow with `Activity.wrap`) and the tutorial pom's Mockito setup (load Mockito as a javaagent in Surefire).
- Rewrote the smoke tests for the tutorial layout (16 end-to-end tests), gave examples/python its own canonical pyproject, exempted the vendored tutorial sources from ruff, and rewrote README and CLAUDE.md. `just check` passes.

## Prompt Inventory

| Prompt/Command | Action Taken | Outcome |
|---|---|---|
| Use MIAB ip-geo code, recreate the learn tutorials | Gateway, sandboxes, four ip-geo programs | All run and test via Codapi |
| Allowlist proxy + green/red toggle, runnable tests, Web UI | Gateway with switch API and UI routing | Live outage demo works in browser |
| Syntax highlighting, accordion, styling, read-only, quiz | Dropped codapi-js widget for custom consoles | Redesigned page |
| Learn site is the standard; copy and edit, don't rewrite | Markdown pipeline from temporal-learning sources | Python page from Go text |
| Ziggy default, accurate IP wording, sidebar collapsed | Greeting change in 4 languages, localStorage navOpen | Done |
| Error output during outage is noisy | Live retry lines from Temporal, Worker log folded | Clean console |
| Edits fine, get to demo state | Java/Go/TS pages, Java fixes, smoke tests, docs | Demo ready; commit waits on GPG unlock |

## Efficiency Insights

**What went well:**
- Driving the real browser (switch off, run, Web UI, switch on) caught two bugs tests missed: codapi-js's hard-coded 10s fetch timeout and stale cached scripts.
- Reading Temporal's own pending-Activity data for the retry lines works identically for all four SDKs.

**What could improve:**
- I wrote fresh tutorial prose before asking whether the user wanted the live text kept; that cost a full rewrite. Ask about fidelity to existing content up front.
- Changing harness behavior required repeated Codapi restarts; a `just restart` recipe would help.

**Course corrections:**
- MIAB code → official tutorial code for Go/TS once the user named learn.temporal.io as the standard.
- Fresh prose → copy-and-edit of the live Markdown.
- Global network switch in the top bar → switch in the tutorial text plus an "network is off" note at every Run button.

## Process Improvements

- When adapting existing documentation, ask "keep the original text and edit, or rewrite?" before writing anything.
- Verify interactive flows in a real browser, not only through the API, before reporting them done.

## Observations

- Temporal dev server's Web UI works behind a path prefix (`--ui-public-path`) and its sidebar state lives in localStorage `navOpen`, so same-origin embedding can control it.
- Java Activity failures surface in Temporal as "Failure exceeds size limit." in the playground; needs investigation on the Java page.

## Suggested Skills for Next Session

- `content-design:tutorial-writing`: Java, Go, and TypeScript pages follow the same copy-and-edit approach.
- `temporal:temporal-developer`: Java failure-size issue and SDK-specific idioms for the Java page.
- `python:python`: updating the smoke tests in `tests/` for the new layout.
