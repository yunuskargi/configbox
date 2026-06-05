package vendor_client

import (
	"fmt"
	"time"
)

func FetchDellConfig(ip string, port int, username, password, enablePassword string) (string, error) {
	commands := []string{}
	if enablePassword != "" {
		commands = append(commands, "enable", enablePassword)
	}
	commands = append(commands, "terminal length 0", "show running-config")

	output, err := runSSHCommands(ip, port, username, password, commands, 30*time.Second, 90*time.Second)
	if err != nil {
		return "", fmt.Errorf("SSH connection failed: %v", err)
	}

	cleaned := cleanSSHOutput(output, "show running-config")
	return cleaned, nil
}

func TestDell(ip string, port int, username, password, enablePassword string) error {
	commands := []string{}
	if enablePassword != "" {
		commands = append(commands, "enable", enablePassword)
	}
	commands = append(commands, "show version")

	_, err := runSSHCommands(ip, port, username, password, commands, 15*time.Second, 15*time.Second)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %v", err)
	}
	return nil
}
