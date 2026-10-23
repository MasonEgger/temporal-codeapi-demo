# temporal-codeapi-demo

A prototype of the learn.temporal.io tutorial **Build a Temporal Application from scratch** that runs entirely in the browser, in Python, Go, Java, and TypeScript.
Learners read the real tutorial text, run the real code, run its tests, explore each run in the Temporal Web UI, and switch off the network to watch Temporal retry failed Activities.
Every run gets its own Temporal Service, Worker, and client in an isolated container, built on [Codapi](https://codapi.org).

## Quick start

You need Docker (Colima works), [just](https://github.com/casey/just), and [uv](https://docs.astral.sh/uv/).

```bash
just build
```

```bash
just install-codapi
```

```bash
just up
```

```bash
just serve
```

Then open <http://localhost:3129/>.
`just build` builds the gateway, the shared base image, and one sandbox image per language; `just up` starts the gateway; `just serve` runs Codapi in the foreground.
The gateway forwards `/v1/` to Codapi, so the browser only ever talks to port 3129.

If Docker stops (for example Colima after a sleep), bring everything back with `colima start`, `docker context use colima`, and `just up`.

To check everything (ruff, mypy, the gateway's Go tests, and end-to-end smoke tests against the running stack):

```bash
just check
```

## How it works

```
browser (web/*.html + tutorial.js)
   │  pages, network switch API, Web UI      ┌──────────────── gateway (Go, :3129 / :3128) ────────────────┐
   ├────────────────────────────────────────►│ serves web/ and examples/                                  │
   │                                         │ per-session network switch: /api/sessions/{id}/network       │
   │                                         │ per-session Web UI reverse proxy: /ui/{id}/…                 │
   │                                         │ forward proxy: the sandbox network's only way out,          │
   │                                         │   allowing only icanhazip.com and ip-api.com                 │
   │                                         └──────────────▲───────────────────────────────▲────────────────┘
   │  POST /v1/exec (all files + session ID + name)          │ registration, HTTP egress     │ Web UI
   │  forwarded by the gateway to Codapi                     │                               │
   ▼                                                         │                               │
Codapi (:1313) ── docker run --detach ──► sandbox container on an internal network ──────────┘
                  docker exec          ─► temporal-serve: dev server + Web UI under /ui/{session}, lingers 10 min
                                          temporal-run:   Worker in the background, client in the foreground
```

- **Sandboxes** (`codapi/sandboxes/temporal-<language>/`): one image per SDK with dependencies preinstalled and caches pre-warmed by running the tutorial's tests at build time. Each has `run`, `history`, and `test` commands.
- **Gateway** (`gateway/`): learners' network switch, Web UI routing, the egress allowlist, and the `/v1/` route to Codapi with an optional per-address run limit. When a learner's switch is off, the proxy returns `503` to their container, so Activities fail and Temporal retries them.
- **Pages** (`web/`): `tutorial.js` renders `web/tutorials/<language>.md` with marked, swaps each `<!--SNIPSTART id-->` block for the live snippet from `examples/` (marked with `@@@SNIPSTART` comments, like Temporal's snipsync), and wires up run consoles, the network switch, the embedded Web UI, and quizzes. While a run waits on retries, the console reads the run's pending Activities from the Web UI API and prints one line per failed attempt.

## Tutorial content

`web/tutorials/source/` holds untouched copies of the live learn.temporal.io tutorials (from `temporalio/temporal-learning`, MIT licensed).
The playground pages are those copies with surgical edits, so every change can be reviewed as a diff:

| Page | Base text | Edits |
|---|---|---|
| Go | Go tutorial | Playground only |
| TypeScript | TypeScript tutorial | Playground only |
| Python | Go tutorial | Python specifics plus playground |
| Java | Go tutorial | Java specifics plus playground |

```bash
git diff --no-index web/tutorials/source/go.md web/tutorials/python.md
```

Playground edits are the same on every page: a playground note in Prerequisites, local setup steps in a collapsed `:::setup` block, an "In the playground…" sentence after each local-only instruction, Run and Run tests buttons (including a "Run the Workflow test" button at the tutorial's first test run, before the Activity tests exist), two exercise buttons (a failing test and a five-attempt retry limit), the network switch, the Web UI panel, and the review questions as a quiz.
Sample output and Web UI screenshots from the original tutorials stay in the text but fold into closed accordions, since the live console and Web UI panel show the real thing.

The code in `examples/go` and `examples/typescript` is adapted from `temporalio/temporal-tutorial-ipgeo-go` and `-ts`; `examples/python` and `examples/java` mirror it.
In all four, the greeting says "the machine running this code" because in the playground the Activities report the server's IP address and location, not the learner's.

## Things worth knowing

- **Capacity.** Codapi runs at most `pool_size` (4) programs at once and answers `busy` beyond that. Each run's Temporal Service then lingers for up to 10 minutes (about 100 MB idle) so the Web UI works, but a learner's next run replaces it.
- **ip-api.com rate limit.** Every learner shares the server's IP address, and ip-api.com's free tier allows 45 requests a minute per IP. A busy page will see the location Activity fail and retry.
- **Screenshots.** The Python and Java pages reuse the Go tutorial's Web UI screenshots, which show Go names.
- **Python exit hang.** Python programs run through `run_program.py`, which exits with `os._exit` to avoid an SDK finalization hang under CPU contention ([temporalio/sdk-python#300](https://github.com/temporalio/sdk-python/issues/300)).
- **Exposure.** Containers run learner code. Locally they share the internal network and can reach each other's Web UIs. The hosted setup in `deploy/` runs on a dedicated droplet, drops traffic between sandboxes and from sandboxes to the host, and limits runs per client address; see `deploy/README.md`.

## Hosting

`deploy/` holds Terraform for a DigitalOcean droplet, firewall, and DNS record, and Ansible for everything on the droplet (Docker, Codapi under systemd, the gateway, Caddy for TLS).
`just infra-apply`, `just deploy`, then `just smoke https://temporal.egger.codes`; `deploy/README.md` has the details and the teardown command.
