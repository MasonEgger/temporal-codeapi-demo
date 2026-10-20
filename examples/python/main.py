# ABOUTME: Hello World for the temporal-python Codapi sandbox.
# Defines a Workflow and Activity, then runs a Worker and Client against the per-run dev server.
"""Run a Hello World Workflow, its Worker, and a Client in a single program."""

import asyncio
import uuid
from datetime import timedelta

from temporalio import activity, workflow
from temporalio.client import Client
from temporalio.worker import Worker

TASK_QUEUE = "hello-task-queue"


@activity.defn
async def compose_greeting(name: str) -> str:
    """Build a greeting for the given name.

    Args:
        name: Who to greet.

    Returns:
        The greeting text.
    """
    return f"Hello, {name}!"


@workflow.defn
class HelloWorkflow:
    """Greet someone by calling a single Activity."""

    @workflow.run
    async def run(self, name: str) -> str:
        """Execute the Workflow.

        Args:
            name: Who to greet.

        Returns:
            The greeting produced by the Activity.
        """
        return await workflow.execute_activity(
            compose_greeting,
            name,
            start_to_close_timeout=timedelta(seconds=10),
        )


async def main() -> None:
    """Start a Worker, execute the Workflow through it, and print the result."""
    client = await Client.connect("localhost:7233")

    # Run the Worker in the background while the Client executes the Workflow.
    async with Worker(
        client,
        task_queue=TASK_QUEUE,
        workflows=[HelloWorkflow],
        activities=[compose_greeting],
    ):
        result = await client.execute_workflow(
            HelloWorkflow.run,
            "Temporal",
            id=f"hello-workflow-{uuid.uuid4()}",
            task_queue=TASK_QUEUE,
        )

    print(f"Workflow result: {result}")


if __name__ == "__main__":
    asyncio.run(main())
