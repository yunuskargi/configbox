package handler

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/yunuskargi/confbox/internal/auth"
	"github.com/yunuskargi/confbox/internal/crypto"
	"github.com/yunuskargi/confbox/internal/database"
	"github.com/yunuskargi/confbox/internal/models"
	"github.com/yunuskargi/confbox/internal/service"
)

var encryptedFields = []string{"auth_token", "ssh_password", "enable_password"}

func enrichDevice(id int) models.DeviceOut {
	var out models.DeviceOut
	row := database.DB.QueryRow(`
		SELECT d.id, d.name, d.vendor, d.ip_address, d.port, d.location_id, l.name,
		       d.vdom, d.platform, d.schedule_cron, d.is_active, d.created_at, d.updated_at,
		       d.auth_token, d.ssh_password, d.enable_password
		FROM devices d LEFT JOIN locations l ON l.id = d.location_id
		WHERE d.id = ?`, id)

	var locID *int
	var locName, vdom, platform, cron *string
	var authToken, sshPass, enablePass *string
	row.Scan(&out.ID, &out.Name, &out.Vendor, &out.IPAddress, &out.Port,
		&locID, &locName, &vdom, &platform, &cron, &out.IsActive, &out.CreatedAt, &out.UpdatedAt,
		&authToken, &sshPass, &enablePass)
	out.LocationID = locID
	out.LocationName = locName
	out.Vdom = vdom
	out.Platform = platform
	out.ScheduleCron = cron
	out.HasToken = authToken != nil && *authToken != ""
	out.HasSSHPassword = sshPass != nil && *sshPass != ""
	out.HasEnablePassword = enablePass != nil && *enablePass != ""

	database.DB.Get(&out.BackupCount, "SELECT COUNT(*) FROM backups WHERE device_id = ? AND status = 'success'", id)
	database.DB.Get(&out.FailedCount, "SELECT COUNT(*) FROM backups WHERE device_id = ? AND status = 'failed'", id)

	var lastBackup *string
	database.DB.Get(&lastBackup, "SELECT created_at FROM backups WHERE device_id = ? AND status = 'success' ORDER BY created_at DESC LIMIT 1", id)
	out.LastBackup = lastBackup

	return out
}

func ListDevices(w http.ResponseWriter, r *http.Request) {
	var ids []int
	database.DB.Select(&ids, "SELECT id FROM devices ORDER BY name")
	result := make([]models.DeviceOut, len(ids))
	for i, id := range ids {
		result[i] = enrichDevice(id)
	}
	writeJSON(w, 200, result)
}

type deviceCreateBody struct {
	Name           string  `json:"name"`
	Vendor         string  `json:"vendor"`
	IPAddress      string  `json:"ip_address"`
	Port           int     `json:"port"`
	LocationID     *int    `json:"location_id"`
	Vdom           *string `json:"vdom"`
	AuthToken      *string `json:"auth_token"`
	SSHUsername    *string `json:"ssh_username"`
	SSHPassword    *string `json:"ssh_password"`
	EnablePassword *string `json:"enable_password"`
	Platform       *string `json:"platform"`
	ScheduleCron   *string `json:"schedule_cron"`
}

func CreateDevice(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	var body deviceCreateBody
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}

	if strings.ContainsAny(body.Name, "/\\..") || body.Name == "" {
		writeError(w, 400, "Invalid device name")
		return
	}

	validVendorsSet := map[string]bool{"fortigate": true, "juniper": true, "cisco": true, "paloalto": true}
	if !validVendorsSet[body.Vendor] {
		writeError(w, 400, "Invalid vendor")
		return
	}

	var exists int
	database.DB.Get(&exists, "SELECT COUNT(*) FROM devices WHERE name = ?", body.Name)
	if exists > 0 {
		writeError(w, 400, "Device name already exists")
		return
	}

	if body.Port == 0 {
		body.Port = 443
	}
	platform := "ios"
	if body.Platform != nil {
		platform = *body.Platform
	}

	authToken := encryptOptional(body.AuthToken)
	sshPass := encryptOptional(body.SSHPassword)
	enablePass := encryptOptional(body.EnablePassword)

	res, _ := database.DB.Exec(`INSERT INTO devices (name, vendor, ip_address, port, location_id, vdom, auth_token, ssh_username, ssh_password, enable_password, platform, schedule_cron, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, datetime('now'), datetime('now'))`,
		body.Name, body.Vendor, body.IPAddress, body.Port, body.LocationID, body.Vdom,
		authToken, body.SSHUsername, sshPass, enablePass, platform, body.ScheduleCron)

	id, _ := res.LastInsertId()
	uid := user.ID
	service.LogAction(&uid, user.Username, "create", "device", body.Name,
		fmt.Sprintf("Vendor: %s, IP: %s", body.Vendor, body.IPAddress), clientIP(r))

	if body.ScheduleCron != nil && *body.ScheduleCron != "" {
		service.ScheduleDevice(int(id), *body.ScheduleCron)
	}

	writeJSON(w, 201, enrichDevice(int(id)))
}

