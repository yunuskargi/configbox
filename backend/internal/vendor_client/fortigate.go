package vendor_client

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

func FetchFortigateConfig(ip string, port int, token string, vdom string) (string, error) {
	url := fmt.Sprintf("https://%s:%d/api/v2/monitor/system/config/backup", ip, port)

	client := &http.Client{
		Timeout:   30 * time.Second,
		// FortiGate devices use self-signed certificates by default.
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	q := req.URL.Query()
	if vdom != "" {
		q.Set("scope", "vdom")
		q.Set("vdom", vdom)
	} else {
		q.Set("scope", "global")
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func TestFortigate(ip string, port int, token string, vdom string) error {
	url := fmt.Sprintf("https://%s:%d/api/v2/monitor/system/status", ip, port)

	client := &http.Client{
		Timeout:   10 * time.Second,
		// FortiGate devices use self-signed certificates by default.
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	if vdom != "" {
		q := req.URL.Query()
		q.Set("vdom", vdom)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}
