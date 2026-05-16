import requests
import urllib3

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


def _headers(token: str) -> dict:
    return {"Authorization": f"Bearer {token}"}


def _backup_params(device) -> dict:
    if device.vdom:
        return {"scope": "vdom", "vdom": device.vdom}
    return {"scope": "global"}


def fetch_fortigate_config(device) -> str:
    url = f"https://{device.ip_address}:{device.port}/api/v2/monitor/system/config/backup"
    resp = requests.get(url, headers=_headers(device.auth_token), params=_backup_params(device), verify=False, timeout=30)
    resp.raise_for_status()
    return resp.text


def test_fortigate(device):
    url = f"https://{device.ip_address}:{device.port}/api/v2/monitor/system/status"
    params = {"vdom": device.vdom} if device.vdom else {}
    resp = requests.get(url, headers=_headers(device.auth_token), params=params, verify=False, timeout=10)
    resp.raise_for_status()
    return resp.json()
