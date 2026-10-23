# ABOUTME: Client for the Python IP geolocation tutorial, mirroring the Go and TypeScript tutorials.
# Starts the Workflow with a name from the command line and prints the result.
# @@@SNIPSTART python-ipgeo-cli-client
import asyncio
import sys
import uuid

from temporalio.client import Client

from shared import TASK_QUEUE_NAME
from workflows import GetAddressFromIP


async def main() -> None:
    if len(sys.argv) <= 1:
        sys.exit("Must specify a name as the command-line argument")
    name = sys.argv[1]

    client = await Client.connect("localhost:7233")

    result = await client.execute_workflow(
        GetAddressFromIP.run,
        name,
        id=f"getAddressFromIP-{uuid.uuid4()}",
        task_queue=TASK_QUEUE_NAME,
    )
    print(result)


if __name__ == "__main__":
    asyncio.run(main())
# @@@SNIPEND
