package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
	if !isAllowedMIMEType(mimeType) {
		return nil, fmt.Errorf("unsupported MIME type: %s", mimeType)
	}
	if size > c.maxUpload {
		return nil, fmt.Errorf("file exceeds maximum upload size of %d bytes", c.maxUpload)
	}

	fileUUID := uuid.NewString()
	storageKey := fmt.Sprintf("originals/%s%s", fileUUID, extensionForFormat(mimeTypeToFormat(mimeType)))

	if err := c.storage.Upload(ctx, storageKey, mimeType, body); err != nil {
		return nil, fmt.Errorf("uploading to storage: %w", err)
	}

	id, err := c.insertFile(ctx, fileUUID, storageKey, mimeType, filename, size)
	if err != nil {
		return nil, err
	}

	return c.fileByID(ctx, id)
}

// upload handles POST /upload — a multipart form with a single "file" field.
func (c *CDN) upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(c.maxUpload); err != nil {
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	uploaded, err := c.uploadFile(r.Context(), header.Filename, mimeType, header.Size, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"id":  uploaded.FileUUID,
		"url": c.fileURL(uploaded.FileUUID),
	}); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to encode upload response", logger.Error(err))
	}
}

// serveFile handles GET /files/{id} — streams the original bytes back with an
// immutable cache header (the UUID namespace means content never changes).
func (c *CDN) serveFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing file id", http.StatusBadRequest)
		return
	}

	meta, err := c.fileByUUID(r.Context(), id)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
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

	w.Header().Set("Content-Type", meta.MIMEType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	if size > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	}
	if _, err := io.Copy(w, reader); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to stream file", logger.Error(err))
	}
}
