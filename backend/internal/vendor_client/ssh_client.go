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
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
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
