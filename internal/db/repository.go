package db

import (
	"database/sql"
	"errors"
	"time"
)

// Repository provides methods to query and update DB data
type Repository struct {
	db *DB
}

// NewRepository creates a new Repository instance
func NewRepository(db *DB) *Repository {
	return &Repository{db: db}
}

// CreateDownload inserts a new download item
func (r *Repository) CreateDownload(item *DownloadItem) error {
	query := `
	INSERT INTO downloads (id, name, type, source, save_path, total_bytes, downloaded_bytes, status, error_message, created_at, updated_at, metadata)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	_, err := r.db.Exec(query,
		item.ID,
		item.Name,
		item.Type,
		item.Source,
		item.SavePath,
		item.TotalBytes,
		item.DownloadedBytes,
		item.Status,
		item.ErrorMessage,
		item.CreatedAt,
		item.UpdatedAt,
		item.Metadata,
	)
	return err
}

// GetDownload retrieves a single download by ID
func (r *Repository) GetDownload(id string) (*DownloadItem, error) {
	query := `
	SELECT id, name, type, source, save_path, total_bytes, downloaded_bytes, status, error_message, created_at, updated_at, completed_at, metadata
	FROM downloads WHERE id = ?
	`
	row := r.db.QueryRow(query, id)

	var item DownloadItem
	var completedAt sql.NullTime

	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Type,
		&item.Source,
		&item.SavePath,
		&item.TotalBytes,
		&item.DownloadedBytes,
		&item.Status,
		&item.ErrorMessage,
		&item.CreatedAt,
		&item.UpdatedAt,
		&completedAt,
		&item.Metadata,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if completedAt.Valid {
		item.CompletedAt = &completedAt.Time
	}

	return &item, nil
}

// ListDownloads retrieves all downloads, ordered by created_at DESC
func (r *Repository) ListDownloads() ([]*DownloadItem, error) {
	query := `
	SELECT id, name, type, source, save_path, total_bytes, downloaded_bytes, status, error_message, created_at, updated_at, completed_at, metadata
	FROM downloads
	ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*DownloadItem
	for rows.Next() {
		var item DownloadItem
		var completedAt sql.NullTime

		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Type,
			&item.Source,
			&item.SavePath,
			&item.TotalBytes,
			&item.DownloadedBytes,
			&item.Status,
			&item.ErrorMessage,
			&item.CreatedAt,
			&item.UpdatedAt,
			&completedAt,
			&item.Metadata,
		)
		if err != nil {
			return nil, err
		}

		if completedAt.Valid {
			item.CompletedAt = &completedAt.Time
		}

		items = append(items, &item)
	}

	return items, nil
}

// UpdateProgress updates the progress and timestamp
func (r *Repository) UpdateProgress(id string, downloadedBytes int64, totalBytes int64, status DownloadStatus) error {
	query := `
	UPDATE downloads
	SET downloaded_bytes = ?, total_bytes = ?, status = ?, updated_at = ?
	WHERE id = ?
	`
	_, err := r.db.Exec(query, downloadedBytes, totalBytes, status, time.Now(), id)
	return err
}

// UpdateStatus updates download status and optional error message
func (r *Repository) UpdateStatus(id string, status DownloadStatus, errMsg string) error {
	var query string
	now := time.Now()

	if status == StatusCompleted {
		query = `UPDATE downloads SET status = ?, error_message = ?, updated_at = ?, completed_at = ? WHERE id = ?`
		_, err := r.db.Exec(query, status, errMsg, now, now, id)
		return err
	}

	query = `UPDATE downloads SET status = ?, error_message = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, status, errMsg, now, id)
	return err
}

// UpdateNameAndPath updates name, save_path, and metadata (used after torrent metadata is resolved)
func (r *Repository) UpdateDetails(id string, name string, savePath string, totalBytes int64, metadata string) error {
	query := `
	UPDATE downloads
	SET name = ?, save_path = ?, total_bytes = ?, metadata = ?, updated_at = ?
	WHERE id = ?
	`
	_, err := r.db.Exec(query, name, savePath, totalBytes, metadata, time.Now(), id)
	return err
}

// DeleteDownload removes a download from the database
func (r *Repository) DeleteDownload(id string) error {
	query := `DELETE FROM downloads WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// GetSetting reads a setting value
func (r *Repository) GetSetting(key string) (string, error) {
	var val string
	err := r.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&val)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return val, err
}

// SetSetting writes or updates a setting
func (r *Repository) SetSetting(key, val string) error {
	query := `
	INSERT INTO settings (key, value) VALUES (?, ?)
	ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`
	_, err := r.db.Exec(query, key, val)
	return err
}
