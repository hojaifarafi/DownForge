package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRepositoryCRUD(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "downforge_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.Close()

	repo := NewRepository(database)

	// Create test download item
	item := &DownloadItem{
		ID:              "test-id-1",
		Name:            "ubuntu-24.04.iso",
		Type:            TypeHTTP,
		Source:          "https://example.com/ubuntu-24.04.iso",
		SavePath:        filepath.Join(tmpDir, "ubuntu-24.04.iso"),
		TotalBytes:      2000000000,
		DownloadedBytes: 500000000,
		Status:          StatusDownloading,
	}

	if err := repo.CreateDownload(item); err != nil {
		t.Fatalf("CreateDownload failed: %v", err)
	}

	// Read
	got, err := repo.GetDownload("test-id-1")
	if err != nil {
		t.Fatalf("GetDownload failed: %v", err)
	}
	if got == nil || got.Name != item.Name {
		t.Fatalf("expected name %s, got %+v", item.Name, got)
	}

	// Update Progress
	if err := repo.UpdateProgress("test-id-1", 1000000000, 2000000000, StatusDownloading); err != nil {
		t.Fatalf("UpdateProgress failed: %v", err)
	}

	// Verify updated progress
	got, _ = repo.GetDownload("test-id-1")
	if got.DownloadedBytes != 1000000000 {
		t.Fatalf("expected downloaded_bytes 1000000000, got %d", got.DownloadedBytes)
	}

	// Update Status to completed
	if err := repo.UpdateStatus("test-id-1", StatusCompleted, ""); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	got, _ = repo.GetDownload("test-id-1")
	if got.Status != StatusCompleted || got.CompletedAt == nil {
		t.Fatalf("expected status completed with timestamp, got %+v", got)
	}

	// List
	list, err := repo.ListDownloads()
	if err != nil {
		t.Fatalf("ListDownloads failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list))
	}

	// Delete
	if err := repo.DeleteDownload("test-id-1"); err != nil {
		t.Fatalf("DeleteDownload failed: %v", err)
	}

	got, _ = repo.GetDownload("test-id-1")
	if got != nil {
		t.Fatalf("expected nil after delete, got %+v", got)
	}
}
