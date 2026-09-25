package cdn

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/voxire/onion/internal/platform/logger"
)

// fileURL builds the public URL for an uploaded file from its UUID.
func (c *CDN) fileURL(fileUUID string) string {
	return fmt.Sprintf("%s/files/%s", c.baseURL, fileUUID)
}

// uploadFile validates, stores the bytes in S3, and records the metadata row.
// It returns the persisted File (CreatedAt populated from the DB).
func (c *CDN) uploadFile(ctx context.Context, filename, mimeType string, size int64, body io.Reader) (*File, error) {
	if size > c.maxUpload {
		return nil, fmt.Errorf("file exceeds maximum upload size of %d bytes", c.maxUpload)
	}

	// Sniff the real content type from the bytes rather than trusting the
	// client's Content-Type, so an HTML/JS payload can't be stored under an
	// image/* type and later served as active content.
	sniffed, body, err := sniffMIME(body)
	if err != nil {
		return nil, fmt.Errorf("reading upload: %w", err)
	}
	mimeType, err = reconcileMIME(mimeType, sniffed)
	if err != nil {
		return nil, err
	}

	body, err = c.moderateImage(ctx, mimeType, size, body)
	if err != nil {
		return nil, err
	}

	fileUUID := uuid.NewString()
	storageKey := fmt.Sprintf("originals/%s%s", fileUUID, extensionForFormat(mimeTypeToFormat(mimeType)))
	if c.storage == nil {
		return nil, errCDNStorageNotConfigured
	}

	if err := c.storage.Upload(ctx, storageKey, mimeType, body); err != nil {
		// Upload can report an error after the object has reached the store;
		// make the same best-effort cleanup for the partial-success case.
		c.cleanupUploadedObject(ctx, storageKey)
		return nil, fmt.Errorf("uploading to storage: %w", err)
	}

	id, err := c.insertFile(ctx, fileUUID, storageKey, mimeType, filename, size)
	if err != nil {
		c.cleanupUploadedObject(ctx, storageKey)
		return nil, err
	}

	uploaded, err := c.fileByID(ctx, id)
	if err != nil {
		c.cleanupUploadedObject(ctx, storageKey)
		return nil, err
	}
	return uploaded, nil
}

const uploadCleanupTimeout = 5 * time.Second

func (c *CDN) cleanupUploadedObject(ctx context.Context, storageKey string) {
	if c.storage == nil {
		return
	}
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), uploadCleanupTimeout)
	defer cancel()
	if err := c.storage.Delete(cleanupCtx, storageKey); err != nil {
		logger.Ctx(ctx).Warn("cdn: failed to clean up orphaned upload",
			logger.String("storage_key", storageKey), logger.Error(err))
	}
}

// maxInMemoryUpload bounds how much of a multipart upload is buffered in RAM;
// anything larger spills to a temp file. The total request size is capped
// separately by http.MaxBytesReader.
const maxInMemoryUpload = 16 << 20 // 16 MiB

// upload handles POST /upload — a multipart form with a single "file" field.
func (c *CDN) upload(w http.ResponseWriter, r *http.Request) {
	// Cap the whole request body so a client can't exhaust disk/memory by
	// streaming past maxUpload — ParseMultipartForm alone only bounds the
	// in-memory portion and spills the rest to temp files.
	r.Body = http.MaxBytesReader(w, r.Body, c.maxUpload)

	// #nosec G120 -- MaxBytesReader caps the complete request before parsing.
	if err := r.ParseMultipartForm(maxInMemoryUpload); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "invalid multipart form", http.StatusBadRequest)
		}
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	if header.Size <= 0 {
		http.Error(w, "empty file", http.StatusBadRequest)
		return
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	uploaded, err := c.uploadFile(r.Context(), header.Filename, mimeType, header.Size, file)
	if err != nil {
		if errors.Is(err, errCDNStorageNotConfigured) {
			http.Error(w, "file uploads are not configured on this server", http.StatusNotImplemented)
			return
		}
		if errors.Is(err, ErrImageBlocked) {
			http.Error(w, "image rejected by content policy", http.StatusUnprocessableEntity)
			return
		}
		// Don't echo internal error details (storage client, DB text) to
		// the client; log them instead.
		logger.Ctx(r.Context()).Error("cdn: upload failed",
			logger.String("filename", header.Filename),
			logger.Error(err),
		)
		http.Error(w, "upload failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(uploadResponse{
		ID:   uploaded.FileUUID,
		URL:  c.fileURL(uploaded.FileUUID),
		MIME: uploaded.MIMEType,
		Size: uploaded.Size,
	}); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to encode upload response", logger.Error(err))
	}
}

// uploadResponse is the JSON body returned by a successful POST /upload. MIME
// is the type detected from the bytes (not necessarily what the client sent).
type uploadResponse struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	MIME string `json:"mime"`
	Size int64  `json:"size"`
}

// serveFile handles GET /files/{id} — streams the original bytes back with an
// immutable cache header (the UUID namespace means content never changes).
func (c *CDN) serveFile(w http.ResponseWriter, r *http.Request) {
	if !c.validateSignedDelivery(w, r) {
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing file id", http.StatusBadRequest)
		return
	}

	meta, err := c.fileByUUID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		logger.Ctx(r.Context()).Error("cdn: file lookup failed",
			logger.String("file_uuid", id), logger.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// The UUID namespace is immutable, so the UUID itself is a perfect ETag.
	// Answer conditional requests before fetching anything from storage.
	etag := quoteETag(meta.FileUUID)
	w.Header().Set("ETag", etag)
	if !meta.CreatedAt.IsZero() {
		w.Header().Set("Last-Modified", meta.CreatedAt.UTC().Format(http.TimeFormat))
	}
	w.Header().Set("Content-Type", meta.MIMEType)
	w.Header().Set("Cache-Control", deliveryCacheControl(r))
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", sanitizeFilename(meta.OriginalFilename)))
	if meta.Size > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", meta.Size))
	}
	if etagMatches(r.Header.Get("If-None-Match"), etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	if r.Method == http.MethodHead {
		return
	}

	reader, size, err := c.storage.Download(r.Context(), meta.StorageKey)
	if err != nil {
		logger.Ctx(r.Context()).Error("cdn: failed to download original",
			logger.String("file_uuid", id), logger.Error(err))
		http.Error(w, "failed to retrieve file", http.StatusInternalServerError)
		return
	}
	defer func() { _ = reader.Close() }()

	if size > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	}
	if _, err := io.Copy(w, reader); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to stream file", logger.Error(err))
	}
}
