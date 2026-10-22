// ABOUTME: Task Queue name shared by the Worker and client, adapted from temporalio/temporal-tutorial-ipgeo-go.
// Both sides must use the same name or the Worker never receives the client's Workflow.
// @@@SNIPSTART go-ipgeo-shared
package iplocate

const TaskQueueName = "ip-address-go"

// @@@SNIPEND
