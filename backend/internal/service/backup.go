package service

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yunuskargi/confbox/internal/config"
	"github.com/yunuskargi/confbox/internal/crypto"
	"github.com/yunuskargi/confbox/internal/database"
	vc "github.com/yunuskargi/confbox/internal/vendor_client"
)

type DeviceInfo struct {
	ID             int
	Name           string
	Vendor         string
	IPAddress      string
	Port           int
	Vdom           string
	AuthToken      string
	SSHUsername    string
	SSHPassword    string
	EnablePassword string
	Platform       string
	LocationName   string
}

func loadDevice(deviceID int) DeviceInfo {
	var d struct {
		ID             int     `db:"id"`
		Name           string  `db:"name"`
		Vendor         string  `db:"vendor"`
		IPAddress      string  `db:"ip_address"`
		Port           int     `db:"port"`
		Vdom           *string `db:"vdom"`
		AuthToken      *string `db:"auth_token"`
		SSHUsername    *string `db:"ssh_username"`
		SSHPassword    *string `db:"ssh_password"`
		EnablePassword *string `db:"enable_password"`
		Platform       *string `db:"platform"`
		LocationID     *int    `db:"location_id"`
	}
	database.DB.Get(&d, "SELECT id, name, vendor, ip_address, port, vdom, auth_token, ssh_username, ssh_password, enable_password, platform, location_id FROM devices WHERE id = ?", deviceID)

	info := DeviceInfo{
		ID: d.ID, Name: d.Name, Vendor: d.Vendor, IPAddress: d.IPAddress, Port: d.Port,
	}
	if d.Vdom != nil {
		info.Vdom = *d.Vdom
	}
	if d.AuthToken != nil {
		info.AuthToken = crypto.Decrypt(*d.AuthToken)
	}
	if d.SSHUsername != nil {
		info.SSHUsername = *d.SSHUsername
	}
	if d.SSHPassword != nil {
		info.SSHPassword = crypto.Decrypt(*d.SSHPassword)
	}
	if d.EnablePassword != nil {
		info.EnablePassword = crypto.Decrypt(*d.EnablePassword)
	}
	if d.Platform != nil {
		info.Platform = *d.Platform
	}
	if d.LocationID != nil {
		database.DB.Get(&info.LocationName, "SELECT name FROM locations WHERE id = ?", *d.LocationID)
	}
	return info
}

func sanitizeName(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "..", "")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	if name == "" || name == "." {
		name = "unknown"
	}
	return name
}

