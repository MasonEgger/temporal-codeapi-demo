// ABOUTME: Hello World for the temporal-java Codapi sandbox.
// Defines a Workflow and Activity, then runs a Worker and Client against the per-run dev server.
import io.temporal.activity.ActivityInterface;
import io.temporal.activity.ActivityOptions;
import io.temporal.client.WorkflowClient;
import io.temporal.client.WorkflowOptions;
import io.temporal.serviceclient.WorkflowServiceStubs;
import io.temporal.worker.Worker;
import io.temporal.worker.WorkerFactory;
import io.temporal.workflow.Workflow;
import io.temporal.workflow.WorkflowInterface;
import io.temporal.workflow.WorkflowMethod;
import java.time.Duration;
import java.util.UUID;

public class Main {
    static final String TASK_QUEUE = "hello-task-queue";

    @ActivityInterface
    public interface GreetingActivities {
        String composeGreeting(String name);
    }

    public static class GreetingActivitiesImpl implements GreetingActivities {
        @Override
        public String composeGreeting(String name) {
            return "Hello, " + name + "!";
        }
    }

    @WorkflowInterface
    public interface HelloWorkflow {
        @WorkflowMethod
        String greet(String name);
    }

    public static class HelloWorkflowImpl implements HelloWorkflow {
        private final GreetingActivities activities = Workflow.newActivityStub(
                GreetingActivities.class,
                ActivityOptions.newBuilder()
                        .setStartToCloseTimeout(Duration.ofSeconds(10))
                        .build());

        @Override
        public String greet(String name) {
            return activities.composeGreeting(name);
        }
    }

    public static void main(String[] args) {
        WorkflowServiceStubs service = WorkflowServiceStubs.newLocalServiceStubs();
        WorkflowClient client = WorkflowClient.newInstance(service);

        // Run the Worker in the background while the Client executes the Workflow.
        WorkerFactory factory = WorkerFactory.newInstance(client);
        Worker worker = factory.newWorker(TASK_QUEUE);
        worker.registerWorkflowImplementationTypes(HelloWorkflowImpl.class);
        worker.registerActivitiesImplementations(new GreetingActivitiesImpl());
        factory.start();

        HelloWorkflow workflow = client.newWorkflowStub(
                HelloWorkflow.class,
                WorkflowOptions.newBuilder()
                        .setWorkflowId("hello-workflow-" + UUID.randomUUID())
                        .setTaskQueue(TASK_QUEUE)
                        .build());
        String result = workflow.greet("Temporal");
        System.out.println("Workflow result: " + result);

        factory.shutdown();
        System.exit(0);
    }
}
