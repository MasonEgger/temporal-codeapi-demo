// ABOUTME: Workflow test for the Java IP geolocation tutorial, mirroring temporalio/temporal-tutorial-ipgeo-go and -ts.
// Mocks both Activities so the Workflow logic is tested on its own.
// @@@SNIPSTART java-ipgeo-workflow-test-setup
package iplocate;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;
import static org.mockito.Mockito.withSettings;

import io.temporal.testing.TestWorkflowEnvironment;
import io.temporal.testing.TestWorkflowExtension;
import io.temporal.worker.Worker;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.RegisterExtension;

// @@@SNIPEND

// @@@SNIPSTART java-ipgeo-workflow-test-workflow
public class GetAddressFromIPWorkflowTest {
    @RegisterExtension
    public static final TestWorkflowExtension testWorkflowExtension = TestWorkflowExtension.newBuilder()
            .registerWorkflowImplementationTypes(GetAddressFromIPWorkflowImpl.class)
            .setDoNotStart(true)
            .build();

    @Test
    public void testGetAddressFromIP(
            TestWorkflowEnvironment testEnv, Worker worker, GetAddressFromIPWorkflow workflow) {
        // Mock activity implementation
        IPActivities activities = mock(IPActivities.class, withSettings().withoutAnnotations());
        when(activities.getIP()).thenReturn("1.1.1.1");
        when(activities.getLocationInfo("1.1.1.1")).thenReturn("Planet Earth");
        worker.registerActivitiesImplementations(activities);
        testEnv.start();

        String result = workflow.getAddressFromIP("Temporal");

        assertEquals("Hello, Temporal. The machine running this code has the IP address 1.1.1.1 and is located in Planet Earth", result);
    }
}
// @@@SNIPEND
