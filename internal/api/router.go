package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// SetupRouter constructs and configures the Chi HTTP router
func SetupRouter(h *Handler, wsHub *WSHub, staticDir string) *chi.Mux {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Cross-Origin Resource Sharing (CORS) config for React dev & localhost access
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// WebSocket endpoint
	r.Get("/ws", wsHub.ServeWS)

	// REST API routes
	r.Route("/api", func(api chi.Router) {
		api.Get("/downloads", h.ListDownloads)
		api.Post("/downloads", h.AddDownload)
		api.Get("/downloads/{id}", h.GetDownload)
		api.Post("/downloads/{id}/pause", h.PauseDownload)
		api.Post("/downloads/{id}/resume", h.ResumeDownload)
		api.Delete("/downloads/{id}", h.DeleteDownload)

		api.Get("/stats", h.GetStats)
		api.Get("/settings", h.GetSettings)
		api.Put("/settings", h.UpdateSettings)
		api.Post("/open-folder", h.OpenFolder)
	})

	// Static files / Frontend UI serving with SPA fallback
	if staticDir != "" {
		serveSPA(r, "/", staticDir)
	}

	return r
}

// serveSPA serves static frontend files and falls back to index.html for client-side routing
func serveSPA(r *chi.Mux, pathPrefix string, root string) {
	fs := http.FileServer(http.Dir(root))

	r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
		// If requesting an API or WS path, let Chi handle 404
		if strings.HasPrefix(req.URL.Path, "/api") || req.URL.Path == "/ws" {
			http.NotFound(w, req)
			return
		}

		filePath := filepath.Join(root, filepath.Clean(req.URL.Path))
		info, err := os.Stat(filePath)

		if err == nil && !info.IsDir() {
			fs.ServeHTTP(w, req)
			return
		}

		// Fallback to index.html
		indexPath := filepath.Join(root, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, req, indexPath)
			return
		}

		// If frontend hasn't been built yet, show a friendly status message
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head><title>DownForge Server</title><style>body{font-family:sans-serif;padding:40px;background:#0f172a;color:#f8fafc;}</style></head>
			<body>
				<h1>⚡ DownForge Server is Running</h1>
				<p>The backend REST API and WebSocket are active.</p>
				<p>API endpoint: <code>/api/downloads</code> | WebSocket: <code>/ws</code></p>
				<p>Build the React UI with <code>npm run build</code> in <code>web/</code> or run Vite dev server at <code>http://localhost:5173</code>.</p>
			</body>
			</html>
		`))
	})
}
