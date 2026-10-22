// ABOUTME: Client for the Java IP geolocation tutorial, mirroring temporalio/temporal-tutorial-ipgeo-go and -ts.
// Starts the Workflow with a name from the command line and prints the result.
// @@@SNIPSTART java-ipgeo-cli-client
package iplocate;

import io.temporal.client.WorkflowClient;
import io.temporal.client.WorkflowOptions;
import io.temporal.serviceclient.WorkflowServiceStubs;
import java.util.UUID;

public class IPLocateClient {
    public static void main(String[] args) {
        if (args.length < 1) {
            System.err.println("Must specify a name as the command-line argument");
            System.exit(1);
        }
        String name = args[0];

        WorkflowServiceStubs service = WorkflowServiceStubs.newLocalServiceStubs();
        WorkflowClient client = WorkflowClient.newInstance(service);

        WorkflowOptions options = WorkflowOptions.newBuilder()
                .setWorkflowId("getAddressFromIP-" + UUID.randomUUID())
                .setTaskQueue(Shared.TASK_QUEUE_NAME)
                .build();

        GetAddressFromIPWorkflow workflow = client.newWorkflowStub(GetAddressFromIPWorkflow.class, options);
        String result = workflow.getAddressFromIP(name);
        System.out.println(result);
        System.exit(0);
    }
}
// @@@SNIPEND
