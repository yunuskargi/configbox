package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/yunuskargi/confbox/internal/auth"
	"github.com/yunuskargi/confbox/internal/config"
	"github.com/yunuskargi/confbox/internal/database"
	"github.com/yunuskargi/confbox/internal/models"
	"github.com/yunuskargi/confbox/internal/service"
)

func isPathSafe(filePath string) bool {
	abs, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}
	backupAbs, _ := filepath.Abs(config.BackupDir)
	return strings.HasPrefix(abs, backupAbs)
}

const downloadTokenTTL = 300

func makeDownloadToken(userID, backupID int) string {
	ts := time.Now().Unix()
	msg := fmt.Sprintf("%d:%d:%d", userID, backupID, ts)
	mac := hmac.New(sha256.New, []byte(config.JWTSecret))
	mac.Write([]byte(msg))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s:%s", msg, sig)
}

func verifyDownloadToken(token string, backupID int) bool {
	parts := strings.Split(token, ":")
	if len(parts) != 4 {
		return false
	}
	msg := strings.Join(parts[:3], ":")
	sig := parts[3]

	bid, err := strconv.Atoi(parts[1])
	if err != nil || bid != backupID {
		return false
	}

	ts, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || time.Now().Unix()-ts > downloadTokenTTL {
		return false
	}

	mac := hmac.New(sha256.New, []byte(config.JWTSecret))
	mac.Write([]byte(msg))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}

func ListBackups(w http.ResponseWriter, r *http.Request) {
	deviceID := queryStr(r, "device_id")
	vendor := queryStr(r, "vendor")
	status := queryStr(r, "status")
	limit := queryInt(r, "limit", 50)
	offset := queryInt(r, "offset", 0)
	if limit > 500 {
		limit = 500
	}

	query := `SELECT b.id, b.device_id, b.file_path, b.file_size, b.status, b.error_message, b.triggered_by, b.created_at,
	          d.name as device_name, d.vendor, l.name as location_name
	          FROM backups b JOIN devices d ON d.id = b.device_id LEFT JOIN locations l ON l.id = d.location_id WHERE 1=1`
	args := []any{}

	if deviceID != "" {
		query += " AND b.device_id = ?"
		args = append(args, deviceID)
	}
	if vendor != "" {
		query += " AND d.vendor = ?"
		args = append(args, vendor)
	}
	if status != "" {
		query += " AND b.status = ?"
		args = append(args, status)
	}

	query += " ORDER BY b.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		writeError(w, 500, "Database error")
		return
	}
	defer rows.Close()

	var result []models.BackupOut
	for rows.Next() {
		var b models.BackupOut
		rows.Scan(&b.ID, &b.DeviceID, &b.FilePath, &b.FileSize, &b.Status, &b.ErrorMessage,
			&b.TriggeredBy, &b.CreatedAt, &b.DeviceName, &b.Vendor, &b.LocationName)
		result = append(result, b)
	}
	if result == nil {
		result = []models.BackupOut{}
	}
	writeJSON(w, 200, result)
}

type downloadAuthBody struct {
	Password string `json:"password"`
	BackupID int    `json:"backup_id"`
}

func AuthorizeDownload(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	var body downloadAuthBody
	if err := decodeBody(r, &body); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}

	if !auth.VerifyPassword(body.Password, user.PasswordHash) {
		writeError(w, 403, "Incorrect password")
		return
	}

	var exists int
	database.DB.Get(&exists, "SELECT COUNT(*) FROM backups WHERE id = ?", body.BackupID)
	if exists == 0 {
		writeError(w, 404, "Backup not found")
		return
	}

	token := makeDownloadToken(user.ID, body.BackupID)

	var fileName string
	database.DB.Get(&fileName, "SELECT file_path FROM backups WHERE id = ?", body.BackupID)
	uid := user.ID
	service.LogAction(&uid, user.Username, "download", "backup", filepath.Base(fileName), "", clientIP(r))

	writeJSON(w, 200, map[string]string{"download_token": token})
}

