// ABOUTME: Worker for the Java IP geolocation tutorial, mirroring temporalio/temporal-tutorial-ipgeo-go and -ts.
// Polls the Task Queue and runs the Workflow and Activities it receives.
// @@@SNIPSTART java-ipgeo-worker
package iplocate;

import io.temporal.client.WorkflowClient;
import io.temporal.serviceclient.WorkflowServiceStubs;
import io.temporal.worker.Worker;
import io.temporal.worker.WorkerFactory;
import java.net.http.HttpClient;

public class IPLocateWorker {
    public static void main(String[] args) {
        // Create the Temporal client
        WorkflowServiceStubs service = WorkflowServiceStubs.newLocalServiceStubs();
        WorkflowClient client = WorkflowClient.newInstance(service);

        // Create the Temporal Worker
        WorkerFactory factory = WorkerFactory.newInstance(client);
        Worker worker = factory.newWorker(Shared.TASK_QUEUE_NAME);

        // Register the Workflow, and inject an HTTP client into the Activities
        worker.registerWorkflowImplementationTypes(GetAddressFromIPWorkflowImpl.class);
        worker.registerActivitiesImplementations(new IPActivitiesImpl(HttpClient.newHttpClient()));

        // Start the Worker
        factory.start();
    }
}
// @@@SNIPEND
