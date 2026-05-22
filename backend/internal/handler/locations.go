package handler

import (
	"net/http"

	"github.com/yunuskargi/configbox/internal/auth"
	"github.com/yunuskargi/configbox/internal/config"
	"github.com/yunuskargi/configbox/internal/database"
	"github.com/yunuskargi/configbox/internal/models"
	"github.com/yunuskargi/configbox/internal/service"
)

func ListLocations(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(`
		SELECT l.id, l.name, l.description, l.created_at, COUNT(d.id) as device_count
		FROM locations l LEFT JOIN devices d ON d.location_id = l.id
		GROUP BY l.id ORDER BY l.name`)
	if err != nil {
		writeError(w, 500, "Database error")
		return
	}
	defer rows.Close()

	var result []models.LocationOut
	for rows.Next() {
		var loc models.LocationOut
		var desc *string
		rows.Scan(&loc.ID, &loc.Name, &desc, &loc.CreatedAt, &loc.DeviceCount)
		loc.Description = desc
		result = append(result, loc)
	}
	if result == nil {
		result = []models.LocationOut{}
	}
	writeJSON(w, 200, result)
}

type locationBody struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func CreateLocation(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	var body locationBody
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}

	var exists int
	database.DB.Get(&exists, "SELECT COUNT(*) FROM locations WHERE name = ?", body.Name)
	if exists > 0 {
		writeError(w, 400, "Location name already exists")
		return
	}

	res, _ := database.DB.Exec("INSERT INTO locations (name, description, created_at) VALUES (?, ?, ?)", body.Name, body.Description, config.Now())
	id, _ := res.LastInsertId()

	var loc models.LocationOut
	database.DB.QueryRow(`
		SELECT l.id, l.name, l.description, l.created_at, COUNT(d.id)
		FROM locations l LEFT JOIN devices d ON d.location_id = l.id
		WHERE l.id = ? GROUP BY l.id`, id).Scan(&loc.ID, &loc.Name, &loc.Description, &loc.CreatedAt, &loc.DeviceCount)

	uid := user.ID
	service.LogAction(&uid, user.Username, "create", "location", body.Name, "", clientIP(r))

	writeJSON(w, 201, loc)
}

func UpdateLocation(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	id := paramInt(r, "id")
	var body locationBody
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}

	var exists int
	database.DB.Get(&exists, "SELECT COUNT(*) FROM locations WHERE id = ?", id)
	if exists == 0 {
		writeError(w, 404, "Location not found")
		return
	}

	if body.Name != "" {
		database.DB.Exec("UPDATE locations SET name = ? WHERE id = ?", body.Name, id)
	}
	if body.Description != nil {
		database.DB.Exec("UPDATE locations SET description = ? WHERE id = ?", *body.Description, id)
	}

	var loc models.LocationOut
	database.DB.QueryRow(`
		SELECT l.id, l.name, l.description, l.created_at, COUNT(d.id)
		FROM locations l LEFT JOIN devices d ON d.location_id = l.id
		WHERE l.id = ? GROUP BY l.id`, id).Scan(&loc.ID, &loc.Name, &loc.Description, &loc.CreatedAt, &loc.DeviceCount)

	uid := user.ID
	service.LogAction(&uid, user.Username, "update", "location", loc.Name, "", clientIP(r))

	writeJSON(w, 200, loc)
}

func DeleteLocation(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.Role != "admin" {
		writeError(w, 403, "Admin only")
		return
	}

	id := paramInt(r, "id")
	var exists int
	database.DB.Get(&exists, "SELECT COUNT(*) FROM locations WHERE id = ?", id)
	if exists == 0 {
		writeError(w, 404, "Location not found")
		return
	}

	var locName string
	database.DB.Get(&locName, "SELECT name FROM locations WHERE id = ?", id)

	database.DB.Exec("UPDATE devices SET location_id = NULL WHERE location_id = ?", id)
	database.DB.Exec("DELETE FROM locations WHERE id = ?", id)

	uid := user.ID
	service.LogAction(&uid, user.Username, "delete", "location", locName, "", clientIP(r))

	writeJSON(w, 200, map[string]string{"message": "Location deleted"})
}
