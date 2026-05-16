package service

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/yunuskargi/confbox/internal/database"
)

func getSmtpSettings() map[string]string {
	keys := []string{"smtp_host", "smtp_port", "smtp_username", "smtp_password", "smtp_use_tls", "smtp_from_email", "smtp_from_name"}
	defaults := map[string]string{
		"smtp_host": "", "smtp_port": "587", "smtp_username": "", "smtp_password": "",
		"smtp_use_tls": "true", "smtp_from_email": "", "smtp_from_name": "ConfBox",
	}
	result := make(map[string]string)
	for _, k := range keys {
		var val *string
		err := database.DB.Get(&val, "SELECT value FROM settings WHERE key = ?", k)
		if err != nil || val == nil {
			result[k] = defaults[k]
		} else {
			result[k] = *val
		}
	}
	return result
}

func getNotifySettings() map[string]string {
	keys := []string{"notify_on_success", "notify_on_failure", "notify_on_change", "notify_daily_summary", "notify_recipients"}
	defaults := map[string]string{
		"notify_on_success": "false", "notify_on_failure": "true",
		"notify_on_change": "false", "notify_daily_summary": "false", "notify_recipients": "",
	}
	result := make(map[string]string)
	for _, k := range keys {
		var val *string
		err := database.DB.Get(&val, "SELECT value FROM settings WHERE key = ?", k)
		if err != nil || val == nil {
			result[k] = defaults[k]
		} else {
			result[k] = *val
		}
	}
	return result
}

func sendEmail(to, subject, bodyHTML string) error {
	s := getSmtpSettings()
	if s["smtp_host"] == "" || to == "" {
		return fmt.Errorf("SMTP not configured")
	}

	from := s["smtp_from_email"]
	msg := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s["smtp_from_name"], from, to, subject, bodyHTML)

	addr := s["smtp_host"] + ":" + s["smtp_port"]
	var auth smtp.Auth
	if s["smtp_username"] != "" {
		auth = smtp.PlainAuth("", s["smtp_username"], s["smtp_password"], s["smtp_host"])
	}

	recipients := strings.Split(to, ",")
	for i := range recipients {
		recipients[i] = strings.TrimSpace(recipients[i])
	}

	return smtp.SendMail(addr, auth, from, recipients, []byte(msg))
}

func SendTestEmail(to string) error {
	html := `<div style="text-align:center;padding:20px"><p style="font-size:18px;color:#16a34a">SMTP Connection Successful</p><p>Email notifications are working.</p></div>`
	return sendEmail(to, "ConfBox - Test Email", html)
}

func NotifyBackup(deviceName, vendor, status, errMsg, filePath string, fileSize int, location, vdom, triggeredBy string) {
	notify := getNotifySettings()
	recipients := notify["notify_recipients"]
	if recipients == "" {
		return
	}
	if status == "success" && notify["notify_on_success"] != "true" {
		return
	}
	if status == "failed" && notify["notify_on_failure"] != "true" {
		return
	}

	statusText := "Success"
	emoji := "✅"
	if status == "failed" {
		statusText = "Failed"
		emoji = "❌"
	}

	sizeStr := ""
	if fileSize > 0 {
		if fileSize > 1024*1024 {
			sizeStr = fmt.Sprintf("%.1f MB", float64(fileSize)/(1024*1024))
		} else if fileSize > 1024 {
			sizeStr = fmt.Sprintf("%.1f KB", float64(fileSize)/1024)
		} else {
			sizeStr = fmt.Sprintf("%d B", fileSize)
		}
	}

	body := fmt.Sprintf(`<h2>Backup %s</h2><p><b>Device:</b> %s</p><p><b>Vendor:</b> %s</p>`, statusText, deviceName, vendor)
	if location != "" {
		body += fmt.Sprintf(`<p><b>Location:</b> %s</p>`, location)
	}
	if vdom != "" {
		body += fmt.Sprintf(`<p><b>VDOM:</b> %s</p>`, vdom)
	}
	body += fmt.Sprintf(`<p><b>Triggered by:</b> %s</p>`, triggeredBy)
	if sizeStr != "" {
		body += fmt.Sprintf(`<p><b>Size:</b> %s</p>`, sizeStr)
	}
	if errMsg != "" {
		body += fmt.Sprintf(`<p style="color:red"><b>Error:</b> %s</p>`, errMsg)
	}

	subject := fmt.Sprintf("ConfBox %s %s - Backup %s", emoji, deviceName, statusText)
	if err := sendEmail(recipients, subject, body); err != nil {
		slog.Error("failed to send backup notification", "error", err)
	}
}

func NotifyConfigChange(deviceName, vendor, location, vdom string) {
	notify := getNotifySettings()
	if notify["notify_on_change"] != "true" {
		return
	}
	recipients := notify["notify_recipients"]
	if recipients == "" {
		return
	}

	body := fmt.Sprintf(`<h2>Config Change Detected</h2><p><b>Device:</b> %s</p><p><b>Vendor:</b> %s</p>`, deviceName, vendor)
	if location != "" {
		body += fmt.Sprintf(`<p><b>Location:</b> %s</p>`, location)
	}

	subject := fmt.Sprintf("ConfBox ⚠️ %s - Config Change", deviceName)
	if err := sendEmail(recipients, subject, body); err != nil {
		slog.Error("failed to send config change notification", "error", err)
	}
}

func SendDailySummary() {
	notify := getNotifySettings()
	if notify["notify_daily_summary"] != "true" {
		return
	}
	recipients := notify["notify_recipients"]
	if recipients == "" {
		return
	}

	var total, success, failed, activeDevices int
	database.DB.Get(&total, "SELECT COUNT(*) FROM backups WHERE created_at >= datetime('now', '-1 day')")
	database.DB.Get(&success, "SELECT COUNT(*) FROM backups WHERE created_at >= datetime('now', '-1 day') AND status = 'success'")
	database.DB.Get(&failed, "SELECT COUNT(*) FROM backups WHERE created_at >= datetime('now', '-1 day') AND status = 'failed'")
	database.DB.Get(&activeDevices, "SELECT COUNT(*) FROM devices WHERE is_active = 1")

	rate := 0.0
	if total > 0 {
		rate = float64(success) / float64(total) * 100
	}

	body := fmt.Sprintf(`<h2>Daily Backup Summary</h2>
		<p><b>Active Devices:</b> %d</p>
		<p><b>Total Backups:</b> %d</p>
		<p><b>Successful:</b> %d</p>
		<p><b>Failed:</b> %d</p>
		<p><b>Success Rate:</b> %s%%</p>`,
		activeDevices, total, success, failed, strconv.FormatFloat(rate, 'f', 1, 64))

	if err := sendEmail(recipients, "ConfBox 📊 Daily Summary", body); err != nil {
		slog.Error("failed to send daily summary", "error", err)
	}
}
