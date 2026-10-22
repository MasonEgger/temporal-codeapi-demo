# ABOUTME: Workflow test for the Python IP geolocation tutorial, mirroring the Go and TypeScript tutorials.
# Mocks both Activities so the Workflow logic is tested on its own.
# @@@SNIPSTART python-ipgeo-workflow-test-setup
import uuid

import pytest
from temporalio import activity
from temporalio.testing import WorkflowEnvironment
from temporalio.worker import Worker

from workflows import GetAddressFromIP

# @@@SNIPEND


# @@@SNIPSTART python-ipgeo-workflow-test-workflow
# Mock Activities registered under the real Activities' names
@activity.defn(name="get_ip")
async def get_ip_mocked() -> str:
    return "1.1.1.1"


@activity.defn(name="get_location_info")
async def get_location_info_mocked(ip: str) -> str:
    return "Planet Earth"


@pytest.mark.asyncio
async def test_get_address_from_ip() -> None:
    task_queue = "test"
    async with (
        await WorkflowEnvironment.start_time_skipping() as env,
        Worker(
            env.client,
            task_queue=task_queue,
            workflows=[GetAddressFromIP],
            activities=[get_ip_mocked, get_location_info_mocked],
        ),
    ):
        result = await env.client.execute_workflow(
            GetAddressFromIP.run,
            "Temporal",
            id=f"test-{uuid.uuid4()}",
            task_queue=task_queue,
        )

    assert result == (
        "Hello, Temporal. The machine running this code has the IP address 1.1.1.1 and is located in Planet Earth"
    )


# @@@SNIPEND
