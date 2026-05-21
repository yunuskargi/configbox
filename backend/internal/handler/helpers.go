package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

var trustedProxy = os.Getenv("TRUSTED_PROXY")

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"detail": detail})
}

func decodeBody(r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20) // 1MB limit
	return json.NewDecoder(r.Body).Decode(v)
}

func paramInt(r *http.Request, name string) int {
	v, _ := strconv.Atoi(chi.URLParam(r, name))
	return v
}

func queryInt(r *http.Request, name string, def int) int {
	v := r.URL.Query().Get(name)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func queryStr(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

func sanitizeFilename(name string) string {
	// Remove control characters, quotes, and backslashes from filename
	var clean []byte
	for _, c := range []byte(name) {
		if c < 32 || c == '"' || c == '\\' || c == ';' {
			continue
		}
		clean = append(clean, c)
	}
	if len(clean) == 0 {
		return "download"
	}
	return string(clean)
}

func clientIP(r *http.Request) string {
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
