// ABOUTME: Task Queue name shared by the Worker and client, adapted from temporalio/temporal-tutorial-ipgeo-ts.
// Both sides must use the same name or the Worker never receives the client's Workflow.
// @@@SNIPSTART ts-ipgeo-shared
export const TASK_QUEUE_NAME="ip-address-ts";
// @@@SNIPEND
