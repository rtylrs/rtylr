package cdn

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/platform/logger"
)

type purgeRequest struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// errBadPurgeInput marks purge targets the caller can fix (unparseable URL,
// unsupported target shape). Handlers answer those with 400; every other
// failure is an infrastructure error and must not leak into the response.
var errBadPurgeInput = errors.New("invalid purge target")

type purgeResponse struct {
	Scope          string `json:"scope"`
	ID             string `json:"id,omitempty"`
	URL            string `json:"url,omitempty"`
	DeletedObjects int    `json:"deleted_objects"`
	DeletedRows    int64  `json:"deleted_rows,omitempty"`
}

func (c *CDN) purgeFileVariants(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		id = strings.TrimSpace(r.URL.Query().Get("id"))
	}
	if id == "" {
		http.Error(w, "missing file id", http.StatusBadRequest)
		return
	}

	deletedObjects, deletedRows, err := c.purgeFileVariantFamily(r, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		logger.Ctx(r.Context()).Error("cdn: file purge failed", logger.String("file_uuid", id), logger.Error(err))
		http.Error(w, "purge failed", http.StatusInternalServerError)
		return
	}

	writePurgeResponse(w, purgeResponse{
		Scope:          "file",
		ID:             id,
		DeletedObjects: deletedObjects,
		DeletedRows:    deletedRows,
	})
}

func (c *CDN) purgeUploadURL(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("url"))
	if id := strings.TrimSpace(r.URL.Query().Get("id")); id != "" {
		deletedObjects, deletedRows, err := c.purgeFileVariantFamily(r, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "file not found", http.StatusNotFound)
				return
			}
			logger.Ctx(r.Context()).Error("cdn: file purge failed", logger.String("file_uuid", id), logger.Error(err))
			http.Error(w, "purge failed", http.StatusInternalServerError)
			return
		}
		writePurgeResponse(w, purgeResponse{
			Scope:          "file",
			ID:             id,
			DeletedObjects: deletedObjects,
			DeletedRows:    deletedRows,
		})
		return
	}
	if raw == "" {
		http.Error(w, "missing url", http.StatusBadRequest)
		return
	}
	deleted, err := c.purgeUploadURLFamily(r, raw)
	if err != nil {
		writePurgeUploadURLError(w, r, raw, err)
		return
	}
	writePurgeResponse(w, purgeResponse{
		Scope:          "upload-url",
		URL:            raw,
		DeletedObjects: deleted,
	})
}

func (c *CDN) purgeControl(w http.ResponseWriter, r *http.Request) {
	var req purgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid purge request", http.StatusBadRequest)
		return
	}
	req.ID = strings.TrimSpace(req.ID)
	req.URL = strings.TrimSpace(req.URL)
	switch {
	case req.ID != "":
		deletedObjects, deletedRows, err := c.purgeFileVariantFamily(r, req.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "file not found", http.StatusNotFound)
				return
			}
			logger.Ctx(r.Context()).Error("cdn: file purge failed", logger.String("file_uuid", req.ID), logger.Error(err))
			http.Error(w, "purge failed", http.StatusInternalServerError)
			return
		}
		writePurgeResponse(w, purgeResponse{
			Scope:          "file",
			ID:             req.ID,
			DeletedObjects: deletedObjects,
			DeletedRows:    deletedRows,
		})
	case req.URL != "":
		deleted, err := c.purgeUploadURLFamily(r, req.URL)
		if err != nil {
			writePurgeUploadURLError(w, r, req.URL, err)
			return
		}
		writePurgeResponse(w, purgeResponse{
			Scope:          "upload-url",
			URL:            req.URL,
			DeletedObjects: deleted,
		})
	default:
		http.Error(w, "missing id or url", http.StatusBadRequest)
	}
}

func (c *CDN) purgeFileVariantFamily(r *http.Request, id string) (int, int64, error) {
	meta, err := c.fileByUUID(r.Context(), id)
	if err != nil {
		return 0, 0, err
	}
	keys, err := c.transformationStorageKeysByFileUUID(r.Context(), meta.FileUUID)
	if err != nil {
		return 0, 0, err
	}

	deletedObjects := 0
	if c.storage != nil {
		for _, key := range keys {
			if err := c.storage.Delete(r.Context(), key); err != nil {
				return deletedObjects, 0, err
			}
			deletedObjects++
		}
	}
	deletedRows, err := c.deleteTransformationsByFileUUID(r.Context(), meta.FileUUID)
	if err != nil {
		return deletedObjects, 0, err
	}
	return deletedObjects, deletedRows, nil
}

func (c *CDN) purgeUploadURLFamily(r *http.Request, raw string) (int, error) {
	source, err := c.resolveUploadURLSource(raw, r.Host)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", errBadPurgeInput, err)
	}
	seed := ""
	switch {
	case source.storageKey != "":
		seed = "storage:" + source.storageKey
	case source.remoteURL != nil:
		seed = "remote:" + source.remoteURL.String()
	case source.redirectURL != "":
		if id, ok := fileUUIDFromCanonicalURL(source.redirectURL); ok {
			deletedObjects, _, err := c.purgeFileVariantFamily(r, id)
			return deletedObjects, err
		}
	}
	if seed == "" {
		return 0, fmt.Errorf("%w: unsupported target", errBadPurgeInput)
	}
	if c.storage == nil {
		return 0, nil
	}
	deleted, err := c.storage.DeletePrefix(r.Context(), uploadURLTransformStoragePrefix(seed))
	if err != nil {
		return deleted, err
	}
	return deleted, nil
}

// writePurgeUploadURLError maps a purgeUploadURLFamily failure onto the same
// status codes the by-id purge routes use: 404 for a missing file, 400 for
// input the caller can fix, and a generic 500 otherwise. Driver and storage
// errors are logged, never echoed — they carry backend details the caller must
// not see.
func writePurgeUploadURLError(w http.ResponseWriter, r *http.Request, raw string, err error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		http.Error(w, "file not found", http.StatusNotFound)
	case errors.Is(err, errBadPurgeInput):
		http.Error(w, "invalid purge request", http.StatusBadRequest)
	default:
		logger.Ctx(r.Context()).Error("cdn: upload-url purge failed", logger.String("url", raw), logger.Error(err))
		http.Error(w, "purge failed", http.StatusInternalServerError)
	}
}

func fileUUIDFromCanonicalURL(raw string) (string, bool) {
	for _, marker := range []string{"/images/", "/files/"} {
		idx := strings.Index(raw, marker)
		if idx < 0 {
			continue
		}
		id := raw[idx+len(marker):]
		if q := strings.Index(id, "?"); q >= 0 {
			id = id[:q]
		}
		if id != "" && !strings.Contains(id, "/") {
			return id, true
		}
	}
	return "", false
}

func writePurgeResponse(w http.ResponseWriter, resp purgeResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Global().Warn("cdn: failed to encode purge response", logger.Error(err))
	}
}
