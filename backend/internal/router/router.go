package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/yunuskargi/confbox/internal/auth"
	"github.com/yunuskargi/confbox/internal/config"
	"github.com/yunuskargi/confbox/internal/handler"
	mw "github.com/yunuskargi/confbox/internal/middleware"
)

func New() http.Handler {
	r := chi.NewRouter()

	r.Use(mw.SecurityHeaders)

	origins := config.CORSOrigins
	if len(origins) == 0 {
		origins = []string{"http://localhost:5173", "http://localhost:6161"}
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}))

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/api/settings/branding", handler.GetBranding)

	r.Route("/api/auth", func(r chi.Router) {
		r.With(mw.RateLimit).Post("/login", handler.Login)

		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware)
			r.Get("/me", handler.Me)
			r.Post("/logout", handler.Logout)
			r.Post("/change-password", handler.ChangePassword)
			r.Post("/2fa/setup", handler.Setup2FA)
			r.Post("/2fa/verify", handler.Verify2FA)
			r.Post("/2fa/disable", handler.Disable2FA)
		})
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(auth.Middleware)

		r.Route("/devices", func(r chi.Router) {
			r.Get("/", handler.ListDevices)
			r.Post("/", handler.CreateDevice)
			r.Get("/bulk/template", handler.BulkTemplate)
			r.Post("/bulk/preview", handler.BulkPreview)
			r.Post("/bulk/import", handler.BulkImport)
			r.Get("/{id}", handler.GetDevice)
			r.Put("/{id}", handler.UpdateDevice)
			r.Delete("/{id}", handler.DeleteDevice)
			r.Post("/{id}/backup", handler.TriggerBackup)
			r.Post("/{id}/test", handler.TestConnection)
			r.Put("/{id}/schedule", handler.SetSchedule)
			r.Delete("/{id}/schedule", handler.RemoveSchedule)
		})

		r.Route("/backups", func(r chi.Router) {
			r.Get("/", handler.ListBackups)
			r.Post("/authorize-download", handler.AuthorizeDownload)
			r.Get("/{id}/download", handler.DownloadBackup)
			r.Get("/{id}/content", handler.GetBackupContent)
			r.Get("/diff/{idA}/{idB}", handler.DiffBackups)
			r.Delete("/{id}", handler.DeleteBackup)
		})

		r.Route("/locations", func(r chi.Router) {
			r.Get("/", handler.ListLocations)
			r.Post("/", handler.CreateLocation)
			r.Put("/{id}", handler.UpdateLocation)
			r.Delete("/{id}", handler.DeleteLocation)
		})

		r.Route("/users", func(r chi.Router) {
			r.Use(auth.AdminOnly)
			r.Get("/", handler.ListUsers)
			r.Post("/", handler.CreateUser)
			r.Put("/{id}", handler.UpdateUser)
			r.Delete("/{id}", handler.DeleteUser)
		})

		r.Route("/settings", func(r chi.Router) {
			r.Get("/", handler.GetSettings)
			r.Put("/", handler.UpdateSettings)
			r.Get("/smtp", handler.GetSMTP)
			r.Put("/smtp", handler.UpdateSMTP)
			r.Post("/smtp/test", handler.TestSMTP)
			r.Get("/notify", handler.GetNotify)
			r.Put("/notify", handler.UpdateNotify)
		})

		r.Route("/dashboard", func(r chi.Router) {
			r.Get("/stats", handler.GetDashboardStats)
			r.Get("/trend", handler.GetBackupTrend)
			r.Get("/size-trend", handler.GetSizeTrend)
		})

		r.Get("/audit", handler.ListAuditLogs)
	})

	return r
}
