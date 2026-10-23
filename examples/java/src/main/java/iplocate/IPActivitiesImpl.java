// ABOUTME: Activity implementations for the Java IP geolocation tutorial, mirroring temporalio/temporal-tutorial-ipgeo-go and -ts.
// Fetch the public IP address and its location; the HTTP client is injected so tests can stub it.
// @@@SNIPSTART java-ipgeo-activity-setup
package iplocate;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.temporal.activity.Activity;
import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;

public class IPActivitiesImpl implements IPActivities {
    private final HttpClient httpClient;
    private final ObjectMapper objectMapper = new ObjectMapper();

    public IPActivitiesImpl(HttpClient httpClient) {
        this.httpClient = httpClient;
    }
// @@@SNIPEND

// @@@SNIPSTART java-ipgeo-activity-ip
    // getIP fetches the public IP address.
    @Override
    public String getIP() {
        HttpRequest request = HttpRequest.newBuilder(URI.create("https://icanhazip.com")).build();
        try {
            HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
            return response.body().trim();
        } catch (IOException | InterruptedException e) {
            throw Activity.wrap(e);
        }
    }
// @@@SNIPEND

// @@@SNIPSTART java-ipgeo-activity-location
    // getLocationInfo uses the IP address to fetch location information.
    @Override
    public String getLocationInfo(String ip) {
        HttpRequest request = HttpRequest.newBuilder(URI.create("http://ip-api.com/json/" + ip)).build();
        try {
            HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
            JsonNode data = objectMapper.readTree(response.body());
            return String.format("%s, %s, %s",
                    data.get("city").asText(),
                    data.get("regionName").asText(),
                    data.get("country").asText());
        } catch (IOException | InterruptedException e) {
            throw Activity.wrap(e);
        }
    }
}
// @@@SNIPEND
