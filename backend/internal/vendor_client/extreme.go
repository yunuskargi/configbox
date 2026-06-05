package vendor_client

import (
	"fmt"
	"time"
)

// FetchExtremeConfig fetches running-config from Extreme Networks SLX-OS devices.
// SLX-OS is the evolution of Brocade NOS (Extreme acquired Brocade's DC business in 2017).
func FetchExtremeConfig(ip string, port int, username, password, enablePassword string) (string, error) {
	commands := []string{"terminal length 0"}
	if enablePassword != "" {
		commands = append(commands, "enable", enablePassword)
	}
	commands = append(commands, "show running-config | nomore")

	output, err := runSSHCommands(ip, port, username, password, commands, 60*time.Second, 120*time.Second)
	if err != nil {
		return "", fmt.Errorf("SSH connection failed: %v", err)
	}

	cleaned := cleanSSHOutput(output, "show running-config")
	return cleaned, nil
}

func TestExtreme(ip string, port int, username, password, enablePassword string) error {
	commands := []string{"terminal length 0"}
	if enablePassword != "" {
		commands = append(commands, "enable", enablePassword)
	}
	commands = append(commands, "show version")

	_, err := runSSHCommands(ip, port, username, password, commands, 30*time.Second, 15*time.Second)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %v", err)
	}
	return nil
}
