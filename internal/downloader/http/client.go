package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Task represents an active HTTP/HTTPS download task
type Task struct {
	ID              string
	URL             string
	SaveDir         string
	CustomName      string
	SavePath        string
	TotalBytes      int64
	DownloadedBytes atomic.Int64
	CurrentSpeed    atomic.Int64 // bytes per second
	CancelFunc      context.CancelFunc
	mu              sync.Mutex
	isPaused        bool
	isDone          bool
	err             error
	onProgress      func(downloaded, total, speed int64)
	onComplete      func(finalPath string, totalBytes int64)
	onError         func(err error)
}

// Config options for the HTTP task
type TaskOptions struct {
	ID         string
	URL        string
	SaveDir    string
	CustomName string
	OnProgress func(downloaded, total, speed int64)
	OnComplete func(finalPath string, totalBytes int64)
	OnError    func(err error)
}

// NewTask creates a new HTTP download task
func NewTask(opts TaskOptions) *Task {
	return &Task{
		ID:         opts.ID,
		URL:        opts.URL,
		SaveDir:    opts.SaveDir,
		CustomName: opts.CustomName,
		onProgress: opts.OnProgress,
		onComplete: opts.OnComplete,
		onError:    opts.OnError,
	}
}

// Start begins or resumes downloading the file
func (t *Task) Start(parentCtx context.Context) error {
	t.mu.Lock()
	if t.isDone {
		t.mu.Unlock()
		return errors.New("download already completed")
	}

	ctx, cancel := context.WithCancel(parentCtx)
	t.CancelFunc = cancel
	t.isPaused = false
	t.mu.Unlock()

	go t.run(ctx)
	return nil
}

// Pause stops the download cleanly
func (t *Task) Pause() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.isPaused = true
	if t.CancelFunc != nil {
		t.CancelFunc()
	}
}

// GetStats returns current progress figures
func (t *Task) GetStats() (downloaded int64, total int64, speed int64) {
	return t.DownloadedBytes.Load(), t.TotalBytes, t.CurrentSpeed.Load()
}

