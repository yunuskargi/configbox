package handler

import (
	"net/http"
	"strconv"

	"fmt"

	"github.com/yunuskargi/confbox/internal/auth"
	"github.com/yunuskargi/confbox/internal/crypto"
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
	archiveDays, _ := strconv.Atoi(getSetting("archive_after_days", "0"))
	writeJSON(w, 200, map[string]any{
		"backup_dir":         getSetting("backup_dir", "/data/backups"),
		"retention_days":     retDays,
		"app_title":          getSetting("app_title", ""),
		"archive_enabled":    getSetting("archive_enabled", "false") == "true",
		"archive_after_days": archiveDays,
	})
}

type settingsUpdateBody struct {
	BackupDir       *string `json:"backup_dir"`
	RetentionDays   *int    `json:"retention_days"`
	AppTitle        *string `json:"app_title"`
	ArchiveEnabled  *bool   `json:"archive_enabled"`
	ArchiveAfterDays *int   `json:"archive_after_days"`
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
	if body.ArchiveEnabled != nil {
		setSetting("archive_enabled", strconv.FormatBool(*body.ArchiveEnabled))
	}
	if body.ArchiveAfterDays != nil {
		setSetting("archive_after_days", strconv.Itoa(*body.ArchiveAfterDays))
	}
	uid := user.ID
	service.LogAction(&uid, user.Username, "update", "settings", "general", "", clientIP(r))

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
	uid := user.ID
	service.LogAction(&uid, user.Username, "update", "settings", "smtp", fmt.Sprintf("Host: %s", body.Host), clientIP(r))

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
		uid := user.ID
		service.LogAction(&uid, user.Username, "test_smtp", "settings", "smtp", fmt.Sprintf("Failed: %s", err.Error()), clientIP(r))
		writeError(w, 400, err.Error())
		return
	}
	uid := user.ID
	service.LogAction(&uid, user.Username, "test_smtp", "settings", "smtp", "Success", clientIP(r))
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
	uid := user.ID
	service.LogAction(&uid, user.Username, "update", "settings", "notifications", "", clientIP(r))

	writeJSON(w, 200, map[string]string{"message": "Notification settings updated"})
}

// S3 Settings

func GetS3(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}

	accessKey := getSetting("s3_access_key", "")
	secretKey := getSetting("s3_secret_key", "")
	maskedAccess, maskedSecret := "", ""
	if accessKey != "" {
		maskedAccess = "••••••••"
	}
	if secretKey != "" {
		maskedSecret = "••••••••"
	}

	writeJSON(w, 200, map[string]any{
		"s3_enabled":    getSetting("s3_enabled", "false") == "true",
		"s3_endpoint":   getSetting("s3_endpoint", ""),
		"s3_region":     getSetting("s3_region", "us-east-1"),
		"s3_bucket":     getSetting("s3_bucket", ""),
		"s3_access_key": maskedAccess,
		"s3_secret_key": maskedSecret,
		"s3_use_ssl":    getSetting("s3_use_ssl", "true") == "true",
		"s3_prefix":     getSetting("s3_prefix", ""),
	})
}

type s3Body struct {
	Enabled   bool   `json:"s3_enabled"`
	Endpoint  string `json:"s3_endpoint"`
	Region    string `json:"s3_region"`
	Bucket    string `json:"s3_bucket"`
	AccessKey string `json:"s3_access_key"`
	SecretKey string `json:"s3_secret_key"`
	UseSSL    bool   `json:"s3_use_ssl"`
	Prefix    string `json:"s3_prefix"`
}

func UpdateS3(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}

	var body s3Body
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}

	setSetting("s3_enabled", strconv.FormatBool(body.Enabled))
	setSetting("s3_endpoint", body.Endpoint)
	setSetting("s3_region", body.Region)
	setSetting("s3_bucket", body.Bucket)
	if body.AccessKey != "" && body.AccessKey != "••••••••" {
		setSetting("s3_access_key", crypto.Encrypt(body.AccessKey))
	}
	if body.SecretKey != "" && body.SecretKey != "••••••••" {
		setSetting("s3_secret_key", crypto.Encrypt(body.SecretKey))
	}
	setSetting("s3_use_ssl", strconv.FormatBool(body.UseSSL))
	setSetting("s3_prefix", body.Prefix)
	uid := user.ID
	service.LogAction(&uid, user.Username, "update", "settings", "s3",
		fmt.Sprintf("Enabled: %v, Bucket: %s", body.Enabled, body.Bucket), clientIP(r))

	writeJSON(w, 200, map[string]string{"message": "S3 settings updated"})
}

