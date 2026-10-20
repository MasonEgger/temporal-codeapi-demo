# ABOUTME: Smoke tests that send each SDK example through a live Codapi server.
# Requires the sandbox images (`just build`) and a running server (`just serve`).
"""End-to-end checks for the Temporal Codapi sandboxes."""

from collections.abc import Iterator
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path

import httpx
import pytest
from pydantic import BaseModel

CODAPI_EXEC_URL = "http://localhost:1313/v1/exec"
EXAMPLES_DIR = Path(__file__).resolve().parent.parent / "examples"
EXPECTED_RESULT = "Workflow result: Hello, Temporal!"

# Files each sandbox receives; the empty name is the command's entry file.
SANDBOX_FILES: dict[str, dict[str, str]] = {
    "temporal-python": {"": "python/main.py"},
    "temporal-go": {"": "go/main.go"},
    "temporal-java": {"": "java/Main.java"},
    "temporal-typescript": {
        "": "typescript/main.ts",
        "workflows.ts": "typescript/workflows.ts",
        "activities.ts": "typescript/activities.ts",
    },
}


class Execution(BaseModel):
    """The subset of a Codapi `/v1/exec` response these tests inspect."""

    ok: bool
    duration: int
    stdout: str
    stderr: str


@pytest.fixture(scope="session")
def codapi() -> Iterator[httpx.Client]:
    """Yield an HTTP client for the local Codapi server.

    Yields:
        A client with a timeout long enough for a cold sandbox run.
    """
    with httpx.Client(timeout=60) as client:
        yield client


def execute(codapi: httpx.Client, sandbox: str, command: str, files: dict[str, str]) -> Execution:
    """Run files in a Codapi sandbox.

    Args:
        codapi: Client for the Codapi server.
        sandbox: Sandbox name, such as ``temporal-python``.
        command: Sandbox command to run, such as ``run`` or ``history``.
        files: File names mapped to their contents.

    Returns:
        The parsed execution result.

    Raises:
        pytest.fail.Exception: The Codapi server is not reachable.
    """
    try:
        response = codapi.post(CODAPI_EXEC_URL, json={"sandbox": sandbox, "command": command, "files": files})
    except httpx.ConnectError as error:
        pytest.fail(f"Codapi is not reachable at {CODAPI_EXEC_URL}; start it with `just serve` ({error})")
    response.raise_for_status()
    return Execution.model_validate_json(response.text)


def example_files(sandbox: str) -> dict[str, str]:
    """Load the example program for a sandbox.

    Args:
        sandbox: Sandbox name, a key of ``SANDBOX_FILES``.

    Returns:
        File names mapped to their contents, ready to send to Codapi.
    """
    return {name: (EXAMPLES_DIR / path).read_text() for name, path in SANDBOX_FILES[sandbox].items()}


@pytest.mark.parametrize("sandbox", SANDBOX_FILES)
class TestExamples:
    """Every SDK example completes its Workflow inside its sandbox."""

    def test_run_prints_workflow_result(self, codapi: httpx.Client, sandbox: str) -> None:
        execution = execute(codapi, sandbox, "run", example_files(sandbox))

        assert execution.ok, execution.stderr
        assert EXPECTED_RESULT in execution.stdout

    def test_history_prints_completed_event_history(self, codapi: httpx.Client, sandbox: str) -> None:
        execution = execute(codapi, sandbox, "history", example_files(sandbox))

        assert execution.ok, execution.stderr
        assert "=== Event History: hello-workflow-" in execution.stdout
        assert "WorkflowExecutionCompleted" in execution.stdout


class TestConcurrency:
    """Runs that share the Codapi worker pool still finish on time."""

    def test_parallel_python_runs_all_exit(self, codapi: httpx.Client) -> None:
        # Four runs fill the default pool; repeated batches surface the intermittent exit hang
        # from temporalio/sdk-python#300 that CPU contention triggers.
        files = example_files("temporal-python")
        with ThreadPoolExecutor(max_workers=4) as pool:
            executions = list(pool.map(lambda _: execute(codapi, "temporal-python", "run", files), range(20)))

        assert all(execution.ok for execution in executions), [execution.stderr for execution in executions]


class TestPythonErrors:
    """Python programs report failures the way running them directly would."""

    def test_exit_message_is_reported(self, codapi: httpx.Client) -> None:
        execution = execute(codapi, "temporal-python", "run", {"": "raise SystemExit('boom')"})

        assert not execution.ok
        assert "boom" in execution.stderr

    def test_uncaught_exception_shows_only_learner_frames(self, codapi: httpx.Client) -> None:
        execution = execute(codapi, "temporal-python", "run", {"": "raise ValueError('bad input')"})

        assert not execution.ok
        assert 'main.py", line 1, in <module>' in execution.stderr
        assert "ValueError: bad input" in execution.stderr
        assert "runpy" not in execution.stderr


class TestIsolation:
    """Learner code cannot escape the limits the sandbox sets."""

    def test_outbound_network_is_blocked(self, codapi: httpx.Client) -> None:
        program = "import socket\nsocket.create_connection(('1.1.1.1', 443), timeout=2)\n"

        execution = execute(codapi, "temporal-python", "run", {"": program})

        assert not execution.ok
        assert "Network is unreachable" in execution.stderr
