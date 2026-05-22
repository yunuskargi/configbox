package service

import (
	"encoding/base64"
	"fmt"
	"html"
	"log/slog"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/yunuskargi/configbox/internal/config"
	"github.com/yunuskargi/configbox/internal/crypto"
	"github.com/yunuskargi/configbox/internal/database"
)

func encodeSubject(s string) string {
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(s)) + "?="
}

func getSmtpSettings() map[string]string {
	keys := []string{"smtp_host", "smtp_port", "smtp_username", "smtp_password", "smtp_use_tls", "smtp_from_email", "smtp_from_name"}
	defaults := map[string]string{
		"smtp_host": "", "smtp_port": "587", "smtp_username": "", "smtp_password": "",
		"smtp_use_tls": "true", "smtp_from_email": "", "smtp_from_name": "ConfigBox",
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
		s["smtp_from_name"], from, to, encodeSubject(subject), bodyHTML)

	addr := s["smtp_host"] + ":" + s["smtp_port"]
	var auth smtp.Auth
	if s["smtp_username"] != "" {
		auth = smtp.PlainAuth("", s["smtp_username"], crypto.Decrypt(s["smtp_password"]), s["smtp_host"])
	}

	recipients := strings.Split(to, ",")
	for i := range recipients {
		recipients[i] = strings.TrimSpace(recipients[i])
	}

	return smtp.SendMail(addr, auth, from, recipients, []byte(msg))
}

