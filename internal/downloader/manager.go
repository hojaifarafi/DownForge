package downloader

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"downforge/internal/config"
	"downforge/internal/db"
	downHttp "downforge/internal/downloader/http"
	downTorrent "downforge/internal/downloader/torrent"
)

// BroadcastFunc is invoked when an event or progress update should be sent over WebSocket
type BroadcastFunc func(event string, payload any)

// Manager coordinates HTTP and Torrent downloads with the database
type Manager struct {
	cfg           *config.Config
	repo          *db.Repository
	torrentEngine *downTorrent.Engine
	httpTasks     map[string]*downHttp.Task
	torrentTasks  map[string]*downTorrent.Task
	mu            sync.RWMutex
	broadcasters  []BroadcastFunc
	bMu           sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewManager initializes the download manager
func NewManager(cfg *config.Config, repo *db.Repository) (*Manager, error) {
	tEngine, err := downTorrent.NewEngine(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize torrent engine: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		cfg:           cfg,
		repo:          repo,
		torrentEngine: tEngine,
		httpTasks:     make(map[string]*downHttp.Task),
		torrentTasks:  make(map[string]*downTorrent.Task),
		ctx:           ctx,
		cancel:        cancel,
	}

	return m, nil
}

// RegisterBroadcaster adds a callback for realtime WebSocket broadcasts
func (m *Manager) RegisterBroadcaster(fn BroadcastFunc) {
	m.bMu.Lock()
	defer m.bMu.Unlock()
	m.broadcasters = append(m.broadcasters, fn)
}

func (m *Manager) broadcast(event string, payload any) {
	m.bMu.RLock()
	defer m.bMu.RUnlock()
	for _, fn := range m.broadcasters {
		fn(event, payload)
	}
}

// Start begins the manager background loops (periodic ticker to persist and broadcast progress)
func (m *Manager) Start() {
	// Restore unfinished downloads: mark previously "downloading" items as "paused"
	items, err := m.repo.ListDownloads()
	if err == nil {
		for _, item := range items {
			if item.Status == db.StatusDownloading {
				_ = m.repo.UpdateStatus(item.ID, db.StatusPaused, "")
			}
		}
	}

	go m.tickerLoop()
}

// Stop gracefully stops the manager and engines
func (m *Manager) Stop() {
	m.cancel()
	if m.torrentEngine != nil {
		m.torrentEngine.Close()
	}
}

func (m *Manager) tickerLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			snapshots := m.ListSnapshots()
			stats := m.CalculateStats(snapshots)

			// Broadcast realtime snapshot & global stats over WebSocket
			m.broadcast("progress", snapshots)
			m.broadcast("stats", stats)

			// Update persistent DB values for actively downloading items
			for _, s := range snapshots {
				if s.Status == db.StatusDownloading {
					_ = m.repo.UpdateProgress(s.ID, s.DownloadedBytes, s.TotalBytes, s.Status)
				}
			}
		}
	}
}

// AddHTTP initiates an HTTP or HTTPS download
func (m *Manager) AddHTTP(url string, customName string, saveDir string) (*Snapshot, error) {
	if url == "" {
		return nil, errors.New("url cannot be empty")
	}

	if saveDir == "" {
		saveDir = m.cfg.DownloadDir
	}
	_ = os.MkdirAll(saveDir, 0755)

	name := customName
	if name == "" {
		name = extractNameFromURL(url)
	}

	id := uuid.New().String()
	targetPath := filepath.Join(saveDir, name)

	item := &db.DownloadItem{
		ID:              id,
		Name:            name,
		Type:            db.TypeHTTP,
		Source:          url,
		SavePath:        targetPath,
		TotalBytes:      0,
		DownloadedBytes: 0,
		Status:          db.StatusDownloading,
	}

	if err := m.repo.CreateDownload(item); err != nil {
		return nil, fmt.Errorf("failed to save download to database: %w", err)
	}

	task := downHttp.NewTask(downHttp.TaskOptions{
		ID:         id,
		URL:        url,
		SaveDir:    saveDir,
		CustomName: customName,
		OnProgress: func(downloaded, total, speed int64) {
			// Instant progress callback if needed
		},
		OnComplete: func(finalPath string, totalBytes int64) {
			_ = m.repo.UpdateDetails(id, filepath.Base(finalPath), finalPath, totalBytes, "")
			_ = m.repo.UpdateStatus(id, db.StatusCompleted, "")
			m.mu.Lock()
			delete(m.httpTasks, id)
			m.mu.Unlock()
			m.broadcast("completed", id)
		},
		OnError: func(err error) {
			_ = m.repo.UpdateStatus(id, db.StatusError, err.Error())
			m.mu.Lock()
			delete(m.httpTasks, id)
			m.mu.Unlock()
			m.broadcast("error", map[string]string{"id": id, "error": err.Error()})
		},
	})

	m.mu.Lock()
	m.httpTasks[id] = task
	m.mu.Unlock()

	if err := task.Start(m.ctx); err != nil {
		_ = m.repo.UpdateStatus(id, db.StatusError, err.Error())
		return nil, err
	}

	m.broadcast("added", id)
	return m.GetSnapshot(id)
}

