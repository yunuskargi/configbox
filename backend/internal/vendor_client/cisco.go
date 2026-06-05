package vendor_client

import (
	"fmt"
	"time"
)

var backupCommands = map[string]string{
	"ios":  "show running-config",
	"nxos": "show running-config",
	"asa":  "show running-config",
}

var testCommands = map[string]string{
	"ios":  "show version",
	"nxos": "show version",
	"asa":  "show version",
}

func FetchCiscoConfig(ip string, port int, username, password, enablePassword, platform string) (string, error) {
	if platform == "" {
		platform = "ios"
	}

	commands := []string{"terminal length 0"}
	if enablePassword != "" {
		commands = append(commands, "enable", enablePassword)
	}

	cmd := backupCommands[platform]
	if cmd == "" {
		cmd = "show running-config"
	}
	commands = append(commands, cmd)

	output, err := runSSHCommands(ip, port, username, password, commands, 30*time.Second, 60*time.Second)
	if err != nil {
		return "", fmt.Errorf("SSH connection failed: %v", err)
	}

	cleaned := cleanSSHOutput(output, cmd)
	return cleaned, nil
}

func TestCisco(ip string, port int, username, password, enablePassword, platform string) error {
	if platform == "" {
		platform = "ios"
	}

	commands := []string{"terminal length 0"}
	if enablePassword != "" {
		commands = append(commands, "enable", enablePassword)
	}

	cmd := testCommands[platform]
	if cmd == "" {
		cmd = "show version"
	}
	commands = append(commands, cmd)

	_, err := runSSHCommands(ip, port, username, password, commands, 10*time.Second, 15*time.Second)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %v", err)
	}
	return nil
}
