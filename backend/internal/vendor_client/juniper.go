package vendor_client

import (
	"fmt"
	"time"
)

func FetchJuniperConfig(ip string, port int, username, password string) (string, error) {
	client, err := sshConnect(ip, port, username, password, 30*time.Second)
	if err != nil {
		return "", fmt.Errorf("SSH connection failed: %v", err)
	}
	defer client.Close()

	output, err := sshRunCommand(client, "show configuration | display set", 60*time.Second)
	if err != nil {
		return "", fmt.Errorf("command execution failed: %v", err)
	}

	cleaned := cleanSSHOutput(output, "show configuration")
	return cleaned, nil
}

func TestJuniper(ip string, port int, username, password string) error {
	client, err := sshConnect(ip, port, username, password, 10*time.Second)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %v", err)
	}
	defer client.Close()

	_, err = sshRunCommand(client, "show version", 15*time.Second)
	if err != nil {
		return fmt.Errorf("command execution failed: %v", err)
	}
	return nil
}
