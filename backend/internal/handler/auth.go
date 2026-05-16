package handler

import (
	"net/http"

	"github.com/yunuskargi/confbox/internal/auth"
	"github.com/yunuskargi/confbox/internal/database"
	"github.com/yunuskargi/confbox/internal/models"
	"github.com/yunuskargi/confbox/internal/service"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	TOTPCode string `json:"totp_code"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Role        string `json:"role"`
	Requires2FA bool   `json:"requires_2fa"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}

	var user models.User
	err := database.DB.Get(&user, "SELECT * FROM users WHERE username = ?", body.Username)
	if err != nil || !auth.VerifyPassword(body.Password, user.PasswordHash) {
		service.LogAction(nil, body.Username, "login_failed", "auth", body.Username, "Failed login attempt", clientIP(r))
		writeError(w, 401, "Invalid credentials")
		return
	}

	if user.TOTPEnabled && user.TOTPSecret.Valid {
		if body.TOTPCode == "" {
			writeJSON(w, 200, tokenResponse{AccessToken: "", TokenType: "bearer", Role: user.Role, Requires2FA: true})
			return
		}
		if !auth.ValidateTOTP(body.TOTPCode, user.TOTPSecret.String) {
			uid := user.ID
			service.LogAction(&uid, user.Username, "login_failed", "auth", user.Username, "Invalid 2FA code", clientIP(r))
			writeError(w, 401, "Invalid 2FA code")
			return
		}
	}

	token, err := auth.CreateToken(user.Username, user.Role)
	if err != nil {
		writeError(w, 500, "Failed to create token")
		return
	}

	uid := user.ID
	service.LogAction(&uid, user.Username, "login", "auth", user.Username, "Successful login", clientIP(r))
	writeJSON(w, 200, tokenResponse{AccessToken: token, TokenType: "bearer", Role: user.Role, Requires2FA: false})
}

func Me(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	writeJSON(w, 200, models.UserOut{
		ID:          user.ID,
		Username:    user.Username,
		Role:        user.Role,
		TOTPEnabled: user.TOTPEnabled,
		CreatedAt:   user.CreatedAt,
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	token := auth.GetToken(r)
	auth.BlacklistToken(token)
	writeJSON(w, 200, map[string]string{"message": "Logged out"})
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	var body changePasswordRequest
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}

	if !auth.VerifyPassword(body.CurrentPassword, user.PasswordHash) {
		writeError(w, 400, "Current password is incorrect")
		return
	}
	if msg := auth.ValidatePassword(body.NewPassword); msg != "" {
		writeError(w, 400, msg)
		return
	}
	hash, _ := auth.HashPassword(body.NewPassword)
	database.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ?", hash, user.ID)
	writeJSON(w, 200, map[string]string{"message": "Password changed"})
}

func Setup2FA(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.TOTPEnabled {
		writeError(w, 400, "2FA already active")
		return
	}

	secret, qrBase64, err := auth.GenerateTOTP(user.Username)
	if err != nil {
		writeError(w, 500, "Failed to generate 2FA")
		return
	}

	database.DB.Exec("UPDATE users SET totp_secret = ? WHERE id = ?", secret, user.ID)
	writeJSON(w, 200, map[string]string{"secret": secret, "qr_code": qrBase64})
}

type verify2FARequest struct {
	Code string `json:"code"`
}

func Verify2FA(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	var body verify2FARequest
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}

	if !user.TOTPSecret.Valid {
		writeError(w, 400, "Setup 2FA first")
		return
	}
	if !auth.ValidateTOTP(body.Code, user.TOTPSecret.String) {
		writeError(w, 400, "Invalid code")
		return
	}
	database.DB.Exec("UPDATE users SET totp_enabled = 1 WHERE id = ?", user.ID)
	writeJSON(w, 200, map[string]string{"message": "2FA activated"})
}

func Disable2FA(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	var body changePasswordRequest
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}
	if !auth.VerifyPassword(body.CurrentPassword, user.PasswordHash) {
		writeError(w, 400, "Incorrect password")
		return
	}
	database.DB.Exec("UPDATE users SET totp_secret = NULL, totp_enabled = 0 WHERE id = ?", user.ID)
	writeJSON(w, 200, map[string]string{"message": "2FA disabled"})
}
