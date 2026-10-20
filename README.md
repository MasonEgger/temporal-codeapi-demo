# temporal-codeapi-demo

Run real Temporal code from a web page with nothing installed.
This prototype uses [Codapi](https://codapi.org) to embed runnable, editable examples for the Python, Go, Java, and TypeScript SDKs.
Every run gets its own throwaway Temporal dev server, Worker, and Client inside an isolated container.

## How it works

```
browser (codapi-js widget)
   │  POST /v1/exec {sandbox, command, files}
   ▼
Codapi server (bin/codapi, port 1313)
   │  docker run --rm --network none --cpus 1 --memory … temporal-codapi/<sdk>
   ▼
temporal-run (inside the container)
   1. starts `temporal server start-dev --headless` on 127.0.0.1 (in-memory SQLite)
   2. waits for the health check
   3. copies the learner's files next to the preinstalled SDK and runs them
   4. with --history, prints `temporal workflow show` for every Workflow the program started
```

The container has no network, so the dev server, Worker, and Client all talk over loopback.
Nothing survives the run, and no state is shared between learners.

Each SDK is a Codapi sandbox under `codapi/sandboxes/temporal-<sdk>/` with a `Dockerfile`, a `box.json` (resource limits), and a `commands.json`.
Every sandbox has two commands: `run` (program output) and `history` (program output plus each Workflow's Event History).

| SDK | Entry file | How it runs | Preinstalled |
|---|---|---|---|
| Python | `main.py` | `python run_program.py main.py` | `temporalio` in a uv-managed venv |
| Go | `main.go` | `go run .` | SDK module and a primed build cache |
| Java | `Main.java` | single-file source launcher (`java Main.java`) | SDK jars on `CLASSPATH`, JDK 25 |
| TypeScript | `main.ts` (+ `workflows.ts`, `activities.ts`) | `node main.ts` with built-in type stripping | SDK packages, Node 24 |

## Requirements

- Docker. On macOS with Colima, only paths under `$HOME` can be mounted into containers; `just serve` handles this by pointing Codapi's temp dir at `.tmp/codapi`.
- [just](https://github.com/casey/just) and [uv](https://docs.astral.sh/uv/).

## Quick start

```bash
just build
```

```bash
just install-codapi
```

```bash
just serve
```

In a second terminal:

```bash
just web
```

Then open <http://localhost:8080/web/>.

To run the smoke tests (they need `just serve` running):

```bash
just check
```

## Layout

```
codapi/
  codapi.json                 server config: worker pool size and default (tight) box limits
  sandboxes/
    shared/                   base image: Temporal CLI + the temporal-run harness
    temporal-python/          box.json, commands.json, Dockerfile, run_program.py
    temporal-go/
    temporal-java/
    temporal-typescript/
examples/                     the Hello World programs, one per SDK
web/index.html                demo page; loads examples/ at runtime
tests/test_sandboxes.py       end-to-end tests through a live Codapi server
```

## Measurements

Measured on an Apple Silicon Mac through Colima (4 CPUs, 8 GB), each box capped at 1 CPU.
Times are Codapi's reported `duration` for a warm image, median of 5 runs.
They include starting the dev server, starting the Worker, running the Workflow, and shutting down.

| SDK | Run time | Peak memory | Image size |
|---|---|---|---|
| Python | 1.4 s | 102 MiB | 560 MB |
| Go | 1.6 s | 173 MiB | 2.0 GB |
| Java | 2.6 s | 229 MiB | 863 MB |
| TypeScript | 2.7 s | 293 MiB | 1.1 GB |

With four concurrent runs (one per SDK), each run took 7 to 30 percent longer, Python the most.

## Things worth knowing

- **Python exit hang.** Under CPU contention, the Python SDK can hang during interpreter shutdown after the program finishes ([temporalio/sdk-python#300](https://github.com/temporalio/sdk-python/issues/300)).
  `run_program.py` runs the learner's file and then exits with `os._exit` to skip finalization.
  Before this, roughly 1 in 15 batches of four parallel runs hit the 30-second timeout; `TestConcurrency` guards against regressions.
- **Busy responses.** Codapi rejects requests beyond `pool_size` (4 here) right away with `busy: try again later` instead of queuing them.
- **SDK logs.** The Go and TypeScript SDKs log at INFO by default (TypeScript also prints its Workflow bundle stats), so their output is noisier than Python's or Java's.
  The examples leave the default loggers alone so the code matches what learners would write.
- **Java warnings.** JDK 25 warns about `sun.misc.Unsafe` and native access from the SDK's shaded netty; `java.args` silences these.
- **Exposure.** Codapi listens on all interfaces and allows any origin. Codapi's docs say to run it on a dedicated machine with no credentials on it; follow that before putting this anywhere public.
