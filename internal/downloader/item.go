package downloader

import (
	"time"

	"downforge/internal/db"
)

// Snapshot represents real-time state of a download item for UI & WebSocket clients
type Snapshot struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Type            db.DownloadType `json:"type"`
	Source          string          `json:"source"`
	SavePath        string          `json:"save_path"`
	TotalBytes      int64           `json:"total_bytes"`
	DownloadedBytes int64           `json:"downloaded_bytes"`
	Progress        float64         `json:"progress"`        // 0.00 to 100.00 %
	DownloadSpeed   int64           `json:"download_speed"`  // bytes per second
	UploadSpeed     int64           `json:"upload_speed"`    // bytes per second (torrents)
	ETA             int64           `json:"eta"`             // remaining seconds (-1 if unknown)
	Status          db.DownloadStatus `json:"status"`
	ErrorMessage    string          `json:"error_message,omitempty"`
	Peers           int             `json:"peers,omitempty"` // For torrents
	Seeds           int             `json:"seeds,omitempty"` // For torrents
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`
}

// CalculateProgress calculates percentage from downloaded and total bytes
func CalculateProgress(downloaded, total int64) float64 {
	if total <= 0 {
		return 0
	}
	progress := (float64(downloaded) / float64(total)) * 100.0
	if progress > 100.0 {
		return 100.0
	}
	return progress
}

// CalculateETA calculates estimated seconds remaining
func CalculateETA(downloaded, total, speed int64) int64 {
	if speed <= 0 || total <= 0 || downloaded >= total {
		return -1
	}
	remaining := total - downloaded
	return remaining / speed
}