func GetDevice(w http.ResponseWriter, r *http.Request) {
	id := paramInt(r, "id")
	var exists int
	database.DB.Get(&exists, "SELECT COUNT(*) FROM devices WHERE id = ?", id)
	if exists == 0 {
		writeError(w, 404, "Device not found")
		return
	}
	writeJSON(w, 200, enrichDevice(id))
}

var allowedDeviceColumns = map[string]bool{
	"name": true, "ip_address": true, "port": true, "location_id": true,
	"vdom": true, "auth_token": true, "ssh_username": true, "ssh_password": true,
	"enable_password": true, "platform": true, "schedule_cron": true, "is_active": true,
}

func UpdateDevice(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	id := paramInt(r, "id")

	var exists int
	database.DB.Get(&exists, "SELECT COUNT(*) FROM devices WHERE id = ?", id)
	if exists == 0 {
		writeError(w, 404, "Device not found")
		return
	}

	var body map[string]any
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}

	for key, val := range body {
		if val == nil {
			continue
		}
		if !allowedDeviceColumns[key] {
			continue
		}
		switch key {
		case "auth_token", "ssh_password", "enable_password":
			if s, ok := val.(string); ok && s != "" {
				val = crypto.Encrypt(s)
			}
		case "port":
			if f, ok := val.(float64); ok {
				val = int(f)
			}
		case "location_id":
			if f, ok := val.(float64); ok {
				val = int(f)
			}
		case "is_active":
			if b, ok := val.(bool); ok {
				if b {
					val = 1
				} else {
					val = 0
				}
			}
		}
		database.DB.Exec(fmt.Sprintf("UPDATE devices SET %s = ? WHERE id = ?", key), val, id)
	}
	database.DB.Exec("UPDATE devices SET updated_at = datetime('now') WHERE id = ?", id)

	var name string
	database.DB.Get(&name, "SELECT name FROM devices WHERE id = ?", id)
	uid := user.ID
	service.LogAction(&uid, user.Username, "update", "device", name, "Device updated", clientIP(r))

	if cronVal, ok := body["schedule_cron"]; ok {
		if cronStr, ok := cronVal.(string); ok && cronStr != "" {
			service.ScheduleDevice(id, cronStr)
		} else {
			service.RemoveDeviceSchedule(id)
		}
	}

	writeJSON(w, 200, enrichDevice(id))
}

func DeleteDevice(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	id := paramInt(r, "id")
	keepBackups := r.URL.Query().Get("keep_backups") != "false"

	var device struct {
		Name   string `db:"name"`
		Vendor string `db:"vendor"`
	}
	err := database.DB.Get(&device, "SELECT name, vendor FROM devices WHERE id = ?", id)
	if err != nil {
		writeError(w, 404, "Device not found")
		return
	}

	if !keepBackups {
		service.DeleteDeviceBackupFiles(device.Vendor, device.Name)
	}

	service.RemoveDeviceSchedule(id)
	database.DB.Exec("DELETE FROM backups WHERE device_id = ?", id)
	database.DB.Exec("DELETE FROM devices WHERE id = ?", id)

	uid := user.ID
	service.LogAction(&uid, user.Username, "delete", "device", device.Name,
		fmt.Sprintf("Backups kept: %v", keepBackups), clientIP(r))
	writeJSON(w, 200, map[string]string{"message": "Device deleted"})
}

func TriggerBackup(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	id := paramInt(r, "id")

	var exists int
	database.DB.Get(&exists, "SELECT COUNT(*) FROM devices WHERE id = ?", id)
	if exists == 0 {
		writeError(w, 404, "Device not found")
		return
	}

	result := service.RunBackup(id, "manual")

	var name string
	database.DB.Get(&name, "SELECT name FROM devices WHERE id = ?", id)
	uid := user.ID
	service.LogAction(&uid, user.Username, "backup", "device", name,
		fmt.Sprintf("Result: %s", result["status"]), clientIP(r))

	writeJSON(w, 200, result)
}

