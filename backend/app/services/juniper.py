from netmiko import ConnectHandler


def fetch_juniper_config(device) -> str:
    conn = ConnectHandler(
        device_type="juniper_junos",
        host=device.ip_address,
        port=device.port,
        username=device.ssh_username,
        password=device.ssh_password,
        timeout=30,
    )
    output = conn.send_command("show configuration | display set", read_timeout=60)
    conn.disconnect()
    return output


def test_juniper(device):
    conn = ConnectHandler(
        device_type="juniper_junos",
        host=device.ip_address,
        port=device.port,
        username=device.ssh_username,
        password=device.ssh_password,
        timeout=10,
    )
    output = conn.send_command("show version", read_timeout=15)
    conn.disconnect()
    return output
