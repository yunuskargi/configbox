package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/yunuskargi/configbox/internal/config"
	"github.com/yunuskargi/configbox/internal/database"
)

func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h), err
}

func VerifyPassword(plain, hashed string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}

func ValidatePassword(password string) string {
	if len(password) < 8 {
		return "Password must be at least 8 characters"
	}
	var hasLower, hasUpper, hasDigit bool
	for _, c := range password {
		switch {
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}
	if !hasLower || !hasUpper || !hasDigit {
		return "Password must contain at least 1 uppercase, 1 lowercase, and 1 digit"
	}
	return ""
}

func CreateToken(username, role string) (string, error) {
	exp := time.Now().Add(time.Duration(config.JWTExpireMin) * time.Minute)
	claims := jwt.MapClaims{
		"sub":  username,
		"role": role,
		"exp":  exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}

func ParseToken(tokenStr string) (username string, role string, err error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return "", "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", jwt.ErrTokenMalformed
	}
	sub, _ := claims["sub"].(string)
	r, _ := claims["role"].(string)
	return sub, r, nil
}

func tokenHash(tokenStr string) string {
	h := sha256.Sum256([]byte(tokenStr))
	return hex.EncodeToString(h[:])
}

func BlacklistToken(tokenStr string) {
	exp := time.Now().Add(time.Duration(config.JWTExpireMin) * time.Minute).Unix()
	database.DB.Exec("INSERT OR IGNORE INTO token_blacklist (token_hash, expires_at) VALUES (?, ?)", tokenHash(tokenStr), exp)
}

func IsBlacklisted(tokenStr string) bool {
	var count int
	database.DB.Get(&count, "SELECT COUNT(*) FROM token_blacklist WHERE token_hash = ?", tokenHash(tokenStr))
	return count > 0
}

func CleanupBlacklist() {
	database.DB.Exec("DELETE FROM token_blacklist WHERE expires_at < ?", time.Now().Unix())
}
