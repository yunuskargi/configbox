import requests
import urllib3

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


def _base_url(device):
    return f"https://{device.ip_address}:{device.port}"


def fetch_paloalto_config(device) -> str:
    url = f"{_base_url(device)}/api/"
    params = {
        "type": "export",
        "category": "configuration",
        "key": device.auth_token,
    }
    resp = requests.get(url, params=params, verify=False, timeout=30)
    resp.raise_for_status()
    if "<response" in resp.text and 'status="error"' in resp.text:
        raise Exception(f"PAN-OS API error: {resp.text[:200]}")
    return resp.text


def test_paloalto(device):
    url = f"{_base_url(device)}/api/"
    params = {
        "type": "op",
        "cmd": "<show><system><info></info></system></show>",
        "key": device.auth_token,
    }
    resp = requests.get(url, params=params, verify=False, timeout=15)
    resp.raise_for_status()
    if 'status="error"' in resp.text:
        raise Exception(f"PAN-OS API error: {resp.text[:200]}")
    return True
