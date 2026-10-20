// ABOUTME: Imports the Temporal Go SDK packages learners use so the image build can prime the Go build cache.
// It is compiled once at build time and never run.
package main

import (
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	_ = activity.GetLogger
	_ = client.Dial
	_ = temporal.NewApplicationError
	_ = worker.New
	_ = workflow.ExecuteActivity
}
