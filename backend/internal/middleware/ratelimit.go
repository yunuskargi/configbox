package middleware

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var trustedProxy = os.Getenv("TRUSTED_PROXY")

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

func init() {
	go cleanupVisitors()
}

func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for key, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, key)
			}
		}
		mu.Unlock()
	}
}

func getVisitor(key string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[key]
	if !exists {
		limiter := rate.NewLimiter(rate.Every(time.Minute/5), 5)
		visitors[key] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func extractIP(r *http.Request) string {
	// Only trust proxy headers if TRUSTED_PROXY is configured
	if trustedProxy != "" {
		remoteIP := r.RemoteAddr
		if idx := strings.LastIndex(remoteIP, ":"); idx != -1 {
			remoteIP = remoteIP[:idx]
		}
		if remoteIP == trustedProxy || remoteIP == "127.0.0.1" || remoteIP == "::1" {
			if ip := r.Header.Get("X-Real-IP"); ip != "" {
				return ip
			}
			if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
				return strings.TrimSpace(strings.Split(ip, ",")[0])
			}
		}
	}
	return r.RemoteAddr
}

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)

		if !getVisitor("ip:"+ip).Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(429)
			w.Write([]byte(`{"detail":"Too many requests. Please wait."}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
