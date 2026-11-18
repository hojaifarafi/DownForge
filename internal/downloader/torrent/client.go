package torrent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"downforge/internal/config"
)

// Engine manages the shared torrent client
type Engine struct {
	client *torrent.Client
	cfg    *config.Config
	mu     sync.RWMutex
	tasks  map[string]*Task
}

// FileInfo provides info about an individual file in a multi-file torrent
type FileInfo struct {
	Path   string `json:"path"`
	Length int64  `json:"length"`
}

// Task represents an active torrent download
type Task struct {
	ID              string
	Source          string // Magnet or torrent file path
	SaveDir         string
	Torrent         *torrent.Torrent
	Name            string
	TotalBytes      int64
	DownloadedBytes atomic.Int64
	DownloadSpeed   atomic.Int64
	UploadSpeed     atomic.Int64
	Peers           atomic.Int32
	isPaused        atomic.Bool
	isDone          atomic.Bool
	cancel          context.CancelFunc
	mu              sync.RWMutex
	onInfoGot       func(name string, totalBytes int64, files []FileInfo)
	onProgress      func(downloaded, total, dSpeed, uSpeed int64, peers int)
	onComplete      func(totalBytes int64)
	onError         func(err error)
}

// NewEngine creates and configures the torrent engine
func NewEngine(cfg *config.Config) (*Engine, error) {
	clientConfig := torrent.NewDefaultClientConfig()
	clientConfig.DataDir = cfg.DownloadDir
	clientConfig.NoDHT = !cfg.EnableDHT
	clientConfig.DisablePEX = !cfg.EnablePEX
	if cfg.TorrentPort > 0 {
		clientConfig.ListenPort = cfg.TorrentPort
	}

	cl, err := torrent.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create torrent client: %w", err)
	}

	return &Engine{
		client: cl,
		cfg:    cfg,
		tasks:  make(map[string]*Task),
	}, nil
}

// Close gracefully closes the engine
func (e *Engine) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.client != nil {
		e.client.Close()
	}
}

// AddMagnet creates and starts a torrent from a magnet URI
func (e *Engine) AddMagnet(id string, magnetURI string, saveDir string,
	onInfoGot func(name string, totalBytes int64, files []FileInfo),
	onProgress func(downloaded, total, dSpeed, uSpeed int64, peers int),
	onComplete func(totalBytes int64),
	onError func(err error),
) (*Task, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	t, err := e.client.AddMagnet(magnetURI)
	if err != nil {
		return nil, fmt.Errorf("failed to add magnet: %w", err)
	}

	task := &Task{
		ID:         id,
		Source:     magnetURI,
		SaveDir:    saveDir,
		Torrent:    t,
		onInfoGot:  onInfoGot,
		onProgress: onProgress,
		onComplete: onComplete,
		onError:    onError,
	}

	e.tasks[id] = task
	task.start()
	return task, nil
}

// AddTorrentFile creates and starts a torrent from .torrent file bytes
func (e *Engine) AddTorrentFile(id string, fileBytes []byte, saveDir string,
	onInfoGot func(name string, totalBytes int64, files []FileInfo),
	onProgress func(downloaded, total, dSpeed, uSpeed int64, peers int),
	onComplete func(totalBytes int64),
	onError func(err error),
) (*Task, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	mi, err := metainfo.Load(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("invalid torrent file: %w", err)
	}

	t, err := e.client.AddTorrent(mi)
	if err != nil {
		return nil, fmt.Errorf("failed to add torrent: %w", err)
	}

	task := &Task{
		ID:         id,
		Source:     "torrent_file",
		SaveDir:    saveDir,
		Torrent:    t,
		onInfoGot:  onInfoGot,
		onProgress: onProgress,
		onComplete: onComplete,
		onError:    onError,
	}

	e.tasks[id] = task
	task.start()
	return task, nil
}

// GetTask returns a task by ID
func (e *Engine) GetTask(id string) *Task {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.tasks[id]
}

// RemoveTask drops a torrent from the engine
func (e *Engine) RemoveTask(id string) {
	e.mu.Lock()
	task, ok := e.tasks[id]
	if ok {
		delete(e.tasks, id)
	}
	e.mu.Unlock()

	if ok && task != nil {
		task.stop()
		if task.Torrent != nil {
			task.Torrent.Drop()
		}
	}
}

func (t *Task) start() {
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	go t.monitor(ctx)
}

func (t *Task) stop() {
	if t.cancel != nil {
		t.cancel()
	}
}

// Pause pauses the torrent downloading
func (t *Task) Pause() {
	t.isPaused.Store(true)
	if t.Torrent != nil {
		t.Torrent.DisallowDataDownload()
	}
}

// Resume resumes the torrent downloading
func (t *Task) Resume() {
	t.isPaused.Store(false)
	if t.Torrent != nil {
		t.Torrent.AllowDataDownload()
		t.Torrent.DownloadAll()
	}
}

// GetStats returns current progress snapshot
func (t *Task) GetStats() (downloaded, total, dSpeed, uSpeed int64, peers int) {
	return t.DownloadedBytes.Load(), t.TotalBytes, t.DownloadSpeed.Load(), t.UploadSpeed.Load(), int(t.Peers.Load())
}

func (t *Task) monitor(ctx context.Context) {
	// Wait for metadata
	select {
	case <-ctx.Done():
		return
	case <-t.Torrent.GotInfo():
	}

	t.Name = t.Torrent.Name()
	t.TotalBytes = t.Torrent.Length()

	var fileList []FileInfo
	for _, f := range t.Torrent.Files() {
		fileList = append(fileList, FileInfo{
			Path:   f.DisplayPath(),
			Length: f.Length(),
		})
	}

	if t.onInfoGot != nil {
		t.onInfoGot(t.Name, t.TotalBytes, fileList)
	}

	// Begin downloading pieces if not paused
	if !t.isPaused.Load() {
		t.Torrent.DownloadAll()
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastBytes int64 = t.Torrent.BytesCompleted()
	var lastUploadBytes int64 = 0
	lastTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			current := t.Torrent.BytesCompleted()
			t.DownloadedBytes.Store(current)

			stats := t.Torrent.Stats()
			t.Peers.Store(int32(len(t.Torrent.PeerConns())))

			elapsed := now.Sub(lastTime).Seconds()
			if elapsed > 0 {
				dSpeed := int64(float64(current-lastBytes) / elapsed)
				if dSpeed < 0 {
					dSpeed = 0
				}
				t.DownloadSpeed.Store(dSpeed)

				currentUpload := stats.BytesWrittenData.Int64()
				uSpeed := int64(float64(currentUpload-lastUploadBytes) / elapsed)
				if uSpeed < 0 {
					uSpeed = 0
				}
				t.UploadSpeed.Store(uSpeed)

				lastBytes = current
				lastUploadBytes = currentUpload
				lastTime = now

				if t.onProgress != nil {
					t.onProgress(current, t.TotalBytes, dSpeed, uSpeed, int(t.Peers.Load()))
				}
			}

			// Check completion
			if t.TotalBytes > 0 && current >= t.TotalBytes && !t.isDone.Load() {
				t.isDone.Store(true)
				if t.onComplete != nil {
					t.onComplete(t.TotalBytes)
				}
			}
		}
	}
}

// SerializeFiles converts file list to JSON string for metadata storage
func SerializeFiles(files []FileInfo) string {
	data, err := json.Marshal(files)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// TorrentSavePath determines where files are saved
func TorrentSavePath(baseDir, name string) string {
	return filepath.Join(baseDir, name)
}
