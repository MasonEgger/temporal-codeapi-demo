# ABOUTME: Activities for the Python IP geolocation tutorial, mirroring the Go and TypeScript tutorials.
# Fetch the public IP address and its location; tests patch requests.get so they never touch the network.
# @@@SNIPSTART python-ipgeo-activity-setup
import requests
from temporalio import activity

# @@@SNIPEND


# @@@SNIPSTART python-ipgeo-activity-ip
# Get the IP address
@activity.defn
def get_ip() -> str:
    response = requests.get("https://icanhazip.com")
    response.raise_for_status()
    return response.text.strip()


# @@@SNIPEND


# @@@SNIPSTART python-ipgeo-activity-location
# Use the IP address to get the location.
@activity.defn
def get_location_info(ip: str) -> str:
    response = requests.get(f"http://ip-api.com/json/{ip}")
    response.raise_for_status()
    result = response.json()
    return f"{result['city']}, {result['regionName']}, {result['country']}"


# @@@SNIPEND
