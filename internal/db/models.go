package db

import (
	"time"
)

// DownloadType indicates whether a download is HTTP or Torrent
type DownloadType string

const (
	TypeHTTP    DownloadType = "http"
	TypeTorrent DownloadType = "torrent"
)

// DownloadStatus represents the current state of a download
type DownloadStatus string

const (
	StatusQueued      DownloadStatus = "queued"
	StatusDownloading DownloadStatus = "downloading"
	StatusPaused      DownloadStatus = "paused"
	StatusCompleted   DownloadStatus = "completed"
	StatusError       DownloadStatus = "error"
)

// DownloadItem represents a stored download in SQLite
type DownloadItem struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Type            DownloadType   `json:"type"`
	Source          string         `json:"source"`           // URL or Magnet URI
	SavePath        string         `json:"save_path"`        // Target file or directory path
	TotalBytes      int64          `json:"total_bytes"`
	DownloadedBytes int64          `json:"downloaded_bytes"`
	Status          DownloadStatus `json:"status"`
	ErrorMessage    string         `json:"error_message,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	CompletedAt     *time.Time     `json:"completed_at,omitempty"`
	Metadata        string         `json:"metadata,omitempty"` // Extra JSON metadata
}

// GlobalStats summarizes download statistics
type GlobalStats struct {
	TotalDownloads     int   `json:"total_downloads"`
	ActiveDownloads    int   `json:"active_downloads"`
	CompletedDownloads int   `json:"completed_downloads"`
	PausedDownloads    int   `json:"paused_downloads"`
	FailedDownloads    int   `json:"failed_downloads"`
	TotalDownloadSpeed int64 `json:"total_download_speed"` // bytes/sec
	TotalUploadSpeed   int64 `json:"total_upload_speed"`   // bytes/sec
}