func DownloadBackup(w http.ResponseWriter, r *http.Request) {
	id := paramInt(r, "id")
	token := queryStr(r, "token")

	if !verifyDownloadToken(token, id) {
		writeError(w, 403, "Invalid or expired download token")
		return
	}

	var filePath string
	err := database.DB.Get(&filePath, "SELECT file_path FROM backups WHERE id = ?", id)
	if err != nil {
		writeError(w, 404, "Backup not found")
		return
	}

	if !isPathSafe(filePath) {
		writeError(w, 403, "Access denied")
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		writeError(w, 404, "Backup file not found on disk")
		return
	}

	if strings.HasSuffix(filePath, ".gz") {
		content, err := service.ReadBackupFile(filePath)
		if err != nil {
			writeError(w, 500, "Failed to read backup file")
			return
		}
		origName := strings.TrimSuffix(filepath.Base(filePath), ".gz")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", origName))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(content)
	} else {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(filePath)))
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeFile(w, r, filePath)
	}
}

const maxContentViewSize = 10 * 1024 * 1024 // 10MB

func GetBackupContent(w http.ResponseWriter, r *http.Request) {
	id := paramInt(r, "id")

	var filePath string
	var fileSize int
	err := database.DB.QueryRow("SELECT file_path, file_size FROM backups WHERE id = ?", id).Scan(&filePath, &fileSize)
	if err != nil {
		writeError(w, 404, "Backup not found")
		return
	}

	if !isPathSafe(filePath) {
		writeError(w, 403, "Access denied")
		return
	}

	if fileSize > maxContentViewSize {
		writeError(w, 413, "File too large for inline viewing. Use download instead.")
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		writeError(w, 404, "Backup file not found on disk")
		return
	}

	content, _ := service.ReadBackupFile(filePath)
	writeJSON(w, 200, map[string]any{"content": string(content), "file_path": filePath, "file_size": fileSize})
}

func DiffBackups(w http.ResponseWriter, r *http.Request) {
	idA := paramInt(r, "idA")
	idB := paramInt(r, "idB")

	var pathA, pathB string
	var nameA, nameB, createdA, createdB string

	err := database.DB.QueryRow(`SELECT b.file_path, d.name, b.created_at FROM backups b JOIN devices d ON d.id = b.device_id WHERE b.id = ?`, idA).Scan(&pathA, &nameA, &createdA)
	if err != nil {
		writeError(w, 404, "Backup not found")
		return
	}
	err = database.DB.QueryRow(`SELECT b.file_path, d.name, b.created_at FROM backups b JOIN devices d ON d.id = b.device_id WHERE b.id = ?`, idB).Scan(&pathB, &nameB, &createdB)
	if err != nil {
		writeError(w, 404, "Backup not found")
		return
	}

	if !isPathSafe(pathA) || !isPathSafe(pathB) {
		writeError(w, 403, "Access denied")
		return
	}
	if _, err := os.Stat(pathA); os.IsNotExist(err) {
		writeError(w, 404, "Backup file not found on disk")
		return
	}
	if _, err := os.Stat(pathB); os.IsNotExist(err) {
		writeError(w, 404, "Backup file not found on disk")
		return
	}

	contentA, _ := service.ReadBackupFile(pathA)
	contentB, _ := service.ReadBackupFile(pathB)

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(contentA)),
		B:        difflib.SplitLines(string(contentB)),
		FromFile: filepath.Base(pathA),
		ToFile:   filepath.Base(pathB),
		Context:  3,
	}
	diffText, _ := difflib.GetUnifiedDiffString(diff)
	diffLines := strings.Split(strings.TrimRight(diffText, "\n"), "\n")
	if diffText == "" {
		diffLines = []string{}
	}

	added, removed := 0, 0
	for _, l := range diffLines {
		if strings.HasPrefix(l, "+") && !strings.HasPrefix(l, "+++") {
			added++
		}
		if strings.HasPrefix(l, "-") && !strings.HasPrefix(l, "---") {
			removed++
		}
	}

	writeJSON(w, 200, map[string]any{
		"diff":  diffLines,
		"stats": map[string]int{"added": added, "removed": removed},
		"backup_a": map[string]any{"id": idA, "device_name": nameA, "created_at": createdA},
		"backup_b": map[string]any{"id": idB, "device_name": nameB, "created_at": createdB},
	})
}

func DeleteBackup(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	id := paramInt(r, "id")

	var filePath string
	err := database.DB.Get(&filePath, "SELECT file_path FROM backups WHERE id = ?", id)
	if err != nil {
		writeError(w, 404, "Backup not found")
		return
	}

	if isPathSafe(filePath) {
		os.Remove(filePath)
	}
	database.DB.Exec("DELETE FROM backups WHERE id = ?", id)
	uid := user.ID
	service.LogAction(&uid, user.Username, "delete", "backup", filepath.Base(filePath), "", clientIP(r))
	writeJSON(w, 200, map[string]string{"message": "Backup deleted"})
}
