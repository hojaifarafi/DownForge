package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps the sql.DB instance
type DB struct {
	*sql.DB
}

// InitDB initializes SQLite database connection and runs initial schema migrations
func InitDB(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// Open SQLite database with WAL mode for high concurrent read/write throughput
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)", dbPath)
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{database}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// migrate creates necessary tables if they don't exist
func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS downloads (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		source TEXT NOT NULL,
		save_path TEXT NOT NULL,
		total_bytes INTEGER DEFAULT 0,
		downloaded_bytes INTEGER DEFAULT 0,
		status TEXT NOT NULL,
		error_message TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME,
		metadata TEXT DEFAULT ''
	);

	CREATE INDEX IF NOT EXISTS idx_downloads_status ON downloads(status);
	CREATE INDEX IF NOT EXISTS idx_downloads_created_at ON downloads(created_at);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);
	`

	_, err := db.Exec(schema)
	return err
}
