package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"downforge/internal/config"
	"downforge/internal/db"
	"downforge/internal/downloader"
)

func TestAPIEndpoints(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "downforge_api_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.Close()

	repo := db.NewRepository(database)
	cfg := config.DefaultConfig()
	cfg.DBPath = dbPath
	cfg.DownloadDir = tmpDir

	wsHub := NewWSHub()
	go wsHub.Run()

	mgr, err := downloader.NewManager(cfg, repo)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Stop()

	handler := NewHandler(mgr, cfg, wsHub)
	router := SetupRouter(handler, wsHub, "")

	server := httptest.NewServer(router)
	defer server.Close()

	// 1. Test GET /api/downloads
	resp, err := http.Get(server.URL + "/api/downloads")
	if err != nil {
		t.Fatalf("GET /api/downloads failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var items []*downloader.Snapshot
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items initially, got %d", len(items))
	}

	// 2. Test GET /api/stats
	respStats, err := http.Get(server.URL + "/api/stats")
	if err != nil {
		t.Fatalf("GET /api/stats failed: %v", err)
	}
	defer respStats.Body.Close()

	if respStats.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for stats, got %d", respStats.StatusCode)
	}

	var stats db.GlobalStats
	if err := json.NewDecoder(respStats.Body).Decode(&stats); err != nil {
		t.Fatalf("failed to decode stats: %v", err)
	}

	// 3. Test GET /api/settings
	respSettings, err := http.Get(server.URL + "/api/settings")
	if err != nil {
		t.Fatalf("GET /api/settings failed: %v", err)
	}
	defer respSettings.Body.Close()

	if respSettings.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for settings, got %d", respSettings.StatusCode)
	}
}
