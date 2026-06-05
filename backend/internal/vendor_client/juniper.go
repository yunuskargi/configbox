package vendor_client

import (
	"fmt"
	"time"
)

func FetchJuniperConfig(ip string, port int, username, password string) (string, error) {
	commands := []string{"show configuration | display set | no-more"}
	output, err := runSSHCommands(ip, port, username, password, commands, 30*time.Second, 60*time.Second)
	if err != nil {
		return "", fmt.Errorf("SSH connection failed: %v", err)
	}
	cleaned := cleanSSHOutput(output, "show configuration")
	return cleaned, nil
}

func TestJuniper(ip string, port int, username, password string) error {
	commands := []string{"show version"}
	_, err := runSSHCommands(ip, port, username, password, commands, 10*time.Second, 15*time.Second)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %v", err)
	}
	return nil
}