// AddMagnet initiates a Torrent download via magnet URI
func (m *Manager) AddMagnet(magnetURI string, saveDir string) (*Snapshot, error) {
	if magnetURI == "" {
		return nil, errors.New("magnet URI cannot be empty")
	}

	if saveDir == "" {
		saveDir = m.cfg.DownloadDir
	}
	_ = os.MkdirAll(saveDir, 0755)

	id := uuid.New().String()
	name := "Retrieving torrent metadata..."

	item := &db.DownloadItem{
		ID:              id,
		Name:            name,
		Type:            db.TypeTorrent,
		Source:          magnetURI,
		SavePath:        filepath.Join(saveDir, id),
		TotalBytes:      0,
		DownloadedBytes: 0,
		Status:          db.StatusDownloading,
	}

	if err := m.repo.CreateDownload(item); err != nil {
		return nil, fmt.Errorf("failed to save torrent to database: %w", err)
	}

	task, err := m.torrentEngine.AddMagnet(
		id,
		magnetURI,
		saveDir,
		func(tName string, totalBytes int64, files []downTorrent.FileInfo) {
			meta := downTorrent.SerializeFiles(files)
			finalPath := filepath.Join(saveDir, tName)
			_ = m.repo.UpdateDetails(id, tName, finalPath, totalBytes, meta)
			m.broadcast("metadata_resolved", id)
		},
		func(downloaded, total, dSpeed, uSpeed int64, peers int) {
			// Real-time progress updates handled by ticker
		},
		func(totalBytes int64) {
			_ = m.repo.UpdateStatus(id, db.StatusCompleted, "")
			m.broadcast("completed", id)
		},
		func(err error) {
			_ = m.repo.UpdateStatus(id, db.StatusError, err.Error())
			m.broadcast("error", map[string]string{"id": id, "error": err.Error()})
		},
	)
	if err != nil {
		_ = m.repo.UpdateStatus(id, db.StatusError, err.Error())
		return nil, err
	}

	m.mu.Lock()
	m.torrentTasks[id] = task
	m.mu.Unlock()

	m.broadcast("added", id)
	return m.GetSnapshot(id)
}

// AddTorrentFile initiates a Torrent download from .torrent file contents
func (m *Manager) AddTorrentFile(fileBytes []byte, filename string, saveDir string) (*Snapshot, error) {
	if len(fileBytes) == 0 {
		return nil, errors.New("torrent file data is empty")
	}

	if saveDir == "" {
		saveDir = m.cfg.DownloadDir
	}
	_ = os.MkdirAll(saveDir, 0755)

	id := uuid.New().String()
	name := strings.TrimSuffix(filename, ".torrent")
	if name == "" {
		name = "Torrent Download"
	}

	item := &db.DownloadItem{
		ID:              id,
		Name:            name,
		Type:            db.TypeTorrent,
		Source:          filename,
		SavePath:        filepath.Join(saveDir, name),
		TotalBytes:      0,
		DownloadedBytes: 0,
		Status:          db.StatusDownloading,
	}

	if err := m.repo.CreateDownload(item); err != nil {
		return nil, fmt.Errorf("failed to save torrent to database: %w", err)
	}

	task, err := m.torrentEngine.AddTorrentFile(
		id,
		fileBytes,
		saveDir,
		func(tName string, totalBytes int64, files []downTorrent.FileInfo) {
			meta := downTorrent.SerializeFiles(files)
			finalPath := filepath.Join(saveDir, tName)
			_ = m.repo.UpdateDetails(id, tName, finalPath, totalBytes, meta)
			m.broadcast("metadata_resolved", id)
		},
		func(downloaded, total, dSpeed, uSpeed int64, peers int) {},
		func(totalBytes int64) {
			_ = m.repo.UpdateStatus(id, db.StatusCompleted, "")
			m.broadcast("completed", id)
		},
		func(err error) {
			_ = m.repo.UpdateStatus(id, db.StatusError, err.Error())
			m.broadcast("error", map[string]string{"id": id, "error": err.Error()})
		},
	)
	if err != nil {
		_ = m.repo.UpdateStatus(id, db.StatusError, err.Error())
		return nil, err
	}

	m.mu.Lock()
	m.torrentTasks[id] = task
	m.mu.Unlock()

	m.broadcast("added", id)
	return m.GetSnapshot(id)
}

