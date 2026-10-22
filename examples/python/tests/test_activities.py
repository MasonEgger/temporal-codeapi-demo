# ABOUTME: Activity tests for the Python IP geolocation tutorial, mirroring the Go and TypeScript tutorials.
# Patches requests.get so the tests never touch the network.
# @@@SNIPSTART python-ipgeo-activity-test-setup
from unittest.mock import MagicMock, patch

from temporalio.testing import ActivityEnvironment

from activities import get_ip, get_location_info

# @@@SNIPEND


# @@@SNIPSTART python-ipgeo-activity-test-ip
def test_get_ip() -> None:
    fake_response = MagicMock()
    fake_response.text = "127.0.0.1\n"

    with patch("activities.requests.get", return_value=fake_response):
        ip = ActivityEnvironment().run(get_ip)

    assert ip == "127.0.0.1"


# @@@SNIPEND


# @@@SNIPSTART python-ipgeo-activity-test-location
def test_get_location_info() -> None:
    fake_response = MagicMock()
    fake_response.json.return_value = {
        "city": "San Francisco",
        "regionName": "California",
        "country": "United States",
    }

    with patch("activities.requests.get", return_value=fake_response):
        location = ActivityEnvironment().run(get_location_info, "127.0.0.1")

    assert location == "San Francisco, California, United States"


# @@@SNIPEND