func normalizeConfig(content, vendor string) string {
	var lines []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		skip := false
		switch vendor {
		case "fortigate":
			if strings.HasPrefix(trimmed, "#") {
				skip = true
			}
		case "cisco":
			if strings.HasPrefix(trimmed, "! Last configuration change") ||
				strings.HasPrefix(trimmed, "! NVRAM config last updated") ||
				strings.HasPrefix(trimmed, "ntp clock-period") ||
				strings.HasPrefix(trimmed, "! No configuration change since") {
				skip = true
			}
		case "juniper":
			if strings.HasPrefix(trimmed, "## Last commit:") ||
				strings.HasPrefix(trimmed, "## Last changed:") {
				skip = true
			}
		}
		if !skip {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func RunBackup(deviceID int, triggeredBy string) map[string]any {
	device := loadDevice(deviceID)
	timestamp := time.Now().In(config.AppTimezone).Format("2006-01-02_150405")
	deviceDir := filepath.Join(config.BackupDir, sanitizeName(device.Vendor), sanitizeName(device.Name))
	os.MkdirAll(deviceDir, 0700)

	filePath := filepath.Join(deviceDir, timestamp+".conf")

	var configContent string
	var fetchErr error

	switch device.Vendor {
	case "fortigate":
		configContent, fetchErr = vc.FetchFortigateConfig(device.IPAddress, device.Port, device.AuthToken, device.Vdom)
	case "juniper":
		configContent, fetchErr = vc.FetchJuniperConfig(device.IPAddress, device.Port, device.SSHUsername, device.SSHPassword)
	case "cisco":
		configContent, fetchErr = vc.FetchCiscoConfig(device.IPAddress, device.Port, device.SSHUsername, device.SSHPassword, device.EnablePassword, device.Platform)
	case "paloalto":
		configContent, fetchErr = vc.FetchPaloAltoConfig(device.IPAddress, device.Port, device.AuthToken)
	default:
		fetchErr = fmt.Errorf("unsupported vendor: %s", device.Vendor)
	}

	if fetchErr != nil {
		database.DB.Exec(`INSERT INTO backups (device_id, file_path, file_size, status, error_message, triggered_by, created_at)
			VALUES (?, ?, 0, 'failed', ?, ?, ?)`, device.ID, filePath, fetchErr.Error(), triggeredBy, config.Now())

		go func() {
			NotifyBackup(device.Name, device.Vendor, "failed", fetchErr.Error(), "", 0, device.LocationName, device.Vdom, triggeredBy)
		}()

		return map[string]any{"status": "failed", "error": "Backup failed. Check backup history for details."}
	}

	if err := os.WriteFile(filePath, []byte(configContent), 0600); err != nil {
		slog.Error("failed to write backup file", "path", filePath, "error", err)
		database.DB.Exec(`INSERT INTO backups (device_id, file_path, file_size, status, error_message, triggered_by, created_at)
			VALUES (?, ?, 0, 'failed', ?, ?, ?)`, device.ID, filePath, "Failed to write backup file", triggeredBy, config.Now())
		return map[string]any{"status": "failed", "error": "Backup failed. Check backup history for details."}
	}
	fi, _ := os.Stat(filePath)
	fileSize := int(fi.Size())

	configChanged := false
	previousConfig := getPreviousConfig(deviceDir)
	if previousConfig != "" {
		configChanged = normalizeConfig(previousConfig, device.Vendor) != normalizeConfig(configContent, device.Vendor)
	}

	database.DB.Exec(`INSERT INTO backups (device_id, file_path, file_size, status, triggered_by, created_at)
		VALUES (?, ?, ?, 'success', ?, ?)`, device.ID, filePath, fileSize, triggeredBy, config.Now())

	go func() {
		NotifyBackup(device.Name, device.Vendor, "success", "", filePath, fileSize, device.LocationName, device.Vdom, triggeredBy)
		if configChanged {
			NotifyConfigChange(device.Name, device.Vendor, device.LocationName, device.Vdom)
		}
	}()

	return map[string]any{"status": "success", "file_path": filePath, "file_size": fileSize, "config_changed": configChanged}
}

func TestDeviceConnection(deviceID int) map[string]any {
	device := loadDevice(deviceID)
	var err error

	switch device.Vendor {
	case "fortigate":
		err = vc.TestFortigate(device.IPAddress, device.Port, device.AuthToken, device.Vdom)
	case "juniper":
		err = vc.TestJuniper(device.IPAddress, device.Port, device.SSHUsername, device.SSHPassword)
	case "cisco":
		err = vc.TestCisco(device.IPAddress, device.Port, device.SSHUsername, device.SSHPassword, device.EnablePassword, device.Platform)
	case "paloalto":
		err = vc.TestPaloAlto(device.IPAddress, device.Port, device.AuthToken)
	default:
		err = fmt.Errorf("unsupported vendor: %s", device.Vendor)
	}

	if err != nil {
		return map[string]any{"status": "failed", "message": err.Error()}
	}
	return map[string]any{"status": "success", "message": "Connection successful"}
}

func DeleteDeviceBackupFiles(vendor, name string) {
	deviceDir := filepath.Join(config.BackupDir, sanitizeName(vendor), sanitizeName(name))
	abs, _ := filepath.Abs(deviceDir)
	backupAbs, _ := filepath.Abs(config.BackupDir)
	if !strings.HasPrefix(abs, backupAbs) {
		return
	}
	os.RemoveAll(deviceDir)
}

func getPreviousConfig(deviceDir string) string {
	entries, err := os.ReadDir(deviceDir)
	if err != nil {
		return ""
	}

	var confs []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".conf") {
			confs = append(confs, filepath.Join(deviceDir, e.Name()))
		}
	}
	if len(confs) == 0 {
		return ""
	}

	sort.Slice(confs, func(i, j int) bool {
		fi, _ := os.Stat(confs[i])
		fj, _ := os.Stat(confs[j])
		return fi.ModTime().After(fj.ModTime())
	})

	content, err := os.ReadFile(confs[0])
	if err != nil {
		return ""
	}
	return string(content)
}

func RunScheduledBackup(deviceID int) {
	var isActive bool
	err := database.DB.Get(&isActive, "SELECT is_active FROM devices WHERE id = ?", deviceID)
	if err != nil || !isActive {
		return
	}
	result := RunBackup(deviceID, "scheduled")
	slog.Info("scheduled backup", "device_id", deviceID, "status", result["status"])
}
