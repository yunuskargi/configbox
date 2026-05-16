package auth

import (
	"regexp"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/yunuskargi/confbox/internal/config"
)

var (
	blacklist   = make(map[string]int64)
	blacklistMu sync.RWMutex
	passwordRe  = regexp.MustCompile(`^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,}$`)
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
	if !passwordRe.MatchString(password) {
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

func BlacklistToken(tokenStr string) {
	blacklistMu.Lock()
	defer blacklistMu.Unlock()
	blacklist[tokenStr] = time.Now().Add(time.Duration(config.JWTExpireMin) * time.Minute).Unix()
}

func IsBlacklisted(tokenStr string) bool {
	blacklistMu.RLock()
	defer blacklistMu.RUnlock()
	_, ok := blacklist[tokenStr]
	return ok
}

func CleanupBlacklist() {
	blacklistMu.Lock()
	defer blacklistMu.Unlock()
	now := time.Now().Unix()
	for k, exp := range blacklist {
		if now > exp {
			delete(blacklist, k)
		}
	}
}
