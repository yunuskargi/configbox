package vendor_client

import (
	"fmt"
	"time"
)

func FetchBrocadeConfig(ip string, port int, username, password, enablePassword string) (string, error) {
	// Brocade NOS uses role-based access (admin/user) — no enable mode.
	commands := []string{"terminal length 0", "skip-page-display", "show running-config | nomore"}
	_ = enablePassword

	output, err := runSSHCommands(ip, port, username, password, commands, 60*time.Second, 120*time.Second)
	if err != nil {
		return "", fmt.Errorf("SSH connection failed: %v", err)
	}

	cleaned := cleanSSHOutput(output, "show running-config")
	return cleaned, nil
}

func TestBrocade(ip string, port int, username, password, enablePassword string) error {
	commands := []string{"terminal length 0", "show version"}
	_ = enablePassword

	_, err := runSSHCommands(ip, port, username, password, commands, 30*time.Second, 15*time.Second)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %v", err)
	}
	return nil
}
