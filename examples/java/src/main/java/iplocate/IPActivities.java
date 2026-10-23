// ABOUTME: Activity interface for the Java IP geolocation tutorial, mirroring temporalio/temporal-tutorial-ipgeo-go and -ts.
// The Workflow calls Activities through this interface; IPActivitiesImpl holds the HTTP calls.
// @@@SNIPSTART java-ipgeo-activity-interface
package iplocate;

import io.temporal.activity.ActivityInterface;

@ActivityInterface
public interface IPActivities {
    String getIP();

    String getLocationInfo(String ip);
}
// @@@SNIPEND