func TestConnection(w http.ResponseWriter, r *http.Request) {
	id := paramInt(r, "id")

	var exists int
	database.DB.Get(&exists, "SELECT COUNT(*) FROM devices WHERE id = ?", id)
	if exists == 0 {
		writeError(w, 404, "Device not found")
		return
	}

	result := service.TestDeviceConnection(id)
	writeJSON(w, 200, result)
}

type scheduleBody struct {
	Cron string `json:"cron"`
}

func SetSchedule(w http.ResponseWriter, r *http.Request) {
	id := paramInt(r, "id")
	var body scheduleBody
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}

	var exists int
	database.DB.Get(&exists, "SELECT COUNT(*) FROM devices WHERE id = ?", id)
	if exists == 0 {
		writeError(w, 404, "Device not found")
		return
	}

	database.DB.Exec("UPDATE devices SET schedule_cron = ?, updated_at = datetime('now') WHERE id = ?", body.Cron, id)
	service.ScheduleDevice(id, body.Cron)
	writeJSON(w, 200, map[string]string{"message": "Schedule updated", "cron": body.Cron})
}

func RemoveSchedule(w http.ResponseWriter, r *http.Request) {
	id := paramInt(r, "id")
	var exists int
	database.DB.Get(&exists, "SELECT COUNT(*) FROM devices WHERE id = ?", id)
	if exists == 0 {
		writeError(w, 404, "Device not found")
		return
	}

	database.DB.Exec("UPDATE devices SET schedule_cron = NULL, updated_at = datetime('now') WHERE id = ?", id)
	service.RemoveDeviceSchedule(id)
	writeJSON(w, 200, map[string]string{"message": "Schedule removed"})
}

func BulkTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=confbox_devices_template.csv")
	wr := csv.NewWriter(w)
	wr.Write([]string{"name", "vendor", "ip_address", "port", "platform", "auth_token", "ssh_username", "ssh_password", "enable_password", "location", "vdom"})
	wr.Write([]string{"FW-Istanbul", "fortigate", "10.0.1.1", "443", "", "api-key-here", "", "", "", "Istanbul DC", ""})
	wr.Write([]string{"SW-Core", "cisco", "10.0.1.2", "22", "ios", "", "admin", "pass123", "enable123", "Ankara DC", ""})
	wr.Write([]string{"JUN-Edge", "juniper", "10.0.1.3", "22", "", "", "admin", "pass456", "", "", ""})
	wr.Flush()
}

func BulkPreview(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "No file uploaded")
		return
	}
	defer file.Close()

	if !strings.HasSuffix(header.Filename, ".csv") {
		writeError(w, 400, "Only CSV files accepted")
		return
	}

	content, _ := io.ReadAll(io.LimitReader(file, 5<<20))
	reader := csv.NewReader(strings.NewReader(string(content)))
	headers, _ := reader.Read()
	colMap := make(map[string]int)
	for i, h := range headers {
		colMap[strings.TrimSpace(strings.ToLower(h))] = i
	}

	var existingNames []string
	database.DB.Select(&existingNames, "SELECT name FROM devices")
	nameSet := make(map[string]bool)
	for _, n := range existingNames {
		nameSet[n] = true
	}

	type locRow struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}
	var locs []locRow
	database.DB.Select(&locs, "SELECT id, name FROM locations")
	locMap := make(map[string]int)
	for _, l := range locs {
		locMap[strings.ToLower(l.Name)] = l.ID
	}

	validVendors := map[string]bool{"fortigate": true, "juniper": true, "cisco": true, "paloalto": true}
	validPlatforms := map[string]bool{"ios": true, "nxos": true, "asa": true}

	var rows []map[string]any
	rowNum := 0
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		rowNum++
		getCol := func(name string) string {
			if idx, ok := colMap[name]; ok && idx < len(record) {
				return strings.TrimSpace(record[idx])
			}
			return ""
		}

		name := getCol("name")
		vendor := strings.ToLower(getCol("vendor"))
		ip := getCol("ip_address")
		portStr := getCol("port")
		platform := strings.ToLower(getCol("platform"))
		if platform == "" {
			platform = "ios"
		}

		var errors []string
		if name == "" {
			errors = append(errors, "Name empty")
		} else if nameSet[name] {
			errors = append(errors, "Device name exists")
		}
		if !validVendors[vendor] {
			errors = append(errors, "Invalid vendor: "+vendor)
		}
		if ip == "" {
			errors = append(errors, "IP empty")
		}
		port := 0
		if portStr != "" {
			p, err := strconv.Atoi(portStr)
			if err != nil {
				errors = append(errors, "Invalid port")
			} else {
				port = p
			}
		} else if vendor == "fortigate" {
			port = 443
		} else {
			port = 22
		}
		if vendor == "cisco" && !validPlatforms[platform] {
			errors = append(errors, "Invalid platform: "+platform)
		}

		locName := getCol("location")
		var locID *int
		if locName != "" {
			if lid, ok := locMap[strings.ToLower(locName)]; ok {
				locID = &lid
			}
		}

		row := map[string]any{
			"row": rowNum, "name": name, "vendor": vendor, "ip_address": ip, "port": port,
			"platform": platform, "auth_token": getCol("auth_token") != "",
			"ssh_username": getCol("ssh_username"), "ssh_password": getCol("ssh_password") != "",
			"enable_password": getCol("enable_password") != "",
			"location": locName, "location_id": locID, "vdom": getCol("vdom"),
			"errors": errors, "valid": len(errors) == 0,
		}
		rows = append(rows, row)
		nameSet[name] = true
	}

	valid := 0
	for _, r := range rows {
		if r["valid"].(bool) {
			valid++
		}
	}
	writeJSON(w, 200, map[string]any{"rows": rows, "total": len(rows), "valid": valid, "invalid": len(rows) - valid})
}

