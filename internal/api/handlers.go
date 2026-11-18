package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-chi/chi/v5"

	"downforge/internal/config"
	"downforge/internal/downloader"
)

// Handler holds dependencies for REST API endpoints
type Handler struct {
	manager *downloader.Manager
	cfg     *config.Config
	wsHub   *WSHub
}

// NewHandler creates a new REST handler
func NewHandler(manager *downloader.Manager, cfg *config.Config, wsHub *WSHub) *Handler {
	return &Handler{
		manager: manager,
		cfg:     cfg,
		wsHub:   wsHub,
	}
}

// JSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// Error writes an error JSON response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// ListDownloads handles GET /api/downloads
func (h *Handler) ListDownloads(w http.ResponseWriter, r *http.Request) {
	snapshots := h.manager.ListSnapshots()
	writeJSON(w, http.StatusOK, snapshots)
}

// GetDownload handles GET /api/downloads/{id}
func (h *Handler) GetDownload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	snap, err := h.manager.GetSnapshot(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, snap)
}

// AddDownloadRequest payload for adding a new download
type AddDownloadRequest struct {
	Type     string `json:"type"` // "http" or "torrent"
	URL      string `json:"url"`  // HTTP URL or Magnet URI
	Name     string `json:"name,omitempty"`
	SaveDir  string `json:"save_dir,omitempty"`
}

// AddDownload handles POST /api/downloads
func (h *Handler) AddDownload(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	// Support multipart/form-data for .torrent file uploads
	if strings.HasPrefix(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(32 << 20) // 32MB max
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to parse multipart form")
			return
		}

		file, header, err := r.FormFile("torrent")
		if err != nil {
			writeError(w, http.StatusBadRequest, "missing 'torrent' file field")
			return
		}
		defer file.Close()

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read torrent file")
			return
		}

		saveDir := r.FormValue("save_dir")
		if saveDir == "" {
			saveDir = h.cfg.DownloadDir
		}

		snap, err := h.manager.AddTorrentFile(fileBytes, header.Filename, saveDir)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, snap)
		return
	}

	// JSON payload
	var req AddDownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "URL cannot be empty")
		return
	}

	saveDir := req.SaveDir
	if saveDir == "" {
		saveDir = h.cfg.DownloadDir
	}

	// Auto-detect type if not provided
	if req.Type == "" {
		if strings.HasPrefix(strings.ToLower(req.URL), "magnet:") {
			req.Type = "torrent"
		} else {
			req.Type = "http"
		}
	}

	var snap *downloader.Snapshot
	var err error

	if req.Type == "torrent" {
		snap, err = h.manager.AddMagnet(req.URL, saveDir)
	} else {
		snap, err = h.manager.AddHTTP(req.URL, req.Name, saveDir)
	}

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, snap)
}

// PauseDownload handles POST /api/downloads/{id}/pause
func (h *Handler) PauseDownload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.manager.Pause(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "paused", "id": id})
}

// ResumeDownload handles POST /api/downloads/{id}/resume
func (h *Handler) ResumeDownload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.manager.Resume(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "resumed", "id": id})
}

// DeleteDownload handles DELETE /api/downloads/{id}
func (h *Handler) DeleteDownload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	deleteFiles := r.URL.Query().Get("deleteFiles") == "true"

	if err := h.manager.Delete(id, deleteFiles); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
}

// GetStats handles GET /api/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	snapshots := h.manager.ListSnapshots()
	stats := h.manager.CalculateStats(snapshots)
	writeJSON(w, http.StatusOK, stats)
}

// GetSettings handles GET /api/settings
func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.cfg)
}

// UpdateSettings handles PUT /api/settings
func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var newCfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	if newCfg.DownloadDir != "" {
		h.cfg.DownloadDir = newCfg.DownloadDir
	}
	if newCfg.MaxConcurrent > 0 {
		h.cfg.MaxConcurrent = newCfg.MaxConcurrent
	}
	if newCfg.Port > 0 {
		h.cfg.Port = newCfg.Port
	}
	h.cfg.EnableDHT = newCfg.EnableDHT
	h.cfg.EnablePEX = newCfg.EnablePEX
	h.cfg.UploadRateLimit = newCfg.UploadRateLimit
	h.cfg.DownloadRateLimit = newCfg.DownloadRateLimit

	if err := config.SaveConfig(h.cfg); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save config")
		return
	}

	writeJSON(w, http.StatusOK, h.cfg)
}

// OpenFolder handles POST /api/open-folder
func (h *Handler) OpenFolder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	target := req.Path
	if target == "" {
		target = h.cfg.DownloadDir
	}

	dir := target
	// If path points to a file, open its enclosing directory
	if filepath.Ext(target) != "" {
		dir = filepath.Dir(target)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", dir)
	case "darwin":
		cmd = exec.Command("open", dir)
	default:
		cmd = exec.Command("xdg-open", dir)
	}

	_ = cmd.Start()
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "path": dir})
}
