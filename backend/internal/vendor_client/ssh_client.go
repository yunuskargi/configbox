package vendor_client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
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
			// Some Juniper devices (QFX, EX series) only accept keyboard-interactive auth, not plain password.
			ssh.KeyboardInteractive(func(name, instruction string, questions []string, echos []bool) ([]string, error) {
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
		// Allow legacy algorithms — older network devices (OpenSSH 6.x, legacy Juniper/Cisco) need them.
		Config: ssh.Config{
			KeyExchanges: []string{
				"curve25519-sha256", "curve25519-sha256@libssh.org",
				"ecdh-sha2-nistp256", "ecdh-sha2-nistp384", "ecdh-sha2-nistp521",
				"diffie-hellman-group-exchange-sha256",
				"diffie-hellman-group14-sha256", "diffie-hellman-group16-sha512",
				"diffie-hellman-group14-sha1", "diffie-hellman-group1-sha1",
				"diffie-hellman-group-exchange-sha1",
			},
			Ciphers: []string{
				"aes128-gcm@openssh.com", "aes256-gcm@openssh.com",
				"chacha20-poly1305@openssh.com",
				"aes128-ctr", "aes192-ctr", "aes256-ctr",
				"aes128-cbc", "3des-cbc",
			},
			MACs: []string{
				"hmac-sha2-256-etm@openssh.com", "hmac-sha2-512-etm@openssh.com",
				"hmac-sha2-256", "hmac-sha2-512",
				"hmac-sha1", "hmac-sha1-96",
			},
		},
		HostKeyAlgorithms: []string{
			"ssh-ed25519",
			"ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521",
			"rsa-sha2-256", "rsa-sha2-512",
			"ssh-rsa", "ssh-dss",
		},
		// Some old SSH servers reject the default "SSH-2.0-Go" version string.
		ClientVersion: "SSH-2.0-OpenSSH_8.9",
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

	// Use a mutex-protected buffer so we can safely read its length while the
	// reader goroutine is writing to it (bytes.Buffer is NOT goroutine-safe).
	output := &safeBuffer{}
	done := make(chan struct{})
	go func() {
		io.Copy(output, stdoutPipe)
		close(done)
	}()

	time.Sleep(1 * time.Second)

	for _, cmd := range commands {
		fmt.Fprintf(stdinPipe, "%s\n", cmd)
		time.Sleep(500 * time.Millisecond)
	}

	// Wait for output to settle (no new bytes for 3 seconds) before sending exit.
	// This is critical for commands like "show running-config" that stream output
	// for tens of seconds — sending exit too early would cut off the output.
	settleDeadline := time.Now().Add(readTimeout)
	prevLen := -1
	stableTicks := 0
	for {
		if time.Now().After(settleDeadline) {
			break
		}
		time.Sleep(1 * time.Second)
		curLen := output.Len()
		if curLen == prevLen {
			stableTicks++
			if stableTicks >= 3 {
				break
			}
		} else {
			stableTicks = 0
			prevLen = curLen
		}
	}

	fmt.Fprintf(stdinPipe, "exit\n")

	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}

	return output.String(), nil
}

// safeBuffer is a goroutine-safe wrapper around bytes.Buffer for concurrent
// read/write access in sshRunInteractive.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Len()
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
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

// runSSHCommands runs interactive SSH commands and returns combined output.
// Tries Go's crypto/ssh first; if that fails (typically due to legacy SSH
// server bugs like the OpenSSH 6.x preauth crash), falls back to the system
// openssh client invoked via sshpass.
func runSSHCommands(host string, port int, username, password string, commands []string, connectTimeout, readTimeout time.Duration) (string, error) {
	client, err := sshConnect(host, port, username, password, connectTimeout)
	if err == nil {
		defer client.Close()
		return sshRunInteractive(client, commands, readTimeout)
	}

	slog.Info("ssh: Go client failed, attempting system openssh fallback", "host", host, "error", err.Error())
	output, fbErr := sshFallbackInteractive(host, port, username, password, commands, connectTimeout+readTimeout)
	if fbErr != nil {
		// Return the original Go SSH error since fallback also failed.
		return "", fmt.Errorf("%v (fallback also failed: %v)", err, fbErr)
	}
	return output, nil
}

// sshFallbackInteractive shells out to system openssh via sshpass for legacy
// devices that Go's crypto/ssh cannot interoperate with (e.g. OpenSSH 6.x
// servers that crash on modern KEXINIT extensions).
func sshFallbackInteractive(host string, port int, username, password string, commands []string, timeout time.Duration) (string, error) {
	args := []string{
		"-e", "ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ConnectTimeout=30",
		"-o", "KexAlgorithms=+diffie-hellman-group1-sha1,diffie-hellman-group14-sha1,diffie-hellman-group-exchange-sha1",
		"-o", "HostKeyAlgorithms=+ssh-rsa,ssh-dss",
		"-o", "PubkeyAuthentication=no",
		"-o", "PreferredAuthentications=password,keyboard-interactive",
		"-p", strconv.Itoa(port),
		fmt.Sprintf("%s@%s", username, host),
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sshpass", args...)
	// Inherit parent environment so PATH/HOME/etc. are available to sshpass and ssh.
	// SSHPASS is appended (not the only var) and read by sshpass with -e flag.
	cmd.Env = append(os.Environ(), "SSHPASS="+password)

	var stdin bytes.Buffer
	for _, c := range commands {
		stdin.WriteString(c + "\n")
	}
	stdin.WriteString("exit\n")
	cmd.Stdin = &stdin

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	out := stdout.String()
	// Some network devices return a non-zero exit code even on success — accept
	// the output if we got something substantial.
	if err != nil && len(out) < 50 {
		return "", fmt.Errorf("system ssh failed: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}
