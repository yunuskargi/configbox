package vendor_client

import (
	"fmt"
	"time"
)

func FetchDellConfig(ip string, port int, username, password, enablePassword string) (string, error) {
	// Dell PowerConnect requires priv 15 for show running-config.
	// Send enable always — if no enable password is set on device, elevation is immediate.
	commands := []string{"enable"}
	// Send the enable password (or an empty line if none) so any "Password:" prompt is consumed safely.
	commands = append(commands, enablePassword)
	commands = append(commands, "terminal length 0", "show running-config")

	output, err := runSSHCommands(ip, port, username, password, commands, 30*time.Second, 150*time.Second)
	if err != nil {
		return "", fmt.Errorf("SSH connection failed: %v", err)
	}

	cleaned := cleanSSHOutput(output, "show running-config")
	return cleaned, nil
}

func TestDell(ip string, port int, username, password, enablePassword string) error {
	commands := []string{"enable"}
	if enablePassword != "" {
		commands = append(commands, enablePassword)
	}
	commands = append(commands, "show version")

	_, err := runSSHCommands(ip, port, username, password, commands, 15*time.Second, 15*time.Second)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %v", err)
	}
	return nil
}
