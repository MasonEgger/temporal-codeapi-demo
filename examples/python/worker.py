# ABOUTME: Worker for the Python IP geolocation tutorial, mirroring the Go and TypeScript tutorials.
# Polls the Task Queue and runs the Workflow and Activities it receives.
# @@@SNIPSTART python-ipgeo-worker
import asyncio
import concurrent.futures

from temporalio.client import Client
from temporalio.worker import Worker

from activities import get_ip, get_location_info
from shared import TASK_QUEUE_NAME
from workflows import GetAddressFromIP


async def main() -> None:
    # Create the Temporal client
    client = await Client.connect("localhost:7233", namespace="default")

    # The Activities are regular functions, so the Worker runs them on a thread pool.
    with concurrent.futures.ThreadPoolExecutor(max_workers=100) as activity_executor:
        # Create the Temporal Worker and register the Workflow and Activities
        worker = Worker(
            client,
            task_queue=TASK_QUEUE_NAME,
            workflows=[GetAddressFromIP],
            activities=[get_ip, get_location_info],
            activity_executor=activity_executor,
        )

        # Start the Worker
        await worker.run()


if __name__ == "__main__":
    asyncio.run(main())
# @@@SNIPEND
