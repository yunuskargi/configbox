from netmiko import ConnectHandler

PLATFORM_MAP = {
    "ios": "cisco_ios",
    "nxos": "cisco_nxos",
    "asa": "cisco_asa",
}

BACKUP_COMMANDS = {
    "ios": "show running-config",
    "nxos": "show running-config",
    "asa": "show running-config",
}

TEST_COMMANDS = {
    "ios": "show version",
    "nxos": "show version",
    "asa": "show version",
}


def _get_device_type(device) -> str:
    platform = getattr(device, "platform", "ios") or "ios"
    return PLATFORM_MAP.get(platform, "cisco_ios")


def _connect(device, timeout=30):
    device_type = _get_device_type(device)
    params = {
        "device_type": device_type,
        "host": device.ip_address,
        "port": device.port,
        "username": device.ssh_username,
        "password": device.ssh_password,
        "timeout": timeout,
    }
    if device.enable_password:
        params["secret"] = device.enable_password
    conn = ConnectHandler(**params)
    conn.enable()
    return conn


def fetch_cisco_config(device) -> str:
    conn = _connect(device, timeout=30)
    platform = getattr(device, "platform", "ios") or "ios"
    cmd = BACKUP_COMMANDS.get(platform, "show running-config")
    output = conn.send_command(cmd, read_timeout=60)
    conn.disconnect()
    return output


def test_cisco(device):
    conn = _connect(device, timeout=10)
    platform = getattr(device, "platform", "ios") or "ios"
    cmd = TEST_COMMANDS.get(platform, "show version")
    output = conn.send_command(cmd, read_timeout=15)
    conn.disconnect()
    return output
