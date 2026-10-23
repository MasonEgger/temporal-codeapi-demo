// ABOUTME: Workflow interface for the Java IP geolocation tutorial, mirroring temporalio/temporal-tutorial-ipgeo-go and -ts.
// Clients start the Workflow through this interface; GetAddressFromIPWorkflowImpl holds the logic.
// @@@SNIPSTART java-ipgeo-workflow-interface
package iplocate;

import io.temporal.workflow.WorkflowInterface;
import io.temporal.workflow.WorkflowMethod;

@WorkflowInterface
public interface GetAddressFromIPWorkflow {
    @WorkflowMethod
    String getAddressFromIP(String name);
}
// @@@SNIPEND
