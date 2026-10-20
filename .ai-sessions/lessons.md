# Lessons Learned

## Recent
<!-- 10 most recent lessons, newest first -->
- Codapi reads `codapi.json`, `box.json`, and `commands.json` only at startup; restart the server after editing them (2026-10-20)
- Colima only mounts paths under `$HOME`; run Codapi with `TMPDIR` inside the repo so its per-run temp dirs reach the containers (2026-10-20)
- The Temporal Python SDK can hang at interpreter exit under CPU contention (sdk-python#300); in sandboxes, run the program via a launcher that flushes output and calls `os._exit` (2026-10-20)
- `temporal server start-dev --headless` works inside a `--network none` container because loopback still exists, so each run can have its own dev server (2026-10-20)
- On JDK 25, pass JVM flags via an `@argfile` instead of `JDK_JAVA_OPTIONS` to avoid the launcher's "Picked up" notice; add `--sun-misc-unsafe-memory-access=allow` and `--enable-native-access=ALL-UNNAMED` for the Temporal SDK's netty (2026-10-20)
- Confirm "build now or plan first?" before turning a feasibility discussion into a full implementation (2026-10-20)

## Workflow
- Confirm "build now or plan first?" before turning a feasibility discussion into a full implementation (2026-10-20)

## Infrastructure
- Colima only mounts paths under `$HOME`; run Codapi with `TMPDIR` inside the repo so its per-run temp dirs reach the containers (2026-10-20)
- Codapi reads `codapi.json`, `box.json`, and `commands.json` only at startup; restart the server after editing them (2026-10-20)

## Temporal
- `temporal server start-dev --headless` works inside a `--network none` container because loopback still exists, so each run can have its own dev server (2026-10-20)
- The Temporal Python SDK can hang at interpreter exit under CPU contention (sdk-python#300); in sandboxes, run the program via a launcher that flushes output and calls `os._exit` (2026-10-20)

## Java
- On JDK 25, pass JVM flags via an `@argfile` instead of `JDK_JAVA_OPTIONS` to avoid the launcher's "Picked up" notice; add `--sun-misc-unsafe-memory-access=allow` and `--enable-native-access=ALL-UNNAMED` for the Temporal SDK's netty (2026-10-20)
