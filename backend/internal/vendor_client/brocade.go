package vendor_client

import (
	"fmt"
	"time"
)

func FetchBrocadeConfig(ip string, port int, username, password, enablePassword string) (string, error) {
	client, err := sshConnect(ip, port, username, password, 60*time.Second)
	if err != nil {
		return "", fmt.Errorf("SSH connection failed: %v", err)
	}
	defer client.Close()

	// Brocade NOS/FastIron/NetIron all accept "terminal length 0" to disable paging.
	commands := []string{"terminal length 0"}
	if enablePassword != "" {
		commands = append(commands, "enable", enablePassword)
	}
	commands = append(commands, "show running-config")

	output, err := sshRunInteractive(client, commands, 60*time.Second)
	if err != nil {
		return "", fmt.Errorf("command execution failed: %v", err)
	}

	cleaned := cleanSSHOutput(output, "show running-config")
	return cleaned, nil
}

func TestBrocade(ip string, port int, username, password, enablePassword string) error {
	client, err := sshConnect(ip, port, username, password, 30*time.Second)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %v", err)
	}
	defer client.Close()

	commands := []string{"terminal length 0"}
	if enablePassword != "" {
		commands = append(commands, "enable", enablePassword)
	}
	commands = append(commands, "show version")

	_, err = sshRunInteractive(client, commands, 15*time.Second)
	if err != nil {
		return fmt.Errorf("command execution failed: %v", err)
	}
	return nil
}
