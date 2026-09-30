package cdn

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// File is a row of cdn_files — an originally uploaded asset. ID is the BIGINT
// primary key; FileUUID is the public identifier used in URLs and cache keys.
// BusinessID is the owning tenant and is 0 for a row that cannot be
// attributed to one, which must never be treated as "matches every tenant".
type File struct {
	ID               int64
	FileUUID         string
	BusinessID       int64
	IsPublic         bool
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

// readScope is the set of cdn_files rows a delivery request may read. It is
// built only by the constructors below so a caller cannot assemble a
// combination the SQL would silently misread.
type readScope struct {
	// tenant is the caller's own business. Its rows are readable in addition
	// to the publicly flagged ones.
	tenant int64
	// signed marks a request carrying a validated delivery signature. The
	// signature is the credential, so tenancy does not apply — this is what
	// makes signed delivery of a private asset work.
	signed bool
}

// anonymousRead is the zero scope: public assets only. It is the default for
// any request that arrives without a resolved tenant.
var anonymousRead = readScope{}

// tenantRead returns the scope of a caller authenticated for businessID.
func tenantRead(businessID int64) readScope {
	if businessID <= 0 {
		return anonymousRead
	}
	return readScope{tenant: businessID}
}

// signedRead is the scope of a request whose delivery signature validated.
func signedRead() readScope { return readScope{signed: true} }

// where renders the row predicate for fileUUID. Every read of cdn_files goes
// through here, so there is exactly one place where the visibility rule lives.
func (s readScope) where(fileUUID string) (string, []any) {
	switch {
	case s.signed:
		return `file_uuid = ?`, []any{fileUUID}
	case s.tenant > 0:
		return `file_uuid = ? AND (business_id = ? OR is_public = TRUE)`, []any{fileUUID, s.tenant}
	default:
		return `file_uuid = ? AND is_public = TRUE`, []any{fileUUID}
	}
}

const fileColumns = `id, file_uuid, business_id, is_public, storage_key, mime_type, size, original_filename, created_at`

func scanFile(row *sql.Row) (*File, error) {
	f := &File{}
	if err := row.Scan(&f.ID, &f.FileUUID, &f.BusinessID, &f.IsPublic, &f.StorageKey, &f.MIMEType, &f.Size, &f.OriginalFilename, &f.CreatedAt); err != nil {
		return nil, err
	}
	return f, nil
}

// insertFile records a new asset under businessID. public makes the asset
// anonymously readable; every caller must pass it explicitly, because a stored
// file that is neither tenant-scoped nor deliberately public would be
// unreachable.
func (c *CDN) insertFile(ctx context.Context, businessID int64, fileUUID, storageKey, mimeType, filename string, size int64, public bool) (int64, error) {
	res, err := c.db.ExecContext(ctx,
		`INSERT INTO cdn_files (file_uuid, business_id, is_public, storage_key, mime_type, size, original_filename)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		fileUUID, businessID, public, storageKey, mimeType, size, filename,
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

// fileForRead resolves a file for delivery under the given scope. A file in
// another tenant is reported as missing, not as forbidden, so the row id
// space is not enumerable across tenants.
func (c *CDN) fileForRead(ctx context.Context, scope readScope, fileUUID string) (*File, error) {
	condition, args := scope.where(fileUUID)
	return scanFile(c.db.QueryRowContext(ctx, `SELECT `+fileColumns+` FROM cdn_files WHERE `+condition, args...))
}

// fileByID reads a file the caller has already proven it owns. businessID must
// be a resolved tenant; a zero businessID matches no row at all, so an
// unresolved caller can never read by numeric id.
func (c *CDN) fileByID(ctx context.Context, businessID, id int64) (*File, error) {
	if businessID <= 0 {
		return nil, sql.ErrNoRows
	}
	return scanFile(c.db.QueryRowContext(ctx,
		`SELECT `+fileColumns+` FROM cdn_files WHERE id = ? AND business_id = ?`, id, businessID))
}

// fileOwnsUUID reports whether fileUUID belongs to businessID. Used to check a
// delivery/purge target that arrived as a URL rather than as a path parameter.
func (c *CDN) fileOwnsUUID(ctx context.Context, businessID int64, fileUUID string) (bool, error) {
	if businessID <= 0 || fileUUID == "" {
		return false, nil
	}
	var one int
	err := c.db.QueryRowContext(ctx,
		`SELECT 1 FROM cdn_files WHERE file_uuid = ? AND business_id = ?`, fileUUID, businessID).Scan(&one)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("checking cdn_files ownership: %w", err)
	}
	return true, nil
}

// transformationByCacheKey resolves a cached derivative. Cache keys are derived
// from a file uuid, so the row is pinned to the owning business rather than
// trusted on the key alone.
func (c *CDN) transformationByCacheKey(ctx context.Context, businessID int64, cacheKey string) (*Transformation, error) {
	if businessID <= 0 {
		return nil, sql.ErrNoRows
	}
	t := &Transformation{}
	err := c.db.QueryRowContext(ctx,
		`SELECT t.id, t.file_id, t.cache_key, t.width, t.height, t.format, t.quality, t.fit, t.storage_key, t.created_at
		 FROM cdn_transformations t
		 JOIN cdn_files f ON f.id = t.file_id
		 WHERE t.cache_key = ? AND f.business_id = ?`,
		cacheKey, businessID,
	).Scan(&t.ID, &t.FileID, &t.CacheKey, &t.Width, &t.Height, &t.Format, &t.Quality, &t.Fit, &t.StorageKey, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// uploadKeyOwnerByHash returns the owner and business of an active, unexpired
// API key whose SHA-256 hash matches and whose scopes include "upload". ok is
// false when no such key exists. Used by the write-endpoint auth middleware so
// any user can authorize uploads with a personal `upload`-scoped key (managed
// in Dash), not just the shared CDN_API_KEY. Reads the hub_api_keys table
// directly — the CDN runmode doesn't load the api_keys module.
func (c *CDN) uploadKeyOwnerByHash(ctx context.Context, keyHash string) (ownerID, businessID int64, ok bool) {
	err := c.db.QueryRowContext(ctx,
		`SELECT owner_id, business_id FROM hub_api_keys
		 WHERE key_hash = ?
		   AND is_active = TRUE
		   AND (expires_at IS NULL OR expires_at > NOW())
		   AND JSON_CONTAINS(scopes, JSON_QUOTE('upload'))
		 LIMIT 1`,
		keyHash,
	).Scan(&ownerID, &businessID)
	if err != nil {
		return 0, 0, false
	}
	return ownerID, businessID, true
}

// insertTransformation records a generated derivative. The INSERT selects the
// parent row rather than naming it blindly, so a derivative can only ever be
// attached to a file the caller has proven it owns. A non-matching parent
// inserts nothing (the derivative is simply regenerated later) rather than
// erroring, which keeps the concurrent-miss INSERT IGNORE behaviour intact.
func (c *CDN) insertTransformation(ctx context.Context, businessID, fileID int64, cacheKey, storageKey string, p Params) error {
	if businessID <= 0 {
		return sql.ErrNoRows
	}
	_, err := c.db.ExecContext(ctx,
		`INSERT IGNORE INTO cdn_transformations (file_id, cache_key, width, height, format, quality, fit, storage_key)
		 SELECT f.id, ?, ?, ?, ?, ?, ?, ?
		   FROM cdn_files f
		  WHERE f.id = ? AND f.business_id = ?`,
		cacheKey, p.Width, p.Height, p.Format, p.Quality, p.Fit, storageKey, fileID, businessID,
	)
	if err != nil {
		return fmt.Errorf("inserting cdn_transformations row: %w", err)
	}
	return nil
}

// transformationStorageKeysByFileUUID lists the cached derivatives of a file
// the caller owns.
func (c *CDN) transformationStorageKeysByFileUUID(ctx context.Context, businessID int64, fileUUID string) ([]string, error) {
	if businessID <= 0 {
		return nil, sql.ErrNoRows
	}
	rows, err := c.db.QueryContext(ctx,
		`SELECT t.storage_key
		   FROM cdn_transformations t
		   JOIN cdn_files f ON f.id = t.file_id
		  WHERE f.file_uuid = ? AND f.business_id = ?`,
		fileUUID, businessID,
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

// deleteTransformationsByFileUUID drops a file's cached derivative rows. The
// join carries the tenant filter, so a purge cannot remove another tenant's
// rows even if the uuid resolved.
func (c *CDN) deleteTransformationsByFileUUID(ctx context.Context, businessID int64, fileUUID string) (int64, error) {
	if businessID <= 0 {
		return 0, sql.ErrNoRows
	}
	res, err := c.db.ExecContext(ctx,
		`DELETE t
		   FROM cdn_transformations t
		   JOIN cdn_files f ON f.id = t.file_id
		  WHERE f.file_uuid = ? AND f.business_id = ?`,
		fileUUID, businessID,
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

// storageKeyFileUUID recovers the cdn_files uuid from a storage key the CDN
// itself wrote, so a purge or sign target supplied as `?url=` or as a bare
// storage path is held to the same tenant check as /files/{uuid}.
func storageKeyFileUUID(key string) (string, bool) {
	parts := strings.Split(strings.Trim(strings.TrimSpace(key), "/"), "/")
	var name string
	switch {
	case len(parts) == 2 && parts[0] == "originals":
		// originals/<uuid>.<ext>
		name = parts[1]
	case len(parts) == 3 && parts[0] == "transformed":
		// transformed/<uuid>/<cache key>.<ext>
		name = parts[1]
	default:
		return "", false
	}
	if name == "" || strings.Contains(name, "/") {
		return "", false
	}
	uuid, _, hadExtension := strings.Cut(name, ".")
	if uuid == "" {
		return "", false
	}
	if hadExtension && uuid == name {
		// The whole segment was the extension match target with no stem.
		return "", false
	}
	return uuid, true
}
