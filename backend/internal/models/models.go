package models

import (
	"database/sql"
)

type User struct {
	ID                 int            `db:"id" json:"id"`
	Username           string         `db:"username" json:"username"`
	PasswordHash       string         `db:"password_hash" json:"-"`
	Role               string         `db:"role" json:"role"`
	TOTPSecret         sql.NullString `db:"totp_secret" json:"-"`
	TOTPEnabled        bool           `db:"totp_enabled" json:"totp_enabled"`
	MustChangePassword bool           `db:"must_change_password" json:"must_change_password"`
	CreatedAt          string         `db:"created_at" json:"created_at"`
}

type Location struct {
	ID          int            `db:"id" json:"id"`
	Name        string         `db:"name" json:"name"`
	Description sql.NullString `db:"description" json:"description"`
	CreatedAt   string         `db:"created_at" json:"created_at"`
}

type Device struct {
	ID             int            `db:"id" json:"id"`
	Name           string         `db:"name" json:"name"`
	Vendor         string         `db:"vendor" json:"vendor"`
	IPAddress      string         `db:"ip_address" json:"ip_address"`
	Port           int            `db:"port" json:"port"`
	LocationID     sql.NullInt64  `db:"location_id" json:"location_id"`
	Vdom           sql.NullString `db:"vdom" json:"vdom"`
	AuthToken      sql.NullString `db:"auth_token" json:"-"`
	SSHUsername    sql.NullString `db:"ssh_username" json:"-"`
	SSHPassword    sql.NullString `db:"ssh_password" json:"-"`
	EnablePassword sql.NullString `db:"enable_password" json:"-"`
	Platform       sql.NullString `db:"platform" json:"platform"`
	ScheduleCron   sql.NullString `db:"schedule_cron" json:"schedule_cron"`
	IsActive       bool           `db:"is_active" json:"is_active"`
	CreatedAt      string         `db:"created_at" json:"created_at"`
	UpdatedAt      string         `db:"updated_at" json:"updated_at"`
}

type Backup struct {
	ID           int            `db:"id" json:"id"`
	DeviceID     int            `db:"device_id" json:"device_id"`
	FilePath     string         `db:"file_path" json:"file_path"`
	FileSize     int            `db:"file_size" json:"file_size"`
	Status       string         `db:"status" json:"status"`
	ErrorMessage sql.NullString `db:"error_message" json:"error_message"`
	TriggeredBy  string         `db:"triggered_by" json:"triggered_by"`
	CreatedAt    string         `db:"created_at" json:"created_at"`
}

type AuditLog struct {
	ID           int            `db:"id" json:"id"`
	UserID       sql.NullInt64  `db:"user_id" json:"user_id"`
	Username     string         `db:"username" json:"username"`
	Action       string         `db:"action" json:"action"`
	ResourceType string         `db:"resource_type" json:"resource_type"`
	ResourceName sql.NullString `db:"resource_name" json:"resource_name"`
	Detail       sql.NullString `db:"detail" json:"detail"`
	IPAddress    sql.NullString `db:"ip_address" json:"ip_address"`
	CreatedAt    string         `db:"created_at" json:"created_at"`
}

type Setting struct {
	Key   string         `db:"key" json:"key"`
	Value sql.NullString `db:"value" json:"value"`
}

// Response DTOs

type UserOut struct {
	ID                 int    `json:"id"`
	Username           string `json:"username"`
	Role               string `json:"role"`
	TOTPEnabled        bool   `json:"totp_enabled"`
	MustChangePassword bool   `json:"must_change_password"`
	CreatedAt          string `json:"created_at"`
}

type LocationOut struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	DeviceCount int     `json:"device_count"`
	CreatedAt   string  `json:"created_at"`
}

type DeviceOut struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	Vendor           string  `json:"vendor"`
	IPAddress        string  `json:"ip_address"`
	Port             int     `json:"port"`
	LocationID       *int    `json:"location_id"`
	LocationName     *string `json:"location_name"`
	Vdom             *string `json:"vdom"`
	Platform         *string `json:"platform"`
	ScheduleCron     *string `json:"schedule_cron"`
	IsActive         bool    `json:"is_active"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
	LastBackup       *string `json:"last_backup"`
	BackupCount      int     `json:"backup_count"`
	FailedCount      int     `json:"failed_count"`
	SSHUsername       *string `json:"ssh_username"`
	HasToken         bool    `json:"has_token"`
	HasSSHPassword   bool    `json:"has_ssh_password"`
	HasEnablePassword bool   `json:"has_enable_password"`
}

type BackupOut struct {
	ID           int     `json:"id"`
	DeviceID     int     `json:"device_id"`
	DeviceName   string  `json:"device_name"`
	Vendor       string  `json:"vendor"`
	LocationName *string `json:"location_name"`
	FilePath     string  `json:"file_path"`
	FileSize     int     `json:"file_size"`
	Status       string  `json:"status"`
	ErrorMessage *string `json:"error_message"`
	TriggeredBy  string  `json:"triggered_by"`
	CreatedAt    string  `json:"created_at"`
}

type DashboardStats struct {
	TotalDevices         int                    `json:"total_devices"`
	ActiveDevices        int                    `json:"active_devices"`
	TotalBackups         int                    `json:"total_backups"`
	SuccessfulBackups    int                    `json:"successful_backups"`
	FailedBackups        int                    `json:"failed_backups"`
	TodayBackups         int                    `json:"today_backups"`
	TodayFailed          int                    `json:"today_failed"`
	SuccessRate          float64                `json:"success_rate"`
	TotalBackupSize      int64                  `json:"total_backup_size"`
	VendorDistribution   map[string]int         `json:"vendor_distribution"`
	LocationDistribution map[string]int         `json:"location_distribution"`
	RecentActivities     []map[string]any       `json:"recent_activities"`
	ScheduledDevices     int                    `json:"scheduled_devices"`
}

type AuditLogOut struct {
	ID           int     `json:"id"`
	Username     string  `json:"username"`
	Action       string  `json:"action"`
	ResourceType string  `json:"resource_type"`
	ResourceName *string `json:"resource_name"`
	Detail       *string `json:"detail"`
	IPAddress    *string `json:"ip_address"`
	CreatedAt    string  `json:"created_at"`
}

