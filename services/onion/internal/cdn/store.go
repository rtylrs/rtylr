package cdn

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// File is a row of cdn_files — an originally uploaded asset. ID is the BIGINT
// primary key; FileUUID is the public identifier used in URLs and cache keys.
type File struct {
	ID               int64
	FileUUID         string
	StorageKey       string
	MIMEType         string
	Size             int64
	OriginalFilename string
	CreatedAt        time.Time
}

// Transformation is a row of cdn_transformations — a cached, resized/re-encoded
// derivative of a File, keyed by a deterministic CacheKey.
type Transformation struct {
	ID         int64
	FileID     int64
	CacheKey   string
	Width      int
	Height     int
	Format     string
	Quality    int
	Fit        string
	StorageKey string
	CreatedAt  time.Time
}

const fileColumns = `id, file_uuid, storage_key, mime_type, size, original_filename, created_at`

func scanFile(row *sql.Row) (*File, error) {
	f := &File{}
	if err := row.Scan(&f.ID, &f.FileUUID, &f.StorageKey, &f.MIMEType, &f.Size, &f.OriginalFilename, &f.CreatedAt); err != nil {
		return nil, err
	}
	return f, nil
}

func (c *CDN) insertFile(ctx context.Context, fileUUID, storageKey, mimeType, filename string, size int64) (int64, error) {
	res, err := c.db.ExecContext(ctx,
		`INSERT INTO cdn_files (file_uuid, storage_key, mime_type, size, original_filename)
		 VALUES (?, ?, ?, ?, ?)`,
		fileUUID, storageKey, mimeType, size, filename,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting cdn_files row: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("reading cdn_files insert id: %w", err)
	}
	return id, nil
}

func (c *CDN) fileByUUID(ctx context.Context, fileUUID string) (*File, error) {
	return scanFile(c.db.QueryRowContext(ctx,
		`SELECT `+fileColumns+` FROM cdn_files WHERE file_uuid = ?`, fileUUID))
}

func (c *CDN) fileByID(ctx context.Context, id int64) (*File, error) {
	return scanFile(c.db.QueryRowContext(ctx,
		`SELECT `+fileColumns+` FROM cdn_files WHERE id = ?`, id))
}

func (c *CDN) transformationByCacheKey(ctx context.Context, cacheKey string) (*Transformation, error) {
	t := &Transformation{}
	err := c.db.QueryRowContext(ctx,
		`SELECT id, file_id, cache_key, width, height, format, quality, fit, storage_key, created_at
		 FROM cdn_transformations WHERE cache_key = ?`,
		cacheKey,
	).Scan(&t.ID, &t.FileID, &t.CacheKey, &t.Width, &t.Height, &t.Format, &t.Quality, &t.Fit, &t.StorageKey, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// uploadKeyOwnerByHash returns the owner of an active, unexpired API key whose
// SHA-256 hash matches and whose scopes include "upload". ok is false when no
// such key exists. Used by the write-endpoint auth middleware so any user can
// authorize uploads with a personal `upload`-scoped key (managed in Dash), not
// just the shared CDN_API_KEY. Reads the hub_api_keys table directly — the CDN
// runmode doesn't load the api_keys module.
func (c *CDN) uploadKeyOwnerByHash(ctx context.Context, keyHash string) (ownerID int64, ok bool) {
	err := c.db.QueryRowContext(ctx,
		`SELECT owner_id FROM hub_api_keys
		 WHERE key_hash = ?
		   AND is_active = TRUE
		   AND (expires_at IS NULL OR expires_at > NOW())
		   AND JSON_CONTAINS(scopes, JSON_QUOTE('upload'))
		 LIMIT 1`,
		keyHash,
	).Scan(&ownerID)
	if err != nil {
		return 0, false
	}
	return ownerID, true
}

func (c *CDN) insertTransformation(ctx context.Context, fileID int64, cacheKey, storageKey string, p Params) error {
	// INSERT IGNORE: two concurrent requests for the same uncached
	// transformation both generate it; the second insert is a harmless no-op.
	_, err := c.db.ExecContext(ctx,
		`INSERT IGNORE INTO cdn_transformations (file_id, cache_key, width, height, format, quality, fit, storage_key)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		fileID, cacheKey, p.Width, p.Height, p.Format, p.Quality, p.Fit, storageKey,
	)
	if err != nil {
		return fmt.Errorf("inserting cdn_transformations row: %w", err)
	}
	return nil
}

func (c *CDN) transformationStorageKeysByFileUUID(ctx context.Context, fileUUID string) ([]string, error) {
	rows, err := c.db.QueryContext(ctx,
		`SELECT t.storage_key
		   FROM cdn_transformations t
		   JOIN cdn_files f ON f.id = t.file_id
		  WHERE f.file_uuid = ?`,
		fileUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing cdn transformations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	keys := []string{}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("scanning cdn transformation key: %w", err)
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating cdn transformation keys: %w", err)
	}
	return keys, nil
}

func (c *CDN) deleteTransformationsByFileUUID(ctx context.Context, fileUUID string) (int64, error) {
	res, err := c.db.ExecContext(ctx,
		`DELETE t
		   FROM cdn_transformations t
		   JOIN cdn_files f ON f.id = t.file_id
		  WHERE f.file_uuid = ?`,
		fileUUID,
	)
	if err != nil {
		return 0, fmt.Errorf("deleting cdn transformations: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("reading deleted transformation count: %w", err)
	}
	return rows, nil
}
