package service

import (
	"github.com/yunuskargi/configbox/internal/config"
	"github.com/yunuskargi/configbox/internal/database"
)

func LogAction(userID *int, username, action, resourceType, resourceName, detail, ipAddress string) {
	database.DB.Exec(
		`INSERT INTO audit_logs (user_id, username, action, resource_type, resource_name, detail, ip_address, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, username, action, resourceType, resourceName, detail, ipAddress, config.Now(),
	)
}