func BulkImport(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20) // 5MB limit
	user := auth.GetUser(r)
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "No file uploaded")
		return
	}
	defer file.Close()

	if !strings.HasSuffix(header.Filename, ".csv") {
		writeError(w, 400, "Only CSV files accepted")
		return
	}

	content, _ := io.ReadAll(io.LimitReader(file, 5<<20))
	reader := csv.NewReader(strings.NewReader(string(content)))
	headers, _ := reader.Read()
	colMap := make(map[string]int)
	for i, h := range headers {
		colMap[strings.TrimSpace(strings.ToLower(h))] = i
	}

	var existingNames []string
	database.DB.Select(&existingNames, "SELECT name FROM devices")
	nameSet := make(map[string]bool)
	for _, n := range existingNames {
		nameSet[n] = true
	}

	type locRow struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}
	var locs []locRow
	database.DB.Select(&locs, "SELECT id, name FROM locations")
	locMap := make(map[string]int)
	for _, l := range locs {
		locMap[strings.ToLower(l.Name)] = l.ID
	}

	validVendors := map[string]bool{"fortigate": true, "juniper": true, "cisco": true, "paloalto": true}
	created, skipped := 0, 0

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		getCol := func(name string) string {
			if idx, ok := colMap[name]; ok && idx < len(record) {
				return strings.TrimSpace(record[idx])
			}
			return ""
		}

		name := getCol("name")
		vendor := strings.ToLower(getCol("vendor"))
		ip := getCol("ip_address")
		portStr := getCol("port")
		platform := strings.ToLower(getCol("platform"))
		if platform == "" {
			platform = "ios"
		}

		if name == "" || ip == "" || !validVendors[vendor] || nameSet[name] {
			skipped++
			continue
		}

		port := 22
		if portStr != "" {
			p, err := strconv.Atoi(portStr)
			if err != nil {
				skipped++
				continue
			}
			port = p
		} else if vendor == "fortigate" {
			port = 443
		}

		locName := getCol("location")
		var locID *int
		if locName != "" {
			if lid, ok := locMap[strings.ToLower(locName)]; ok {
				locID = &lid
			}
		}

		authToken := encryptIfNotEmpty(getCol("auth_token"))
		sshPass := encryptIfNotEmpty(getCol("ssh_password"))
		enablePass := encryptIfNotEmpty(getCol("enable_password"))
		sshUser := getCol("ssh_username")
		vdom := getCol("vdom")

		database.DB.Exec(`INSERT INTO devices (name, vendor, ip_address, port, location_id, vdom, auth_token, ssh_username, ssh_password, enable_password, platform, is_active, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, datetime('now'), datetime('now'))`,
			name, vendor, ip, port, locID, nilIfEmpty(vdom), authToken, nilIfEmpty(sshUser), sshPass, enablePass, platform)

		nameSet[name] = true
		created++
	}

	uid := user.ID
	service.LogAction(&uid, user.Username, "bulk_import", "device",
		fmt.Sprintf("%d devices", created), fmt.Sprintf("Created: %d, Skipped: %d", created, skipped), clientIP(r))
	writeJSON(w, 200, map[string]any{"created": created, "skipped": skipped})
}

func encryptOptional(s *string) *string {
	if s == nil || *s == "" {
		return s
	}
	enc := crypto.Encrypt(*s)
	return &enc
}

func encryptIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	enc := crypto.Encrypt(s)
	return &enc
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
