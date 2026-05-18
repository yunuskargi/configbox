package handler

import (
	"net/http"

	"github.com/yunuskargi/confbox/internal/auth"
	"github.com/yunuskargi/confbox/internal/config"
	"github.com/yunuskargi/confbox/internal/database"
	"github.com/yunuskargi/confbox/internal/models"
)

func ListUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	database.DB.Select(&users, "SELECT * FROM users ORDER BY username")
	result := make([]models.UserOut, len(users))
	for i, u := range users {
		result[i] = models.UserOut{ID: u.ID, Username: u.Username, Role: u.Role, TOTPEnabled: u.TOTPEnabled, CreatedAt: u.CreatedAt}
	}
	writeJSON(w, 200, result)
}

type userCreateBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var body userCreateBody
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}
	if body.Role == "" {
		body.Role = "backup_admin"
	}
	if body.Role != "admin" && body.Role != "backup_admin" {
		writeError(w, 400, "Invalid role")
		return
	}

	var exists int
	database.DB.Get(&exists, "SELECT COUNT(*) FROM users WHERE username = ?", body.Username)
	if exists > 0 {
		writeError(w, 400, "Username already exists")
		return
	}

	if msg := auth.ValidatePassword(body.Password); msg != "" {
		writeError(w, 400, msg)
		return
	}

	hash, _ := auth.HashPassword(body.Password)
	res, _ := database.DB.Exec("INSERT INTO users (username, password_hash, role, created_at) VALUES (?, ?, ?, ?)", body.Username, hash, body.Role, config.Now())
	id, _ := res.LastInsertId()

	var user models.User
	database.DB.Get(&user, "SELECT * FROM users WHERE id = ?", id)
	writeJSON(w, 201, models.UserOut{ID: user.ID, Username: user.Username, Role: user.Role, TOTPEnabled: user.TOTPEnabled, CreatedAt: user.CreatedAt})
}

type userUpdateBody struct {
	Username *string `json:"username"`
	Password *string `json:"password"`
	Role     *string `json:"role"`
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := paramInt(r, "id")
	var body userUpdateBody
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}

	var exists int
	database.DB.Get(&exists, "SELECT COUNT(*) FROM users WHERE id = ?", id)
	if exists == 0 {
		writeError(w, 404, "User not found")
		return
	}

	if body.Username != nil {
		var dup int
		database.DB.Get(&dup, "SELECT COUNT(*) FROM users WHERE username = ? AND id != ?", *body.Username, id)
		if dup > 0 {
			writeError(w, 400, "Username already exists")
			return
		}
		database.DB.Exec("UPDATE users SET username = ? WHERE id = ?", *body.Username, id)
	}
	if body.Role != nil {
		if *body.Role != "admin" && *body.Role != "backup_admin" {
			writeError(w, 400, "Invalid role")
			return
		}
		database.DB.Exec("UPDATE users SET role = ? WHERE id = ?", *body.Role, id)
	}
	if body.Password != nil {
		if msg := auth.ValidatePassword(*body.Password); msg != "" {
			writeError(w, 400, msg)
			return
		}
		hash, _ := auth.HashPassword(*body.Password)
		database.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ?", hash, id)
	}

	var user models.User
	database.DB.Get(&user, "SELECT * FROM users WHERE id = ?", id)
	writeJSON(w, 200, models.UserOut{ID: user.ID, Username: user.Username, Role: user.Role, TOTPEnabled: user.TOTPEnabled, CreatedAt: user.CreatedAt})
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	currentUser := auth.GetUser(r)
	id := paramInt(r, "id")

	if id == currentUser.ID {
		writeError(w, 400, "Cannot delete yourself")
		return
	}

	var exists int
	database.DB.Get(&exists, "SELECT COUNT(*) FROM users WHERE id = ?", id)
	if exists == 0 {
		writeError(w, 404, "User not found")
		return
	}

	database.DB.Exec("DELETE FROM users WHERE id = ?", id)
	writeJSON(w, 200, map[string]string{"message": "User deleted"})
}
