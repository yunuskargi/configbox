package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/yunuskargi/configbox/internal/database"
	"github.com/yunuskargi/configbox/internal/models"
)

type contextKey string

const (
	UserContextKey  contextKey = "user"
	TokenContextKey contextKey = "token"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, `{"detail":"Missing authorization"}`, http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if IsBlacklisted(tokenStr) {
			http.Error(w, `{"detail":"Token revoked"}`, http.StatusUnauthorized)
			return
		}

		username, _, err := ParseToken(tokenStr)
		if err != nil || username == "" {
			http.Error(w, `{"detail":"Invalid token"}`, http.StatusUnauthorized)
			return
		}

		var user models.User
		err = database.DB.Get(&user, "SELECT * FROM users WHERE username = ?", username)
		if err != nil {
			http.Error(w, `{"detail":"User not found"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, &user)
		ctx = context.WithValue(ctx, TokenContextKey, tokenStr)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r)
		if user == nil || user.Role != "admin" {
			http.Error(w, `{"detail":"Admin access required"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetUser(r *http.Request) *models.User {
	u, _ := r.Context().Value(UserContextKey).(*models.User)
	return u
}

func GetToken(r *http.Request) string {
	t, _ := r.Context().Value(TokenContextKey).(string)
	return t
}
