package service

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/yunuskargi/confbox/internal/config"
	"github.com/yunuskargi/confbox/internal/database"
)

var (
	scheduler *cron.Cron
	jobIDs    = make(map[int]cron.EntryID)
)

func StartScheduler() {
	scheduler = cron.New()

	type deviceRow struct {
		ID           int    `db:"id"`
		ScheduleCron string `db:"schedule_cron"`
	}
	var devices []deviceRow
	database.DB.Select(&devices, "SELECT id, schedule_cron FROM devices WHERE schedule_cron IS NOT NULL AND is_active = 1")

	for _, d := range devices {
		ScheduleDevice(d.ID, d.ScheduleCron)
	}

	scheduler.AddFunc("0 8 * * *", func() {
		SendDailySummary()
	})

	scheduler.AddFunc("0 3 * * *", func() {
		CleanupOldBackups()
		ArchiveOldBackups()
		// Cleanup expired download tokens (older than 1 hour)
		database.DB.Exec("DELETE FROM used_download_tokens WHERE used_at < datetime('now', '-1 hour')")
	})

	scheduler.Start()
	slog.Info("scheduler started", "devices", len(devices))
}

func ShutdownScheduler() {
	if scheduler != nil {
		scheduler.Stop()
	}
}

func ScheduleDevice(deviceID int, cronExpr string) {
	if scheduler == nil {
		return
	}

	RemoveDeviceSchedule(deviceID)

	parts := strings.Fields(cronExpr)
	if len(parts) != 5 {
		slog.Warn("invalid cron expression", "device_id", deviceID, "cron", cronExpr)
		return
	}

	id := deviceID
	entryID, err := scheduler.AddFunc(fmt.Sprintf("%s %s %s %s %s", parts[0], parts[1], parts[2], parts[3], parts[4]), func() {
		RunScheduledBackup(id)
	})
	if err != nil {
		slog.Error("failed to schedule device", "device_id", deviceID, "error", err)
		return
	}
	jobIDs[deviceID] = entryID
}

func CleanupOldBackups() {
	var val *string
	err := database.DB.Get(&val, "SELECT value FROM settings WHERE key = 'retention_days'")
	if err != nil || val == nil {
		return
	}
	days, _ := strconv.Atoi(*val)
	if days <= 0 {
		return
	}

	cutoff := time.Now().In(config.AppTimezone).AddDate(0, 0, -days).Format("2006-01-02 15:04:05")
	type row struct {
		ID       int    `db:"id"`
		FilePath string `db:"file_path"`
	}
	var old []row
	database.DB.Select(&old, "SELECT id, file_path FROM backups WHERE created_at < ?", cutoff)

	if len(old) == 0 {
		return
	}

	for _, b := range old {
		if b.FilePath != "" {
			os.Remove(b.FilePath)
		}
		database.DB.Exec("DELETE FROM backups WHERE id = ?", b.ID)
	}
	slog.Info("retention cleanup", "deleted", len(old), "retention_days", days)
}

func ArchiveOldBackups() {
	var val *string
	err := database.DB.Get(&val, "SELECT value FROM settings WHERE key = 'archive_after_days'")
	if err != nil || val == nil {
		return
	}
	days, _ := strconv.Atoi(*val)
	if days <= 0 {
		return
	}

	var enabled *string
	database.DB.Get(&enabled, "SELECT value FROM settings WHERE key = 'archive_enabled'")
	if enabled == nil || *enabled != "true" {
		return
	}

	cutoff := time.Now().In(config.AppTimezone).AddDate(0, 0, -days).Format("2006-01-02 15:04:05")
	type row struct {
		ID       int    `db:"id"`
		FilePath string `db:"file_path"`
	}
	var candidates []row
	database.DB.Select(&candidates, "SELECT id, file_path FROM backups WHERE created_at < ? AND status = 'success'", cutoff)

	archived := 0
	for _, b := range candidates {
		if b.FilePath == "" || strings.HasSuffix(b.FilePath, ".gz") {
			continue
		}
		if _, err := os.Stat(b.FilePath); os.IsNotExist(err) {
			continue
		}

		data, err := os.ReadFile(b.FilePath)
		if err != nil {
			continue
		}

		gzPath := b.FilePath + ".gz"
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		gw.Write(data)
		gw.Close()

		if err := os.WriteFile(gzPath, buf.Bytes(), 0600); err != nil {
			continue
		}

		os.Remove(b.FilePath)
		database.DB.Exec("UPDATE backups SET file_path = ?, file_size = ? WHERE id = ?", gzPath, buf.Len(), b.ID)
		archived++
	}

	if archived > 0 {
		slog.Info("archive completed", "archived", archived, "older_than_days", days)
	}
}

func RemoveDeviceSchedule(deviceID int) {
	if scheduler == nil {
		return
	}
	if entryID, ok := jobIDs[deviceID]; ok {
		scheduler.Remove(entryID)
		delete(jobIDs, deviceID)
	}
}
