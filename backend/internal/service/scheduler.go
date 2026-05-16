package service

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/robfig/cron/v3"

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

func RemoveDeviceSchedule(deviceID int) {
	if scheduler == nil {
		return
	}
	if entryID, ok := jobIDs[deviceID]; ok {
		scheduler.Remove(entryID)
		delete(jobIDs, deviceID)
	}
}
