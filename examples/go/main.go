// ABOUTME: Hello World for the temporal-go Codapi sandbox.
// Defines a Workflow and Activity, then runs a Worker and Client against the per-run dev server.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

const taskQueue = "hello-task-queue"

// ComposeGreeting builds a greeting for the given name.
func ComposeGreeting(ctx context.Context, name string) (string, error) {
	return fmt.Sprintf("Hello, %s!", name), nil
}

// HelloWorkflow greets someone by calling a single Activity.
func HelloWorkflow(ctx workflow.Context, name string) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})

	var greeting string
	err := workflow.ExecuteActivity(ctx, ComposeGreeting, name).Get(ctx, &greeting)
	return greeting, err
}

func main() {
	c, err := client.Dial(client.Options{HostPort: "localhost:7233"})
	if err != nil {
		log.Fatalln("Unable to create client:", err)
	}
	defer c.Close()

	// Run the Worker in the background while the Client executes the Workflow.
	w := worker.New(c, taskQueue, worker.Options{})
	w.RegisterWorkflow(HelloWorkflow)
	w.RegisterActivity(ComposeGreeting)
	if err := w.Start(); err != nil {
		log.Fatalln("Unable to start Worker:", err)
	}
	defer w.Stop()

	run, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:        "hello-workflow-" + uuid.NewString(),
		TaskQueue: taskQueue,
	}, HelloWorkflow, "Temporal")
	if err != nil {
		log.Fatalln("Unable to start Workflow:", err)
	}

	var result string
	if err := run.Get(context.Background(), &result); err != nil {
		log.Fatalln("Workflow failed:", err)
	}
	fmt.Println("Workflow result:", result)
}