func (t *Task) run(ctx context.Context) {
	// First, send a HEAD or GET request with Range: bytes=0-0 to probe filename and accept-ranges
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, t.URL, nil)
	if err != nil {
		t.handleError(fmt.Errorf("failed to create head request: %w", err))
		return
	}
	req.Header.Set("User-Agent", "DownForge/1.0")

	resp, err := client.Do(req)
	var filename string
	var totalLength int64 = -1
	acceptsRanges := false

	if err == nil {
		defer resp.Body.Close()
		totalLength = resp.ContentLength
		if resp.Header.Get("Accept-Ranges") == "bytes" {
			acceptsRanges = true
		}
		filename = extractFilename(resp.Header.Get("Content-Disposition"), t.URL)
	} else {
		// Fallback: If HEAD fails (some servers deny HEAD), try GET with Range
		filename = extractFilename("", t.URL)
	}

	if t.CustomName != "" {
		filename = t.CustomName
	}
	if filename == "" {
		filename = fmt.Sprintf("download_%s", t.ID[:8])
	}

	// Prepare file paths
	targetPath := filepath.Join(t.SaveDir, filename)
	partPath := targetPath + ".part"
	t.SavePath = targetPath

	// Check if part file already exists
	var existingBytes int64 = 0
	if fi, err := os.Stat(partPath); err == nil {
		existingBytes = fi.Size()
	}

	// Setup actual GET download request
	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, t.URL, nil)
	if err != nil {
		t.handleError(err)
		return
	}
	getReq.Header.Set("User-Agent", "DownForge/1.0")

	// If we have existing bytes, attempt resume
	resuming := false
	if existingBytes > 0 && acceptsRanges {
		getReq.Header.Set("Range", fmt.Sprintf("bytes=%d-", existingBytes))
		resuming = true
	}

	// Use transport with no fixed long timeout for downloading large files
	downloadClient := &http.Client{
		Timeout: 0,
	}

	getResp, err := downloadClient.Do(getReq)
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return // Normal pause / cancel
		}
		t.handleError(fmt.Errorf("failed to start download: %w", err))
		return
	}
	defer getResp.Body.Close()

	if totalLength <= 0 {
		if resuming && getResp.StatusCode == http.StatusPartialContent {
			totalLength = existingBytes + getResp.ContentLength
		} else {
			totalLength = getResp.ContentLength
		}
	}
	t.TotalBytes = totalLength

	var file *os.File
	if resuming && getResp.StatusCode == http.StatusPartialContent {
		// Append to existing part file
		file, err = os.OpenFile(partPath, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			t.handleError(fmt.Errorf("failed to open part file for append: %w", err))
			return
		}
		t.DownloadedBytes.Store(existingBytes)
	} else {
		// Create or truncate part file
		file, err = os.OpenFile(partPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			t.handleError(fmt.Errorf("failed to create part file: %w", err))
			return
		}
		t.DownloadedBytes.Store(0)
	}
	defer file.Close()

	// Speed calculation ticker
	speedTicker := time.NewTicker(500 * time.Millisecond)
	defer speedTicker.Stop()

	var lastBytes int64 = t.DownloadedBytes.Load()
	lastCheckTime := time.Now()

	stopSpeedMonitor := make(chan struct{})
	defer close(stopSpeedMonitor)

	go func() {
		for {
			select {
			case <-stopSpeedMonitor:
				t.CurrentSpeed.Store(0)
				return
			case now := <-speedTicker.C:
				current := t.DownloadedBytes.Load()
				elapsed := now.Sub(lastCheckTime).Seconds()
				if elapsed > 0 {
					speed := int64(float64(current-lastBytes) / elapsed)
					if speed < 0 {
						speed = 0
					}
					t.CurrentSpeed.Store(speed)
					lastBytes = current
					lastCheckTime = now

					if t.onProgress != nil {
						t.onProgress(current, t.TotalBytes, speed)
					}
				}
			}
		}
	}()

	// Buffer for copying
	buf := make([]byte, 64*1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, rErr := getResp.Body.Read(buf)
		if n > 0 {
			if _, wErr := file.Write(buf[:n]); wErr != nil {
				t.handleError(fmt.Errorf("failed to write to file: %w", wErr))
				return
			}
			t.DownloadedBytes.Add(int64(n))
		}

		if rErr != nil {
			if errors.Is(rErr, io.EOF) {
				break
			}
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			t.handleError(fmt.Errorf("download read error: %w", rErr))
			return
		}
	}

	_ = file.Sync()
	_ = file.Close()

	// Download finished successfully: rename .part to final target
	uniquePath := ensureUniqueFilename(targetPath)
	if err := os.Rename(partPath, uniquePath); err != nil {
		t.handleError(fmt.Errorf("failed to rename part file: %w", err))
		return
	}

	t.mu.Lock()
	t.isDone = true
	t.SavePath = uniquePath
	t.mu.Unlock()

	finalSize := t.DownloadedBytes.Load()
	if t.TotalBytes <= 0 {
		t.TotalBytes = finalSize
	}

	if t.onComplete != nil {
		t.onComplete(uniquePath, finalSize)
	}
}

func (t *Task) handleError(err error) {
	t.mu.Lock()
	t.err = err
	t.mu.Unlock()

	if t.onError != nil {
		t.onError(err)
	}
}

// extractFilename extracts the filename from headers or URL
func extractFilename(contentDisposition, rawURL string) string {
	if contentDisposition != "" {
		if _, params, err := mime.ParseMediaType(contentDisposition); err == nil {
			if filename, ok := params["filename"]; ok && filename != "" {
				return filepath.Base(filename)
			}
			if filename, ok := params["filename*"]; ok && filename != "" {
				// Format might be UTF-8''filename
				parts := strings.SplitN(filename, "''", 2)
				if len(parts) == 2 {
					if decoded, err := url.PathUnescape(parts[1]); err == nil {
						return filepath.Base(decoded)
					}
				}
			}
		}
	}

	if u, err := url.Parse(rawURL); err == nil {
		cleanPath := path.Clean(u.Path)
		base := path.Base(cleanPath)
		if base != "" && base != "." && base != "/" {
			return base
		}
	}

	return ""
}

// ensureUniqueFilename returns a unique filename if target already exists
func ensureUniqueFilename(targetPath string) string {
	if _, err := os.Stat(targetPath); errors.Is(err, os.ErrNotExist) {
		return targetPath
	}

	dir := filepath.Dir(targetPath)
	ext := filepath.Ext(targetPath)
	base := strings.TrimSuffix(filepath.Base(targetPath), ext)

	for i := 1; i < 10000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", base, i, ext))
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate
		}
	}

	return targetPath
}
