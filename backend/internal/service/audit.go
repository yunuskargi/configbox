package service

import (
	"github.com/yunuskargi/confbox/internal/database"
)

func LogAction(userID *int, username, action, resourceType, resourceName, detail, ipAddress string) {
	database.DB.Exec(
		`INSERT INTO audit_logs (user_id, username, action, resource_type, resource_name, detail, ip_address, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
		userID, username, action, resourceType, resourceName, detail, ipAddress,
	)
}
