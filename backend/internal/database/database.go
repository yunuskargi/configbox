package database

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/yunuskargi/confbox/internal/config"
)

var DB *sqlx.DB

func Open() {
	var err error
	DB, err = sqlx.Open("sqlite3", config.DatabasePath+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		slog.Error("failed to open database", "error", err)
		panic(err)
	}
	DB.SetMaxOpenConns(1)

	createTables()
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func createTables() {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'admin',
		totp_secret TEXT,
		totp_enabled INTEGER DEFAULT 0,
		created_at TEXT DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS locations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		created_at TEXT DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS devices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		vendor TEXT NOT NULL,
		ip_address TEXT NOT NULL,
		port INTEGER NOT NULL,
		location_id INTEGER REFERENCES locations(id),
		vdom TEXT,
		auth_token TEXT,
		ssh_username TEXT,
		ssh_password TEXT,
		enable_password TEXT,
		platform TEXT DEFAULT 'ios',
		schedule_cron TEXT,
		is_active INTEGER DEFAULT 1,
		created_at TEXT DEFAULT (datetime('now')),
		updated_at TEXT DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS backups (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id INTEGER NOT NULL REFERENCES devices(id),
		file_path TEXT NOT NULL,
		file_size INTEGER DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'success',
		error_message TEXT,
		triggered_by TEXT NOT NULL DEFAULT 'manual',
		created_at TEXT DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER REFERENCES users(id),
		username TEXT NOT NULL,
		action TEXT NOT NULL,
		resource_type TEXT NOT NULL,
		resource_name TEXT,
		detail TEXT,
		ip_address TEXT,
		created_at TEXT DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT
	);

	CREATE TABLE IF NOT EXISTS token_blacklist (
		token_hash TEXT PRIMARY KEY,
		expires_at INTEGER NOT NULL
	);
	`
	DB.MustExec(schema)
}
