# ABOUTME: Smoke tests that run each language's tutorial code through a live gateway and Codapi server.
# Requires the images (`just build`), the gateway (`just up`), and the Codapi server (`just serve`).
"""End-to-end checks for the Temporal tutorial playground sandboxes.

Set ``PLAYGROUND_URL`` to point the suite at a deployed playground instead of the local gateway.
"""

import os
import uuid
from collections.abc import Iterator
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path

import httpx
import pytest
from pydantic import BaseModel

PLAYGROUND_URL = os.environ.get("PLAYGROUND_URL", "http://localhost:3129").rstrip("/")
CODAPI_EXEC_URL = f"{PLAYGROUND_URL}/v1/exec"
EXAMPLES_DIR = Path(__file__).resolve().parent.parent / "examples"
GREETING = "Hello, Ziggy. The machine running this code has the IP address"
LANGUAGES = ["python", "go", "java", "typescript"]


class Execution(BaseModel):
    """The subset of a Codapi `/v1/exec` response these tests inspect."""

    ok: bool
    duration: int
    stdout: str
    stderr: str


@pytest.fixture(scope="session")
def codapi() -> Iterator[httpx.Client]:
    """Yield an HTTP client for the playground's Codapi API.

    Yields:
        A client with a timeout long enough for a cold sandbox run.
    """
    with httpx.Client(timeout=180) as client:
        yield client


def tutorial_files(language: str) -> dict[str, str]:
    """Load every example file for a language, keyed by its path inside the project.

    Args:
        language: One of ``LANGUAGES``.

    Returns:
        File paths mapped to their contents, plus a fresh session ID and the learner's name.
    """
    root = EXAMPLES_DIR / language
    # Skip caches that local tooling (pytest, mypy, ruff) leaves next to the sources.
    cache_dirs = {"__pycache__", ".mypy_cache", ".pytest_cache", ".ruff_cache"}
    files = {
        str(path.relative_to(root)): path.read_text()
        for path in root.rglob("*")
        if path.is_file() and not cache_dirs.intersection(path.relative_to(root).parts)
    }
    files["session.id"] = str(uuid.uuid4())
    files["name.txt"] = "Ziggy"
    return files


def execute(codapi: httpx.Client, sandbox: str, command: str, files: dict[str, str]) -> Execution:
    """Run files in a Codapi sandbox through the gateway.

    Args:
        codapi: Client for the playground's Codapi API.
        sandbox: Sandbox name, such as ``temporal-python``.
        command: Sandbox command to run: ``run``, ``history``, or ``test``.
        files: File names mapped to their contents.

    Returns:
        The parsed execution result.

    Raises:
        pytest.fail.Exception: The playground is not reachable.
    """
    try:
        response = codapi.post(CODAPI_EXEC_URL, json={"sandbox": sandbox, "command": command, "files": files})
    except httpx.ConnectError as error:
        pytest.fail(f"The playground is not reachable at {CODAPI_EXEC_URL}; run `just up` and `just serve` ({error})")
    response.raise_for_status()
    return Execution.model_validate_json(response.text)


@pytest.mark.parametrize("language", LANGUAGES)
class TestTutorials:
    """Every language's tutorial code runs, records its history, and passes its own tests."""

    def test_run_prints_greeting(self, codapi: httpx.Client, language: str) -> None:
        execution = execute(codapi, f"temporal-{language}", "run", tutorial_files(language))

        assert execution.ok, execution.stderr
        assert GREETING in execution.stdout

    def test_history_shows_completed_workflow(self, codapi: httpx.Client, language: str) -> None:
        execution = execute(codapi, f"temporal-{language}", "history", tutorial_files(language))

        assert execution.ok, execution.stderr
        assert "=== Event History: getAddressFromIP-" in execution.stdout
        assert "WorkflowExecutionCompleted" in execution.stdout

    def test_tutorial_tests_pass(self, codapi: httpx.Client, language: str) -> None:
        execution = execute(codapi, f"temporal-{language}", "test", tutorial_files(language))

        assert execution.ok, execution.stderr