// Pause pauses an active download
func (m *Manager) Pause(id string) error {
	m.mu.Lock()
	httpTask, hasHttp := m.httpTasks[id]
	torrentTask, hasTorrent := m.torrentTasks[id]
	m.mu.Unlock()

	if hasHttp && httpTask != nil {
		httpTask.Pause()
		m.mu.Lock()
		delete(m.httpTasks, id)
		m.mu.Unlock()
	}

	if hasTorrent && torrentTask != nil {
		torrentTask.Pause()
	}

	err := m.repo.UpdateStatus(id, db.StatusPaused, "")
	m.broadcast("paused", id)
	return err
}

// Resume resumes a paused download
func (m *Manager) Resume(id string) error {
	item, err := m.repo.GetDownload(id)
	if err != nil || item == nil {
		return fmt.Errorf("download not found: %s", id)
	}

	if item.Status == db.StatusDownloading {
		return nil // already downloading
	}

	if item.Type == db.TypeHTTP {
		task := downHttp.NewTask(downHttp.TaskOptions{
			ID:         id,
			URL:        item.Source,
			SaveDir:    filepath.Dir(item.SavePath),
			CustomName: filepath.Base(item.SavePath),
			OnComplete: func(finalPath string, totalBytes int64) {
				_ = m.repo.UpdateDetails(id, filepath.Base(finalPath), finalPath, totalBytes, "")
				_ = m.repo.UpdateStatus(id, db.StatusCompleted, "")
				m.mu.Lock()
				delete(m.httpTasks, id)
				m.mu.Unlock()
				m.broadcast("completed", id)
			},
			OnError: func(err error) {
				_ = m.repo.UpdateStatus(id, db.StatusError, err.Error())
				m.mu.Lock()
				delete(m.httpTasks, id)
				m.mu.Unlock()
				m.broadcast("error", map[string]string{"id": id, "error": err.Error()})
			},
		})

		m.mu.Lock()
		m.httpTasks[id] = task
		m.mu.Unlock()

		if err := task.Start(m.ctx); err != nil {
			_ = m.repo.UpdateStatus(id, db.StatusError, err.Error())
			return err
		}
	} else if item.Type == db.TypeTorrent {
		m.mu.RLock()
		torrentTask := m.torrentTasks[id]
		m.mu.RUnlock()

		if torrentTask != nil {
			torrentTask.Resume()
		} else {
			// Re-add to engine if not active
			saveDir := filepath.Dir(item.SavePath)
			newTask, err := m.torrentEngine.AddMagnet(
				id,
				item.Source,
				saveDir,
				func(tName string, totalBytes int64, files []downTorrent.FileInfo) {
					meta := downTorrent.SerializeFiles(files)
					_ = m.repo.UpdateDetails(id, tName, filepath.Join(saveDir, tName), totalBytes, meta)
				},
				func(downloaded, total, dSpeed, uSpeed int64, peers int) {},
				func(totalBytes int64) {
					_ = m.repo.UpdateStatus(id, db.StatusCompleted, "")
					m.broadcast("completed", id)
				},
				func(err error) {
					_ = m.repo.UpdateStatus(id, db.StatusError, err.Error())
					m.broadcast("error", map[string]string{"id": id, "error": err.Error()})
				},
			)
			if err != nil {
				return err
			}
			m.mu.Lock()
			m.torrentTasks[id] = newTask
			m.mu.Unlock()
		}
	}

	err = m.repo.UpdateStatus(id, db.StatusDownloading, "")
	m.broadcast("resumed", id)
	return err
}

// Delete removes a download and optionally deletes downloaded files
func (m *Manager) Delete(id string, deleteFiles bool) error {
	m.mu.Lock()
	httpTask, hasHttp := m.httpTasks[id]
	if hasHttp && httpTask != nil {
		httpTask.Pause()
		delete(m.httpTasks, id)
	}

	_, hasTorrent := m.torrentTasks[id]
	if hasTorrent {
		delete(m.torrentTasks, id)
	}
	m.mu.Unlock()

	if hasTorrent {
		m.torrentEngine.RemoveTask(id)
	}

	item, _ := m.repo.GetDownload(id)

	if deleteFiles && item != nil && item.SavePath != "" {
		_ = os.Remove(item.SavePath)
		_ = os.Remove(item.SavePath + ".part")
		_ = os.RemoveAll(item.SavePath) // If directory (multi-file torrent)
	}

	err := m.repo.DeleteDownload(id)
	m.broadcast("deleted", id)
	return err
}

