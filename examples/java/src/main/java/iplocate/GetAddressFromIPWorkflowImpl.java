// ABOUTME: Workflow implementation for the Java IP geolocation tutorial, mirroring temporalio/temporal-tutorial-ipgeo-go and -ts.
// Calls the two Activities in order with a Retry Policy, then builds the greeting.
// @@@SNIPSTART java-ipgeo-workflow-imports
package iplocate;

import io.temporal.activity.ActivityOptions;
import io.temporal.common.RetryOptions;
import io.temporal.failure.ActivityFailure;
import io.temporal.failure.ApplicationFailure;
import io.temporal.workflow.Workflow;
import java.time.Duration;

// @@@SNIPEND

// @@@SNIPSTART java-ipgeo-workflow-code
// The Temporal Workflow.
public class GetAddressFromIPWorkflowImpl implements GetAddressFromIPWorkflow {
    // Load Activities and assign the Retry Policy
    private final IPActivities activities = Workflow.newActivityStub(
            IPActivities.class,
            ActivityOptions.newBuilder()
                    // maximum time allowed for a single Activity Task Execution
                    .setStartToCloseTimeout(Duration.ofMinutes(1))
                    .setRetryOptions(RetryOptions.newBuilder()
                            // amount of time that must elapse before the first retry occurs
                            .setInitialInterval(Duration.ofSeconds(1))
                            // maximum interval between retries
                            .setMaximumInterval(Duration.ofMinutes(1))
                            // how much the retry interval increases
                            .setBackoffCoefficient(2)
                            // Uncomment this if you want to limit attempts
                            // .setMaximumAttempts(5)
                            .build())
                    .build());

    @Override
    public String getAddressFromIP(String name) {
        String ip;
        try {
            ip = activities.getIP();
        } catch (ActivityFailure e) {
            throw ApplicationFailure.newFailureWithCause("Failed to get IP", "GetIPFailed", e);
        }

        String location;
        try {
            location = activities.getLocationInfo(ip);
        } catch (ActivityFailure e) {
            throw ApplicationFailure.newFailureWithCause("Failed to get location", "GetLocationFailed", e);
        }

        return String.format("Hello, %s. The machine running this code has the IP address %s and is located in %s", name, ip, location);
    }
}
// @@@SNIPEND