# The Activity test file each page leaves out for its "Run the Workflow test" button (data-omit-files).
ACTIVITY_TEST_FILES = {
    "python": "tests/test_activities.py",
    "go": "activities_test.go",
    "java": "src/test/java/iplocate/IPActivitiesTest.java",
    "typescript": "src/mocha/activities.test.ts",
}


@pytest.mark.parametrize("language", LANGUAGES)
class TestFirstTest:
    """The tutorial's first test run, before the Activity tests exist, runs the Workflow test alone."""

    def test_workflow_test_runs_without_the_activity_tests(self, codapi: httpx.Client, language: str) -> None:
        files = tutorial_files(language)
        del files[ACTIVITY_TEST_FILES[language]]

        execution = execute(codapi, f"temporal-{language}", "test", files)

        assert execution.ok, execution.stderr
        assert ACTIVITY_TEST_FILES[language].rsplit("/", 1)[-1].split(".")[0] not in execution.stdout


class TestConcurrency:
    """Runs that share the Codapi worker pool still finish on time."""

    def test_parallel_python_runs_all_exit(self, codapi: httpx.Client) -> None:
        # Four runs fill the default pool; repeated batches surface the intermittent exit hang
        # from temporalio/sdk-python#300 that CPU contention triggers.
        with ThreadPoolExecutor(max_workers=4) as pool:
            executions = list(
                pool.map(lambda _: execute(codapi, "temporal-python", "run", tutorial_files("python")), range(12))
            )

        assert all(execution.ok for execution in executions), [execution.stderr for execution in executions]


class TestIsolation:
    """Learner code can only reach the two services the tutorial uses."""

    def test_failing_program_reports_error(self, codapi: httpx.Client) -> None:
        files = tutorial_files("python")
        files["client.py"] = "raise SystemExit('boom')\n"

        execution = execute(codapi, "temporal-python", "run", files)

        assert not execution.ok
        assert "boom" in execution.stderr

    def test_hosts_outside_the_allowlist_are_blocked(self, codapi: httpx.Client) -> None:
        files = tutorial_files("python")
        files["client.py"] = "import requests\nprint(requests.get('https://example.com', timeout=10).status_code)\n"

        execution = execute(codapi, "temporal-python", "run", files)

        assert not execution.ok
        assert "403" in execution.stderr

    def test_direct_connections_bypassing_the_proxy_fail(self, codapi: httpx.Client) -> None:
        files = tutorial_files("python")
        files["client.py"] = "import socket\nsocket.create_connection(('1.1.1.1', 443), timeout=5)\n"

        execution = execute(codapi, "temporal-python", "run", files)

        assert not execution.ok
        assert "unreachable" in execution.stderr.lower() or "timed out" in execution.stderr.lower()


class TestSite:
    """The gateway serves the tutorial pages and the example sources the pages load."""

    @pytest.mark.parametrize(
        "path",
        ["", "python.html", "go.html", "java.html", "typescript.html", "tutorial.js", "tutorial.css"],
    )
    def test_page_is_served(self, codapi: httpx.Client, path: str) -> None:
        # The gateway bind-mounts web/ from the checkout; a stale mount after a redeploy 404s every page
        # while /v1/exec keeps working, so the run tests alone cannot catch it.
        response = codapi.get(f"{PLAYGROUND_URL}/{path}")

        assert response.status_code == 200, response.text[:200]
        assert "<html" in response.text or path.endswith((".js", ".css"))

    @pytest.mark.parametrize("language", LANGUAGES)
    def test_tutorial_markdown_and_examples_are_served(self, codapi: httpx.Client, language: str) -> None:
        tutorial = codapi.get(f"{PLAYGROUND_URL}/tutorials/{language}.md")
        example = codapi.get(f"{PLAYGROUND_URL}/examples/{language}/" + next(iter(tutorial_files(language))))

        assert tutorial.status_code == 200, tutorial.text[:200]
        assert "SNIPSTART" in tutorial.text
        assert example.status_code == 200, example.text[:200]
