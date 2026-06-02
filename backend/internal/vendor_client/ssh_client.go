package vendor_client

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHResult struct {
	Output string
	Err    error
}

func sshConnect(host string, port int, username, password string, timeout time.Duration) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
			ssh.KeyboardInteractive(func(name, instruction string, questions []string, echos []bool) ([]string, error) {
				// Some devices use keyboard-interactive instead of password auth
				answers := make([]string, len(questions))
				for i := range questions {
					answers[i] = password
				}
				return answers, nil
			}),
		},
		// Network devices (routers, firewalls) use self-signed SSH keys with no PKI; host key verification is not feasible.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
		// Allow legacy algorithms — many older network devices (Cisco IOS, older Juniper, legacy switches) still require them.
		Config: ssh.Config{
			KeyExchanges: []string{
				"curve25519-sha256", "curve25519-sha256@libssh.org",
				"ecdh-sha2-nistp256", "ecdh-sha2-nistp384", "ecdh-sha2-nistp521",
				"diffie-hellman-group-exchange-sha256",
				"diffie-hellman-group14-sha256", "diffie-hellman-group16-sha512",
				// Legacy
				"diffie-hellman-group14-sha1", "diffie-hellman-group1-sha1",
				"diffie-hellman-group-exchange-sha1",
			},
			Ciphers: []string{
				"aes128-gcm@openssh.com", "aes256-gcm@openssh.com",
				"chacha20-poly1305@openssh.com",
				"aes128-ctr", "aes192-ctr", "aes256-ctr",
				// Legacy
				"aes128-cbc", "aes192-cbc", "aes256-cbc", "3des-cbc",
			},
			MACs: []string{
				"hmac-sha2-256-etm@openssh.com", "hmac-sha2-512-etm@openssh.com",
				"hmac-sha2-256", "hmac-sha2-512",
				// Legacy
				"hmac-sha1", "hmac-sha1-96",
			},
		},
		HostKeyAlgorithms: []string{
			"ssh-ed25519",
			"ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521",
			"rsa-sha2-256", "rsa-sha2-512",
			// Legacy
			"ssh-rsa", "ssh-dss",
		},
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	return ssh.Dial("tcp", addr, config)
}

func sshRunCommand(client *ssh.Client, command string, readTimeout time.Duration) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdout bytes.Buffer
	session.Stdout = &stdout

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	session.RequestPty("xterm", 80, 200, modes)

	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return "", err
	}

	if err := session.Shell(); err != nil {
		return "", err
	}

	time.Sleep(500 * time.Millisecond)

	fmt.Fprintf(stdinPipe, "%s\n", command)
	time.Sleep(500 * time.Millisecond)
	fmt.Fprintf(stdinPipe, "exit\n")

	done := make(chan error, 1)
	go func() { done <- session.Wait() }()

	select {
	case <-done:
	case <-time.After(readTimeout):
	}

	return stdout.String(), nil
}

func sshRunInteractive(client *ssh.Client, commands []string, readTimeout time.Duration) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	session.RequestPty("xterm", 80, 500, modes)

	stdinPipe, _ := session.StdinPipe()
	stdoutPipe, _ := session.StdoutPipe()

	if err := session.Shell(); err != nil {
		return "", err
	}

	time.Sleep(1 * time.Second)

	for _, cmd := range commands {
		fmt.Fprintf(stdinPipe, "%s\n", cmd)
		time.Sleep(500 * time.Millisecond)
	}

	time.Sleep(2 * time.Second)
	fmt.Fprintf(stdinPipe, "exit\n")

	done := make(chan struct{})
	var output bytes.Buffer
	go func() {
		io.Copy(&output, stdoutPipe)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(readTimeout):
	}

	return output.String(), nil
}

func cleanSSHOutput(output, command string) string {
	lines := strings.Split(output, "\n")
	var result []string
	started := false
	for _, line := range lines {
		if !started && strings.Contains(line, command) {
			started = true
			continue
		}
		if started {
			if strings.Contains(line, "exit") || strings.Contains(line, "logout") {
				break
			}
			result = append(result, line)
		}
	}
	if len(result) == 0 {
		return output
	}
	return strings.Join(result, "\n")
}
