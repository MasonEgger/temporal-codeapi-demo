// ABOUTME: Task Queue name shared by the Worker and client, mirroring temporalio/temporal-tutorial-ipgeo-go and -ts.
// Both sides must use the same name or the Worker never receives the client's Workflow.
// @@@SNIPSTART java-ipgeo-shared
package iplocate;

public final class Shared {
    public static final String TASK_QUEUE_NAME = "ip-address-java";

    private Shared() {}
}
// @@@SNIPEND
