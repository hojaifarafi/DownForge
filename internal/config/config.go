package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// Config holds the application configuration
type Config struct {
	Host              string `json:"host"`
	Port              int    `json:"port"`
	DownloadDir       string `json:"download_dir"`
	DBPath            string `json:"db_path"`
	MaxConcurrent     int    `json:"max_concurrent"`
	TorrentPort       int    `json:"torrent_port"`
	EnableDHT         bool   `json:"enable_dht"`
	EnablePEX         bool   `json:"enable_pex"`
	UploadRateLimit   int    `json:"upload_rate_limit"`   // in KB/s, 0 for unlimited
	DownloadRateLimit int    `json:"download_rate_limit"` // in KB/s, 0 for unlimited
}

// getAppDir returns the directory for storing configuration and database
func getAppDir() string {
	homeDir, err := os.UserHomeDir()
	if err == nil {
		appDir := filepath.Join(homeDir, ".downforge")
		if err := os.MkdirAll(appDir, 0755); err == nil {
			return appDir
		}
	}

	// Fallback to local ./data folder if home directory is read-only or inaccessible
	localDir := "./data"
	_ = os.MkdirAll(localDir, 0755)
	return localDir
}

// getDownloadsDir returns the default downloads folder
func getDownloadsDir() string {
	homeDir, err := os.UserHomeDir()
	if err == nil {
		downloads := filepath.Join(homeDir, "Downloads", "DownForge")
		if runtime.GOOS == "windows" {
			downloads = filepath.Join(homeDir, "Downloads", "DownForge")
		}
		if err := os.MkdirAll(downloads, 0755); err == nil {
			return downloads
		}
	}

	// Fallback to local ./downloads directory
	localDownloads := "./downloads"
	_ = os.MkdirAll(localDownloads, 0755)
	return localDownloads
}

// DefaultConfig returns reasonable default configuration values
func DefaultConfig() *Config {
	appDir := getAppDir()
	downloadDir := getDownloadsDir()

	return &Config{
		Host:              "0.0.0.0",
		Port:              8080,
		DownloadDir:       downloadDir,
		DBPath:            filepath.Join(appDir, "downforge.db"),
		MaxConcurrent:     3,
		TorrentPort:       42070,
		EnableDHT:         true,
		EnablePEX:         true,
		UploadRateLimit:   0,
		DownloadRateLimit: 0,
	}
}

// ConfigFilePath returns the path to config.json
func ConfigFilePath() string {
	appDir := getAppDir()
	return filepath.Join(appDir, "config.json")
}

// LoadConfig loads the configuration from disk or returns defaults
func LoadConfig() *Config {
	cfg := DefaultConfig()
	filePath := ConfigFilePath()

	data, err := os.ReadFile(filePath)
	if err == nil {
		_ = json.Unmarshal(data, cfg)
	} else {
		_ = SaveConfig(cfg)
	}

	// Ensure download directory exists
	_ = os.MkdirAll(cfg.DownloadDir, 0755)

	return cfg
}

// SaveConfig saves configuration to the disk
func SaveConfig(cfg *Config) error {
	filePath := ConfigFilePath()
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}
