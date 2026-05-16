package handler

import (
	"net/http"
	"strconv"

	"github.com/yunuskargi/confbox/internal/auth"
	"github.com/yunuskargi/confbox/internal/database"
	"github.com/yunuskargi/confbox/internal/service"
)

func getSetting(key, def string) string {
	var val *string
	err := database.DB.Get(&val, "SELECT value FROM settings WHERE key = ?", key)
	if err != nil || val == nil {
		return def
	}
	return *val
}

func setSetting(key, value string) {
	database.DB.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", key, value)
}

func GetBranding(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{
		"app_title": getSetting("app_title", ""),
	})
}

func GetSettings(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}
	retDays, _ := strconv.Atoi(getSetting("retention_days", "90"))
	writeJSON(w, 200, map[string]any{
		"backup_dir":     getSetting("backup_dir", "/data/backups"),
		"retention_days": retDays,
		"app_title":      getSetting("app_title", ""),
	})
}

type settingsUpdateBody struct {
	BackupDir     *string `json:"backup_dir"`
	RetentionDays *int    `json:"retention_days"`
	AppTitle      *string `json:"app_title"`
}

func UpdateSettings(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}

	var body settingsUpdateBody
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}
	if body.BackupDir != nil {
		setSetting("backup_dir", *body.BackupDir)
	}
	if body.RetentionDays != nil {
		setSetting("retention_days", strconv.Itoa(*body.RetentionDays))
	}
	if body.AppTitle != nil {
		setSetting("app_title", *body.AppTitle)
	}
	writeJSON(w, 200, map[string]string{"message": "Settings updated"})
}

func GetSMTP(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}

	port, _ := strconv.Atoi(getSetting("smtp_port", "587"))
	password := getSetting("smtp_password", "")
	maskedPassword := ""
	if password != "" {
		maskedPassword = "••••••••"
	}
	writeJSON(w, 200, map[string]any{
		"smtp_host":       getSetting("smtp_host", ""),
		"smtp_port":       port,
		"smtp_username":   getSetting("smtp_username", ""),
		"smtp_password":   maskedPassword,
		"smtp_use_tls":    getSetting("smtp_use_tls", "true") == "true",
		"smtp_from_email": getSetting("smtp_from_email", ""),
		"smtp_from_name":  getSetting("smtp_from_name", "ConfBox"),
	})
}

type smtpBody struct {
	Host      string `json:"smtp_host"`
	Port      int    `json:"smtp_port"`
	Username  string `json:"smtp_username"`
	Password  string `json:"smtp_password"`
	UseTLS    bool   `json:"smtp_use_tls"`
	FromEmail string `json:"smtp_from_email"`
	FromName  string `json:"smtp_from_name"`
}

func UpdateSMTP(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}

	var body smtpBody
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}

	setSetting("smtp_host", body.Host)
	setSetting("smtp_port", strconv.Itoa(body.Port))
	setSetting("smtp_username", body.Username)
	if body.Password != "" && body.Password != "••••••••" {
		setSetting("smtp_password", body.Password)
	}
	setSetting("smtp_use_tls", strconv.FormatBool(body.UseTLS))
	setSetting("smtp_from_email", body.FromEmail)
	setSetting("smtp_from_name", body.FromName)
	writeJSON(w, 200, map[string]string{"message": "SMTP settings updated"})
}

func TestSMTP(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}

	recipients := getSetting("notify_recipients", "")
	if recipients == "" {
		writeError(w, 400, "No recipients configured")
		return
	}

	if err := service.SendTestEmail(recipients); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"message": "Test email sent"})
}

func GetNotify(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}

	writeJSON(w, 200, map[string]any{
		"notify_on_success":   getSetting("notify_on_success", "false") == "true",
		"notify_on_failure":   getSetting("notify_on_failure", "true") == "true",
		"notify_on_change":    getSetting("notify_on_change", "false") == "true",
		"notify_daily_summary": getSetting("notify_daily_summary", "false") == "true",
		"notify_recipients":   getSetting("notify_recipients", ""),
	})
}

type notifyBody struct {
	OnSuccess    bool   `json:"notify_on_success"`
	OnFailure    bool   `json:"notify_on_failure"`
	OnChange     bool   `json:"notify_on_change"`
	DailySummary bool   `json:"notify_daily_summary"`
	Recipients   string `json:"notify_recipients"`
}

func UpdateNotify(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}

	var body notifyBody
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}

	setSetting("notify_on_success", strconv.FormatBool(body.OnSuccess))
	setSetting("notify_on_failure", strconv.FormatBool(body.OnFailure))
	setSetting("notify_on_change", strconv.FormatBool(body.OnChange))
	setSetting("notify_daily_summary", strconv.FormatBool(body.DailySummary))
	setSetting("notify_recipients", body.Recipients)
	writeJSON(w, 200, map[string]string{"message": "Notification settings updated"})
}
