# ABOUTME: Workflow for the Python IP geolocation tutorial, mirroring the Go and TypeScript tutorials.
# Calls the two Activities in order with a Retry Policy, then builds the greeting.
# @@@SNIPSTART python-ipgeo-workflow-imports
from datetime import timedelta

from temporalio import workflow
from temporalio.common import RetryPolicy
from temporalio.exceptions import ActivityError, ApplicationError

# Only pass the Activities through the sandbox; don't re-import them on every Workflow run.
with workflow.unsafe.imports_passed_through():
    from activities import get_ip, get_location_info
# @@@SNIPEND


# @@@SNIPSTART python-ipgeo-workflow-code
# The Temporal Workflow.
@workflow.defn
class GetAddressFromIP:
    @workflow.run
    async def run(self, name: str) -> str:
        retry_policy = RetryPolicy(
            initial_interval=timedelta(seconds=1),  # amount of time that must elapse before the first retry occurs
            maximum_interval=timedelta(minutes=1),  # maximum interval between retries
            backoff_coefficient=2,  # how much the retry interval increases
            # maximum_attempts=5,  # Uncomment this if you want to limit attempts
        )

        try:
            ip = await workflow.execute_activity(
                get_ip,
                start_to_close_timeout=timedelta(minutes=1),
                retry_policy=retry_policy,
            )
        except ActivityError as e:
            raise ApplicationError("Failed to get IP") from e

        try:
            location = await workflow.execute_activity(
                get_location_info,
                ip,
                start_to_close_timeout=timedelta(minutes=1),
                retry_policy=retry_policy,
            )
        except ActivityError as e:
            raise ApplicationError("Failed to get location") from e

        return f"Hello, {name}. The machine running this code has the IP address {ip} and is located in {location}"


# @@@SNIPEND
