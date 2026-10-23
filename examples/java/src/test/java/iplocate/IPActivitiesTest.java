// ABOUTME: Activity tests for the Java IP geolocation tutorial, mirroring temporalio/temporal-tutorial-ipgeo-go and -ts.
// Injects a mock HTTP client so the tests never touch the network.
// @@@SNIPSTART java-ipgeo-activity-test-setup
package iplocate;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import io.temporal.testing.TestActivityEnvironment;
import java.net.http.HttpClient;
import java.net.http.HttpResponse;
import org.junit.jupiter.api.Test;

public class IPActivitiesTest {
    // Build an HTTP client whose every request returns the given body
    @SuppressWarnings("unchecked")
    private static HttpClient mockHttpClient(String body) throws Exception {
        HttpResponse<String> response = mock(HttpResponse.class);
        when(response.body()).thenReturn(body);
        HttpClient httpClient = mock(HttpClient.class);
        when(httpClient.send(any(), any(HttpResponse.BodyHandler.class))).thenReturn(response);
        return httpClient;
    }

// @@@SNIPEND

// @@@SNIPSTART java-ipgeo-activity-test-ip
    @Test
    public void testGetIP() throws Exception {
        TestActivityEnvironment env = TestActivityEnvironment.newInstance();
        env.registerActivitiesImplementations(new IPActivitiesImpl(mockHttpClient("127.0.0.1\n")));
        IPActivities activities = env.newActivityStub(IPActivities.class);

        assertEquals("127.0.0.1", activities.getIP());
    }

// @@@SNIPEND

// @@@SNIPSTART java-ipgeo-activity-test-location
    @Test
    public void testGetLocationInfo() throws Exception {
        String body = """
                {"city": "San Francisco", "regionName": "California", "country": "United States"}
                """;
        TestActivityEnvironment env = TestActivityEnvironment.newInstance();
        env.registerActivitiesImplementations(new IPActivitiesImpl(mockHttpClient(body)));
        IPActivities activities = env.newActivityStub(IPActivities.class);

        assertEquals("San Francisco, California, United States", activities.getLocationInfo("127.0.0.1"));
    }
}
// @@@SNIPEND
