package handler

import (
	"net/http"
	"time"

	"github.com/yunuskargi/configbox/internal/config"
	"github.com/yunuskargi/configbox/internal/database"
	"github.com/yunuskargi/configbox/internal/models"
)

func GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	var stats models.DashboardStats

	database.DB.Get(&stats.TotalDevices, "SELECT COUNT(*) FROM devices")
	database.DB.Get(&stats.ActiveDevices, "SELECT COUNT(*) FROM devices WHERE is_active = 1")
	database.DB.Get(&stats.TotalBackups, "SELECT COUNT(*) FROM backups")
	database.DB.Get(&stats.SuccessfulBackups, "SELECT COUNT(*) FROM backups WHERE status = 'success'")
	database.DB.Get(&stats.FailedBackups, "SELECT COUNT(*) FROM backups WHERE status = 'failed'")
	database.DB.Get(&stats.ScheduledDevices, "SELECT COUNT(*) FROM devices WHERE schedule_cron IS NOT NULL AND is_active = 1")

	var totalSize *int64
	database.DB.Get(&totalSize, "SELECT SUM(file_size) FROM backups WHERE status = 'success'")
	if totalSize != nil {
		stats.TotalBackupSize = *totalSize
	}

	todayStart := time.Now().In(config.AppTimezone).Format("2006-01-02") + " 00:00:00"
	database.DB.Get(&stats.TodayBackups, "SELECT COUNT(*) FROM backups WHERE created_at >= ?", todayStart)
	database.DB.Get(&stats.TodayFailed, "SELECT COUNT(*) FROM backups WHERE created_at >= ? AND status = 'failed'", todayStart)

	if stats.TotalBackups > 0 {
		stats.SuccessRate = float64(stats.SuccessfulBackups) / float64(stats.TotalBackups) * 100
		stats.SuccessRate = float64(int(stats.SuccessRate*10)) / 10
	}

	type vendorRow struct {
		Vendor string `db:"vendor"`
		Count  int    `db:"count"`
	}
	var vendors []vendorRow
	database.DB.Select(&vendors, "SELECT vendor, COUNT(*) as count FROM devices GROUP BY vendor")
	stats.VendorDistribution = make(map[string]int)
	for _, v := range vendors {
		stats.VendorDistribution[v.Vendor] = v.Count
	}

	type locRow struct {
		Name  string `db:"name"`
		Count int    `db:"count"`
	}
	var locs []locRow
	database.DB.Select(&locs, "SELECT l.name, COUNT(d.id) as count FROM locations l LEFT JOIN devices d ON d.location_id = l.id GROUP BY l.name")
	stats.LocationDistribution = make(map[string]int)
	for _, l := range locs {
		stats.LocationDistribution[l.Name] = l.Count
	}

	type recentRow struct {
		ID           int     `db:"id"`
		DeviceName   string  `db:"device_name"`
		Vendor       string  `db:"vendor"`
		Status       string  `db:"status"`
		TriggeredBy  string  `db:"triggered_by"`
		CreatedAt    string  `db:"created_at"`
		ErrorMessage *string `db:"error_message"`
		FileSize     int     `db:"file_size"`
	}
	var recent []recentRow
	database.DB.Select(&recent, `
		SELECT b.id, d.name as device_name, d.vendor, b.status, b.triggered_by, b.created_at, b.error_message, b.file_size
		FROM backups b JOIN devices d ON d.id = b.device_id
		ORDER BY b.created_at DESC LIMIT 10`)

	stats.RecentActivities = make([]map[string]any, len(recent))
	for i, r := range recent {
		stats.RecentActivities[i] = map[string]any{
			"id": r.ID, "device_name": r.DeviceName, "vendor": r.Vendor,
			"status": r.Status, "triggered_by": r.TriggeredBy,
			"created_at": r.CreatedAt, "error_message": r.ErrorMessage, "file_size": r.FileSize,
		}
	}

	writeJSON(w, 200, stats)
}

func GetBackupTrend(w http.ResponseWriter, r *http.Request) {
	days := queryInt(r, "days", 30)
	if days > 365 {
		days = 365
	}

	since := time.Now().In(config.AppTimezone).AddDate(0, 0, -days)
	sinceStr := since.Format("2006-01-02")

	type trendRow struct {
		Date   string `db:"date"`
		Status string `db:"status"`
		Count  int    `db:"count"`
	}
	var rows []trendRow
	database.DB.Select(&rows, `
		SELECT date(created_at) as date, status, COUNT(*) as count
		FROM backups WHERE created_at >= ? GROUP BY date(created_at), status`, sinceStr)

	dateMap := make(map[string]map[string]int)
	for _, r := range rows {
		if dateMap[r.Date] == nil {
			dateMap[r.Date] = map[string]int{}
		}
		dateMap[r.Date][r.Status] = r.Count
	}

	var result []map[string]any
	current := since
	now := time.Now().In(config.AppTimezone)
	for !current.After(now) {
		d := current.Format("2006-01-02")
		entry := map[string]any{"date": d, "success": 0, "failed": 0}
		if m, ok := dateMap[d]; ok {
			entry["success"] = m["success"]
			entry["failed"] = m["failed"]
		}
		result = append(result, entry)
		current = current.AddDate(0, 0, 1)
	}
	writeJSON(w, 200, result)
}

func GetSizeTrend(w http.ResponseWriter, r *http.Request) {
	days := queryInt(r, "days", 30)
	if days > 365 {
		days = 365
	}

	since := time.Now().In(config.AppTimezone).AddDate(0, 0, -days)
	sinceStr := since.Format("2006-01-02")

	type sizeRow struct {
		Date      string `db:"date"`
		TotalSize int64  `db:"total_size"`
	}
	var rows []sizeRow
	database.DB.Select(&rows, `
		SELECT date(created_at) as date, SUM(file_size) as total_size
		FROM backups WHERE created_at >= ? AND status = 'success'
		GROUP BY date(created_at)`, sinceStr)

	dateMap := make(map[string]int64)
	for _, r := range rows {
		dateMap[r.Date] = r.TotalSize
	}

	var result []map[string]any
	current := since
	now := time.Now().In(config.AppTimezone)
	var cumulative int64
	for !current.After(now) {
		d := current.Format("2006-01-02")
		daySize := dateMap[d]
		cumulative += daySize
		result = append(result, map[string]any{"date": d, "size": daySize, "cumulative": cumulative})
		current = current.AddDate(0, 0, 1)
	}

	writeJSON(w, 200, result)
}