func mailWrapper(title, content string) string {
	now := time.Now().In(config.AppTimezone).Format("02 Jan 2006, 15:04")
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"></head>
<body style="margin:0;padding:0;background-color:#f1f5f9;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif">
<table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f1f5f9;padding:32px 16px">
<tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%">
<!-- Header -->
<tr><td style="background:linear-gradient(135deg,#0e7490 0%%,#0891b2 100%%);padding:32px 40px;border-radius:12px 12px 0 0">
<table width="100%%"><tr>
<td><span style="font-size:24px;font-weight:700;color:#ffffff;letter-spacing:-0.5px">ConfigBox</span></td>
<td align="right"><span style="font-size:12px;color:#cffafe">%s</span></td>
</tr></table>
<p style="margin:8px 0 0;font-size:14px;color:#e0f7fa">Network Configuration Backup Manager</p>
</td></tr>
<!-- Title Bar -->
<tr><td style="background-color:#ffffff;padding:24px 40px 0;border-left:1px solid #e2e8f0;border-right:1px solid #e2e8f0">
<h1 style="margin:0;font-size:20px;font-weight:600;color:#0e7490">%s</h1>
<hr style="border:none;border-top:2px solid #cffafe;margin:16px 0 0">
</td></tr>
<!-- Content -->
<tr><td style="background-color:#ffffff;padding:24px 40px 32px;border-left:1px solid #e2e8f0;border-right:1px solid #e2e8f0">
%s
</td></tr>
<!-- Footer -->
<tr><td style="background-color:#f0fdfa;padding:20px 40px;border-radius:0 0 12px 12px;border:1px solid #e2e8f0;border-top:none">
<p style="margin:0;font-size:12px;color:#0e7490;text-align:center">
This is an automated notification from ConfigBox. Do not reply to this email.
</p>
</td></tr>
</table>
</td></tr></table>
</body></html>`, now, html.EscapeString(title), content)
}

func infoRow(label, value, bgColor string) string {
	if value == "" {
		return ""
	}
	return fmt.Sprintf(`<tr><td style="padding:10px 16px;font-size:13px;font-weight:600;color:#475569;background-color:%s;width:140px;border-bottom:1px solid #f1f5f9">%s</td><td style="padding:10px 16px;font-size:13px;color:#1e293b;background-color:%s;border-bottom:1px solid #f1f5f9">%s</td></tr>`,
		bgColor, html.EscapeString(label), bgColor, html.EscapeString(value))
}

func badge(text, bgColor, textColor string) string {
	return fmt.Sprintf(`<span style="display:inline-block;padding:4px 12px;border-radius:20px;font-size:12px;font-weight:600;background-color:%s;color:%s">%s</span>`,
		bgColor, textColor, html.EscapeString(text))
}

func statusBanner(status string) string {
	if status == "success" {
		return `<div style="background-color:#f0fdf4;border:1px solid #bbf7d0;border-radius:8px;padding:16px 20px;margin-bottom:20px">
<span style="font-size:20px;vertical-align:middle">✅</span>
<span style="font-size:15px;font-weight:600;color:#166534;margin-left:8px;vertical-align:middle">Backup Completed Successfully</span>
</div>`
	}
	return `<div style="background-color:#fef2f2;border:1px solid #fecaca;border-radius:8px;padding:16px 20px;margin-bottom:20px">
<span style="font-size:20px;vertical-align:middle">❌</span>
<span style="font-size:15px;font-weight:600;color:#991b1b;margin-left:8px;vertical-align:middle">Backup Failed</span>
</div>`
}

func formatSize(fileSize int) string {
	if fileSize <= 0 {
		return ""
	}
	if fileSize > 1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(fileSize)/(1024*1024))
	}
	if fileSize > 1024 {
		return fmt.Sprintf("%.1f KB", float64(fileSize)/1024)
	}
	return fmt.Sprintf("%d B", fileSize)
}

func SendTestEmail(to string) error {
	content := `<div style="text-align:center;padding:20px 0">
<div style="background-color:#ecfeff;border:1px solid #a5f3fc;border-radius:8px;padding:24px;margin-bottom:20px">
<span style="font-size:48px;display:block;margin-bottom:12px">✅</span>
<p style="margin:0;font-size:18px;font-weight:600;color:#0e7490">SMTP Connection Successful</p>
<p style="margin:8px 0 0;font-size:14px;color:#0891b2">Email notifications are configured and working correctly.</p>
</div>
<table width="100%" style="margin-top:20px;border-collapse:collapse;border-radius:8px;overflow:hidden;border:1px solid #e2e8f0">
<tr><td style="padding:12px 16px;font-size:13px;font-weight:600;color:#475569;background-color:#f0fdfa;border-bottom:1px solid #e2e8f0">Status</td><td style="padding:12px 16px;font-size:13px;color:#1e293b;background-color:#f0fdfa;border-bottom:1px solid #e2e8f0">` + badge("Connected", "#cffafe", "#0e7490") + `</td></tr>
<tr><td style="padding:12px 16px;font-size:13px;font-weight:600;color:#475569;background-color:#ffffff;border-bottom:1px solid #e2e8f0">Recipient</td><td style="padding:12px 16px;font-size:13px;color:#1e293b;background-color:#ffffff;border-bottom:1px solid #e2e8f0">` + html.EscapeString(to) + `</td></tr>
<tr><td style="padding:12px 16px;font-size:13px;font-weight:600;color:#475569;background-color:#f8fafc">Timestamp</td><td style="padding:12px 16px;font-size:13px;color:#1e293b;background-color:#f8fafc">` + time.Now().In(config.AppTimezone).Format("02 Jan 2006, 15:04:05") + `</td></tr>
</table>
<p style="margin:20px 0 0;font-size:12px;color:#64748b">If you received this email, your SMTP settings are configured correctly.</p>
</div>`

	body := mailWrapper("SMTP Test Notification", content)
	return sendEmail(to, "ConfigBox - Test Email - Connection Successful", body)
}

func NotifyBackup(deviceName, vendor, status, errMsg, filePath string, fileSize int, location, vdom, triggeredBy string, remote RemoteResult) {
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

	now := time.Now().In(config.AppTimezone).Format("02 Jan 2006, 15:04:05")

	var content strings.Builder
	content.WriteString(statusBanner(status))

	content.WriteString(`<table width="100%" style="border-collapse:collapse;border-radius:8px;overflow:hidden;border:1px solid #e2e8f0;margin-bottom:20px">`)
	content.WriteString(infoRow("Device", deviceName, "#f8fafc"))
	content.WriteString(infoRow("Vendor", strings.ToUpper(vendor), "#ffffff"))
	if location != "" {
		content.WriteString(infoRow("Location", location, "#f8fafc"))
	}
	if vdom != "" {
		content.WriteString(infoRow("VDOM", vdom, "#ffffff"))
	}
	content.WriteString(infoRow("Triggered By", triggeredBy, "#f8fafc"))
	content.WriteString(infoRow("Timestamp", now, "#ffffff"))

	if status == "success" {
		sizeStr := formatSize(fileSize)
		if sizeStr != "" {
			content.WriteString(infoRow("File Size", sizeStr, "#f8fafc"))
		}
		if filePath != "" {
			content.WriteString(infoRow("File Path", filePath, "#ffffff"))
		}
	}
	content.WriteString(`</table>`)

	if status == "success" && (remote.S3Enabled || remote.GDriveEnabled) {
		content.WriteString(`<div style="background-color:#f8fafc;border:1px solid #e2e8f0;border-radius:8px;padding:16px 20px;margin-top:16px">
<p style="margin:0 0 10px;font-size:12px;font-weight:600;color:#0e7490;text-transform:uppercase;letter-spacing:0.5px">Remote Backup</p>`)
		if remote.S3Enabled {
			if remote.S3OK {
				content.WriteString(`<p style="margin:0 0 4px;font-size:13px;color:#166534">✅ S3 — Uploaded successfully</p>`)
			} else {
				content.WriteString(fmt.Sprintf(`<p style="margin:0 0 4px;font-size:13px;color:#991b1b">❌ S3 — %s</p>`, html.EscapeString(remote.S3Error)))
			}
		}
		if remote.GDriveEnabled {
			if remote.GDriveOK {
				content.WriteString(`<p style="margin:0;font-size:13px;color:#166534">✅ Google Drive — Uploaded successfully</p>`)
			} else {
				content.WriteString(fmt.Sprintf(`<p style="margin:0;font-size:13px;color:#991b1b">❌ Google Drive — %s</p>`, html.EscapeString(remote.GDriveError)))
			}
		}
		content.WriteString(`</div>`)
	}

	if status == "failed" && errMsg != "" {
		content.WriteString(fmt.Sprintf(`<div style="background-color:#fef2f2;border:1px solid #fecaca;border-left:4px solid #dc2626;border-radius:8px;padding:16px 20px;margin-top:16px">
<p style="margin:0 0 8px;font-size:12px;font-weight:600;color:#991b1b;text-transform:uppercase;letter-spacing:0.5px">Error Details</p>
<p style="margin:0;font-size:13px;color:#7f1d1d;font-family:monospace;word-break:break-all">%s</p>
</div>`, html.EscapeString(errMsg)))
	}

	statusText := "Success"
	if status == "failed" {
		statusText = "Failed"
	}

	title := fmt.Sprintf("Backup %s - %s", statusText, deviceName)
	subject := fmt.Sprintf("ConfigBox - %s - Backup %s", deviceName, statusText)
	body := mailWrapper(title, content.String())

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

	now := time.Now().In(config.AppTimezone).Format("02 Jan 2006, 15:04:05")

	var content strings.Builder
	content.WriteString(`<div style="background-color:#fffbeb;border:1px solid #fde68a;border-radius:8px;padding:16px 20px;margin-bottom:20px">
<span style="font-size:20px;vertical-align:middle">⚠️</span>
<span style="font-size:15px;font-weight:600;color:#92400e;margin-left:8px;vertical-align:middle">Configuration Change Detected</span>
</div>`)

	content.WriteString(`<p style="margin:0 0 16px;font-size:14px;color:#475569">A configuration change has been detected on the following device. The new configuration differs from the previous backup.</p>`)

	content.WriteString(`<table width="100%" style="border-collapse:collapse;border-radius:8px;overflow:hidden;border:1px solid #e2e8f0;margin-bottom:20px">`)
	content.WriteString(infoRow("Device", deviceName, "#f8fafc"))
	content.WriteString(infoRow("Vendor", strings.ToUpper(vendor), "#ffffff"))
	if location != "" {
		content.WriteString(infoRow("Location", location, "#f8fafc"))
	}
	if vdom != "" {
		content.WriteString(infoRow("VDOM", vdom, "#ffffff"))
	}
	content.WriteString(infoRow("Detected At", now, "#f8fafc"))
	content.WriteString(`</table>`)

	content.WriteString(`<div style="background-color:#ecfeff;border:1px solid #a5f3fc;border-radius:8px;padding:16px 20px">
<p style="margin:0;font-size:13px;color:#0e7490">💡 <strong>Tip:</strong> Use the ConfigBox dashboard to compare configurations and view the exact changes using the diff viewer.</p>
</div>`)

	title := fmt.Sprintf("Config Change - %s", deviceName)
	subject := fmt.Sprintf("ConfigBox - %s - Configuration Changed", deviceName)
	body := mailWrapper(title, content.String())

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

	rateColor := "#166534"
	rateBg := "#dcfce7"
	if rate < 80 {
		rateColor = "#991b1b"
		rateBg = "#fee2e2"
	} else if rate < 95 {
		rateColor = "#92400e"
		rateBg = "#fef3c7"
	}

	var content strings.Builder

	content.WriteString(`<p style="margin:0 0 20px;font-size:14px;color:#475569">Here is your daily backup performance summary for the last 24 hours.</p>`)

	// Stat cards
	content.WriteString(`<table width="100%" cellpadding="0" cellspacing="0" style="margin-bottom:24px"><tr>`)
	content.WriteString(statCard("Active Devices", strconv.Itoa(activeDevices), "#0891b2", "#ecfeff"))
	content.WriteString(statCard("Total Backups", strconv.Itoa(total), "#0e7490", "#cffafe"))
	content.WriteString(statCard("Successful", strconv.Itoa(success), "#16a34a", "#f0fdf4"))
	content.WriteString(statCard("Failed", strconv.Itoa(failed), "#dc2626", "#fef2f2"))
	content.WriteString(`</tr></table>`)

	// Success rate bar
	content.WriteString(fmt.Sprintf(`<div style="background-color:#f8fafc;border:1px solid #e2e8f0;border-radius:8px;padding:20px;margin-bottom:24px">
<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
<span style="font-size:14px;font-weight:600;color:#0e7490">Success Rate</span>
<span style="font-size:20px;font-weight:700;color:%s">%s%%</span>
</div>
<div style="background-color:#e2e8f0;border-radius:100px;height:10px;overflow:hidden">
<div style="background-color:%s;height:10px;border-radius:100px;width:%s%%"></div>
</div>
</div>`, rateColor, strconv.FormatFloat(rate, 'f', 1, 64), rateColor, strconv.FormatFloat(rate, 'f', 0, 64)))

	// Success rate badge
	content.WriteString(fmt.Sprintf(`<div style="text-align:center;margin-bottom:24px">%s</div>`, badge(fmt.Sprintf("%.1f%% Success Rate", rate), rateBg, rateColor)))

	// Failed devices table
	if failed > 0 {
		type failedDevice struct {
			DeviceName string `db:"device_name"`
			Vendor     string `db:"vendor"`
			ErrMsg     string `db:"error_message"`
		}
		var failedList []failedDevice
		database.DB.Select(&failedList, `SELECT d.name as device_name, d.vendor, b.error_message
			FROM backups b JOIN devices d ON d.id = b.device_id
			WHERE b.created_at >= datetime('now', '-1 day') AND b.status = 'failed'
			ORDER BY b.created_at DESC LIMIT 10`)

		if len(failedList) > 0 {
			content.WriteString(`<div style="margin-bottom:20px">
<p style="margin:0 0 12px;font-size:14px;font-weight:600;color:#991b1b">❌ Failed Backups</p>
<table width="100%" style="border-collapse:collapse;border-radius:8px;overflow:hidden;border:1px solid #fecaca">
<tr style="background-color:#fef2f2">
<th style="padding:10px 12px;font-size:12px;font-weight:600;color:#991b1b;text-align:left;border-bottom:1px solid #fecaca">Device</th>
<th style="padding:10px 12px;font-size:12px;font-weight:600;color:#991b1b;text-align:left;border-bottom:1px solid #fecaca">Vendor</th>
<th style="padding:10px 12px;font-size:12px;font-weight:600;color:#991b1b;text-align:left;border-bottom:1px solid #fecaca">Error</th>
</tr>`)
			for i, fd := range failedList {
				bg := "#ffffff"
				if i%2 == 1 {
					bg = "#fff5f5"
				}
				errTrunc := fd.ErrMsg
				if len(errTrunc) > 60 {
					errTrunc = errTrunc[:60] + "..."
				}
				content.WriteString(fmt.Sprintf(`<tr style="background-color:%s">
<td style="padding:8px 12px;font-size:13px;color:#1e293b;border-bottom:1px solid #fee2e2">%s</td>
<td style="padding:8px 12px;font-size:13px;color:#475569;border-bottom:1px solid #fee2e2">%s</td>
<td style="padding:8px 12px;font-size:12px;color:#7f1d1d;font-family:monospace;border-bottom:1px solid #fee2e2">%s</td>
</tr>`, bg, html.EscapeString(fd.DeviceName), html.EscapeString(strings.ToUpper(fd.Vendor)), html.EscapeString(errTrunc)))
			}
			content.WriteString(`</table></div>`)
		}
	}

	// Footer note
	if total == 0 {
		content.WriteString(`<div style="background-color:#f8fafc;border:1px solid #e2e8f0;border-radius:8px;padding:16px 20px;text-align:center">
<p style="margin:0;font-size:13px;color:#64748b">No backup jobs were executed in the last 24 hours.</p>
</div>`)
	}

	title := "Daily Backup Summary"
	subject := fmt.Sprintf("ConfigBox - Daily Summary - %s", time.Now().In(config.AppTimezone).Format("02 Jan 2006"))
	body := mailWrapper(title, content.String())

	if err := sendEmail(recipients, subject, body); err != nil {
		slog.Error("failed to send daily summary", "error", err)
	}
}

func statCard(label, value, color, bgColor string) string {
	return fmt.Sprintf(`<td width="25%%" style="padding:0 4px"><div style="background-color:%s;border-radius:8px;padding:16px;text-align:center">
<p style="margin:0;font-size:22px;font-weight:700;color:%s">%s</p>
<p style="margin:4px 0 0;font-size:11px;font-weight:600;color:%s;text-transform:uppercase;letter-spacing:0.5px">%s</p>
</div></td>`, bgColor, color, html.EscapeString(value), color, html.EscapeString(label))
}
