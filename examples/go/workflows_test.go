// ABOUTME: Workflow test for the Go IP geolocation tutorial, adapted from temporalio/temporal-tutorial-ipgeo-go.
// Mocks both Activities so the Workflow logic is tested on its own.
// @@@SNIPSTART go-ipgeo-workflow-test-setup
package iplocate_test

import (
	"testing"

	"temporal-ip-geolocation/iplocate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.temporal.io/sdk/testsuite"
)

// @@@SNIPEND

// @@@SNIPSTART go-ipgeo-workflow-test-workflow

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	activities := &iplocate.IPActivities{}

	// Mock activity implementation
	env.OnActivity(activities.GetIP, mock.Anything).Return("1.1.1.1", nil)
	env.OnActivity(activities.GetLocationInfo, mock.Anything, "1.1.1.1").Return("Planet Earth", nil)

	env.ExecuteWorkflow(iplocate.GetAddressFromIP, "Temporal")

	var result string
	assert.NoError(t, env.GetWorkflowResult(&result))
	assert.Equal(t, "Hello, Temporal. The machine running this code has the IP address 1.1.1.1 and is located in Planet Earth", result)
}

// @@@SNIPEND
