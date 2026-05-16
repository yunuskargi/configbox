package vendor_client

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func FetchPaloAltoConfig(ip string, port int, token string) (string, error) {
	url := fmt.Sprintf("https://%s:%d/api/", ip, port)

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	req, _ := http.NewRequest("GET", url, nil)
	q := req.URL.Query()
	q.Set("type", "export")
	q.Set("category", "configuration")
	q.Set("key", token)
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

	text := string(body)
	if strings.Contains(text, "<response") && strings.Contains(text, `status="error"`) {
		return "", fmt.Errorf("PAN-OS API error: %s", text[:min(200, len(text))])
	}
	return text, nil
}

func TestPaloAlto(ip string, port int, token string) error {
	url := fmt.Sprintf("https://%s:%d/api/", ip, port)

	client := &http.Client{
		Timeout:   15 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	req, _ := http.NewRequest("GET", url, nil)
	q := req.URL.Query()
	q.Set("type", "op")
	q.Set("cmd", "<show><system><info></info></system></show>")
	q.Set("key", token)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(body), `status="error"`) {
		return fmt.Errorf("PAN-OS API error")
	}
	return nil
}