func TestS3(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}

	if err := service.TestS3Connection(); err != nil {
		uid := user.ID
		service.LogAction(&uid, user.Username, "test_s3", "settings", "s3", fmt.Sprintf("Failed: %s", err.Error()), clientIP(r))
		writeError(w, 400, err.Error())
		return
	}
	uid := user.ID
	service.LogAction(&uid, user.Username, "test_s3", "settings", "s3", "Success", clientIP(r))
	writeJSON(w, 200, map[string]string{"message": "S3 connection successful"})
}

// Google Drive Settings

func GetGDrive(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}

	clientSecret := getSetting("gdrive_client_secret", "")
	maskedSecret := ""
	if clientSecret != "" {
		maskedSecret = "••••••••"
	}
	refreshToken := getSetting("gdrive_refresh_token", "")

	writeJSON(w, 200, map[string]any{
		"gdrive_enabled":       getSetting("gdrive_enabled", "false") == "true",
		"gdrive_client_id":     getSetting("gdrive_client_id", ""),
		"gdrive_client_secret": maskedSecret,
		"gdrive_folder_id":     getSetting("gdrive_folder_id", ""),
		"gdrive_authorized":    refreshToken != "",
	})
}

type gdriveBody struct {
	Enabled      bool   `json:"gdrive_enabled"`
	ClientID     string `json:"gdrive_client_id"`
	ClientSecret string `json:"gdrive_client_secret"`
	FolderID     string `json:"gdrive_folder_id"`
}

func UpdateGDrive(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}

	var body gdriveBody
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}

	setSetting("gdrive_enabled", strconv.FormatBool(body.Enabled))
	setSetting("gdrive_client_id", body.ClientID)
	if body.ClientSecret != "" && body.ClientSecret != "••••••••" {
		setSetting("gdrive_client_secret", crypto.Encrypt(body.ClientSecret))
	}
	setSetting("gdrive_folder_id", body.FolderID)
	uid := user.ID
	service.LogAction(&uid, user.Username, "update", "settings", "gdrive",
		fmt.Sprintf("Enabled: %v", body.Enabled), clientIP(r))

	writeJSON(w, 200, map[string]string{"message": "Google Drive settings updated"})
}

func GDriveAuthURL(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}

	url, err := service.GetGDriveAuthURL()
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"url": url})
}

func GDriveCallback(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}

	var body struct {
		Code string `json:"code"`
	}
	if err := decodeBody(r, &body); err != nil || body.Code == "" {
		writeError(w, 400, "Authorization code is required")
		return
	}

	if err := service.ExchangeGDriveCode(body.Code); err != nil {
		uid := user.ID
		service.LogAction(&uid, user.Username, "gdrive_auth", "settings", "gdrive", fmt.Sprintf("Failed: %s", err.Error()), clientIP(r))
		writeError(w, 400, err.Error())
		return
	}
	uid := user.ID
	service.LogAction(&uid, user.Username, "gdrive_auth", "settings", "gdrive", "Authorized successfully", clientIP(r))
	writeJSON(w, 200, map[string]string{"message": "Google Drive authorized successfully"})
}

func TestGDrive(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}

	if err := service.TestGDriveConnection(); err != nil {
		uid := user.ID
		service.LogAction(&uid, user.Username, "test_gdrive", "settings", "gdrive", fmt.Sprintf("Failed: %s", err.Error()), clientIP(r))
		writeError(w, 400, err.Error())
		return
	}
	uid := user.ID
	service.LogAction(&uid, user.Username, "test_gdrive", "settings", "gdrive", "Success", clientIP(r))
	writeJSON(w, 200, map[string]string{"message": "Google Drive connection successful"})
}

func RunArchive(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}
	service.ArchiveOldBackups()

	uid := user.ID
	service.LogAction(&uid, user.Username, "archive", "settings", "archive", "Manual archive triggered", clientIP(r))

	writeJSON(w, 200, map[string]string{"message": "Archive completed"})
}