// GetSnapshot retrieves the snapshot for a given download
func (m *Manager) GetSnapshot(id string) (*Snapshot, error) {
	item, err := m.repo.GetDownload(id)
	if err != nil || item == nil {
		return nil, errors.New("download not found")
	}

	snap := &Snapshot{
		ID:              item.ID,
		Name:            item.Name,
		Type:            item.Type,
		Source:          item.Source,
		SavePath:        item.SavePath,
		TotalBytes:      item.TotalBytes,
		DownloadedBytes: item.DownloadedBytes,
		Status:          item.Status,
		ErrorMessage:    item.ErrorMessage,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
		CompletedAt:     item.CompletedAt,
		ETA:             -1,
	}

	m.mu.RLock()
	httpTask := m.httpTasks[id]
	torrentTask := m.torrentTasks[id]
	m.mu.RUnlock()

	if httpTask != nil {
		dl, total, speed := httpTask.GetStats()
		snap.DownloadedBytes = dl
		if total > 0 {
			snap.TotalBytes = total
		}
		snap.DownloadSpeed = speed
		snap.Progress = CalculateProgress(snap.DownloadedBytes, snap.TotalBytes)
		snap.ETA = CalculateETA(snap.DownloadedBytes, snap.TotalBytes, speed)
	} else if torrentTask != nil {
		dl, total, dSpeed, uSpeed, peers := torrentTask.GetStats()
		snap.DownloadedBytes = dl
		if total > 0 {
			snap.TotalBytes = total
		}
		snap.DownloadSpeed = dSpeed
		snap.UploadSpeed = uSpeed
		snap.Peers = peers
		snap.Progress = CalculateProgress(snap.DownloadedBytes, snap.TotalBytes)
		snap.ETA = CalculateETA(snap.DownloadedBytes, snap.TotalBytes, dSpeed)
	} else {
		snap.Progress = CalculateProgress(snap.DownloadedBytes, snap.TotalBytes)
	}

	return snap, nil
}

// ListSnapshots returns all downloads with real-time stats
func (m *Manager) ListSnapshots() []*Snapshot {
	items, err := m.repo.ListDownloads()
	if err != nil {
		return []*Snapshot{}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []*Snapshot
	for _, item := range items {
		snap := &Snapshot{
			ID:              item.ID,
			Name:            item.Name,
			Type:            item.Type,
			Source:          item.Source,
			SavePath:        item.SavePath,
			TotalBytes:      item.TotalBytes,
			DownloadedBytes: item.DownloadedBytes,
			Status:          item.Status,
			ErrorMessage:    item.ErrorMessage,
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.UpdatedAt,
			CompletedAt:     item.CompletedAt,
			ETA:             -1,
		}

		if httpTask, ok := m.httpTasks[item.ID]; ok && httpTask != nil {
			dl, total, speed := httpTask.GetStats()
			snap.DownloadedBytes = dl
			if total > 0 {
				snap.TotalBytes = total
			}
			snap.DownloadSpeed = speed
			snap.Progress = CalculateProgress(snap.DownloadedBytes, snap.TotalBytes)
			snap.ETA = CalculateETA(snap.DownloadedBytes, snap.TotalBytes, speed)
		} else if torrentTask, ok := m.torrentTasks[item.ID]; ok && torrentTask != nil {
			dl, total, dSpeed, uSpeed, peers := torrentTask.GetStats()
			snap.DownloadedBytes = dl
			if total > 0 {
				snap.TotalBytes = total
			}
			snap.DownloadSpeed = dSpeed
			snap.UploadSpeed = uSpeed
			snap.Peers = peers
			snap.Progress = CalculateProgress(snap.DownloadedBytes, snap.TotalBytes)
			snap.ETA = CalculateETA(snap.DownloadedBytes, snap.TotalBytes, dSpeed)
		} else {
			snap.Progress = CalculateProgress(snap.DownloadedBytes, snap.TotalBytes)
		}

		list = append(list, snap)
	}

	return list
}

// CalculateStats aggregates stats from snapshots
func (m *Manager) CalculateStats(snapshots []*Snapshot) *db.GlobalStats {
	stats := &db.GlobalStats{
		TotalDownloads: len(snapshots),
	}

	for _, s := range snapshots {
		switch s.Status {
		case db.StatusDownloading:
			stats.ActiveDownloads++
			stats.TotalDownloadSpeed += s.DownloadSpeed
			stats.TotalUploadSpeed += s.UploadSpeed
		case db.StatusCompleted:
			stats.CompletedDownloads++
		case db.StatusPaused:
			stats.PausedDownloads++
		case db.StatusError:
			stats.FailedDownloads++
		}
	}

	return stats
}

func extractNameFromURL(rawURL string) string {
	parts := strings.Split(rawURL, "/")
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		if idx := strings.Index(last, "?"); idx != -1 {
			last = last[:idx]
		}
		if last != "" {
			return last
		}
	}
	return "download"
}
