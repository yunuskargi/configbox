package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	BackupDir       string
	DatabasePath    string
	JWTSecret       string
	JWTExpireMin    int
	TZOffset        int
	AppTimezone     *time.Location
	DefaultAdmin    string
	DefaultPassword string
	CORSOrigins     []string
)

func Load() {
	godotenv.Load()

	baseDir, _ := os.Getwd()

	BackupDir = getEnv("BACKUP_DIR", filepath.Join(baseDir, "..", "backups"))
	DatabasePath = getEnv("DATABASE_PATH", filepath.Join(baseDir, "confbox.db"))

	TZOffset, _ = strconv.Atoi(getEnv("TZ_OFFSET", "3"))
	AppTimezone = time.FixedZone("APP", TZOffset*3600)

	JWTSecret = getJWTSecret(baseDir)
	JWTExpireMin, _ = strconv.Atoi(getEnv("JWT_EXPIRE_MINUTES", "480"))

	DefaultAdmin = getEnv("DEFAULT_ADMIN_USER", "admin")
	DefaultPassword = getEnv("DEFAULT_ADMIN_PASS", "admin")

	origins := getEnv("CORS_ORIGINS", "")
	if origins != "" {
		for _, o := range strings.Split(origins, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				CORSOrigins = append(CORSOrigins, o)
			}
		}
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getJWTSecret(baseDir string) string {
	if env := os.Getenv("JWT_SECRET"); env != "" && env != "change-me-in-production" {
		return env
	}

	secretFile := filepath.Join(baseDir, ".jwt_secret")
	if data, err := os.ReadFile(secretFile); err == nil {
		return strings.TrimSpace(string(data))
	}

	buf := make([]byte, 32)
	rand.Read(buf)
	generated := hex.EncodeToString(buf)
	os.WriteFile(secretFile, []byte(generated), 0600)
	return generated
}
