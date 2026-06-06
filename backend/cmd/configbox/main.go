package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yunuskargi/configbox/internal/auth"
	"github.com/yunuskargi/configbox/internal/config"
	"github.com/yunuskargi/configbox/internal/crypto"
	"github.com/yunuskargi/configbox/internal/database"
	"github.com/yunuskargi/configbox/internal/router"
	"github.com/yunuskargi/configbox/internal/service"
)

func main() {
	config.Load()
	database.Open()
	defer database.Close()

	crypto.Init()

	if len(os.Args) >= 4 && os.Args[1] == "reset-password" {
		resetPassword(os.Args[2], os.Args[3])
		return
	}

	seedAdmin()
	os.MkdirAll(config.BackupDir, 0755)
	service.StartScheduler()
	defer service.ShutdownScheduler()

	auth.CleanupBlacklist() // cleanup on startup
	go func() {
		for {
			time.Sleep(30 * time.Minute)
			auth.CleanupBlacklist()
		}
	}()

	handler := router.New()
	server := &http.Server{
		Addr:         ":8000",
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 180 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("server starting", "addr", ":8000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func seedAdmin() {
	var count int
	database.DB.Get(&count, "SELECT COUNT(*) FROM users WHERE username = ?", config.DefaultAdmin)
	if count > 0 {
		return
	}
	hash, _ := auth.HashPassword(config.DefaultPassword)
	database.DB.Exec("INSERT INTO users (username, password_hash, role, must_change_password, created_at) VALUES (?, ?, 'admin', 1, datetime('now'))",
		config.DefaultAdmin, hash)
	slog.Info("admin user created", "username", config.DefaultAdmin)
}

func resetPassword(username, newPassword string) {
	if msg := auth.ValidatePassword(newPassword); msg != "" {
		fmt.Println("Error:", msg)
		os.Exit(1)
	}
	var exists int
	database.DB.Get(&exists, "SELECT COUNT(*) FROM users WHERE username = ?", username)
	if exists == 0 {
		fmt.Printf("Error: user '%s' not found\n", username)
		os.Exit(1)
	}
	hash, _ := auth.HashPassword(newPassword)
	database.DB.Exec("UPDATE users SET password_hash = ? WHERE username = ?", hash, username)
	fmt.Printf("Password reset for user '%s'\n", username)
}
