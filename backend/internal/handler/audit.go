package handler

import (
	"net/http"

	"github.com/yunuskargi/confbox/internal/auth"
	"github.com/yunuskargi/confbox/internal/database"
	"github.com/yunuskargi/confbox/internal/models"
)

func ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}

	action := queryStr(r, "action")
	resourceType := queryStr(r, "resource_type")
	username := queryStr(r, "username")
	limit := queryInt(r, "limit", 50)
	offset := queryInt(r, "offset", 0)

	if limit > 500 {
		limit = 500
	}

	query := "SELECT id, username, action, resource_type, resource_name, detail, ip_address, created_at FROM audit_logs WHERE 1=1"
	args := []any{}

	if action != "" {
		query += " AND action = ?"
		args = append(args, action)
	}
	if resourceType != "" {
		query += " AND resource_type = ?"
		args = append(args, resourceType)
	}
	if username != "" {
		query += " AND username = ?"
		args = append(args, username)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		writeError(w, 500, "Database error")
		return
	}
	defer rows.Close()

	var result []models.AuditLogOut
	for rows.Next() {
		var log models.AuditLogOut
		rows.Scan(&log.ID, &log.Username, &log.Action, &log.ResourceType, &log.ResourceName, &log.Detail, &log.IPAddress, &log.CreatedAt)
		result = append(result, log)
	}
	if result == nil {
		result = []models.AuditLogOut{}
	}
	writeJSON(w, 200, result)
}
