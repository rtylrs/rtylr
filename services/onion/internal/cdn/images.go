package cdn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/platform/logger"
)

// engineTimeout caps how long a single media-engine invocation may run before
// it is killed, so a malformed or pathological input can't pin a worker.
const engineTimeout = 60 * time.Second

// transform returns a transformed image, generating it via the media-engine on
// a cache miss and persisting the result to S3 + cdn_transformations. fileID is
// the BIGINT key; fileUUID is the public id used for cache keys and storage paths.
func (c *CDN) transform(ctx context.Context, fileID int64, fileUUID, originalStorageKey string, p Params) (io.ReadCloser, int64, string, error) {
	p.Defaults()
	if err := p.Validate(); err != nil {
		return nil, 0, "", fmt.Errorf("invalid params: %w", err)
	}

	key := p.CacheKey(fileUUID)
	mimeType := mimeForFormat(p.Format)

	// Cache hit — stream the already-transformed asset straight from S3.
	cached, err := c.transformationByCacheKey(ctx, key)
	if err == nil {
		reader, size, dlErr := c.storage.Download(ctx, cached.StorageKey)
		if dlErr != nil {
			return nil, 0, "", fmt.Errorf("downloading cached transformation: %w", dlErr)
		}
		return reader, size, mimeType, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, 0, "", fmt.Errorf("checking transformation cache: %w", err)
	}

	// Cache miss — pull the original to a temp file, run the media-engine,
	// upload the result, record it, then stream it back.
	tmpIn, tmpOut, cleanup, err := c.stageOriginal(ctx, originalStorageKey)
	if err != nil {
		return nil, 0, "", err
	}
	defer cleanup()

	// Bound concurrent media-engine processes; block until a slot frees up or
	// the request is cancelled.
	select {
	case c.engineSem <- struct{}{}:
		defer func() { <-c.engineSem }()
	case <-ctx.Done():
		return nil, 0, "", ctx.Err()
	}

	execCtx, cancel := context.WithTimeout(ctx, engineTimeout)
	defer cancel()
	// #nosec G204 -- mediaEngine is a server-side configured binary path; CLI args are built from validated params and temp paths.
	cmd := exec.CommandContext(execCtx, c.mediaEngine, p.CLIArgs(tmpIn, tmpOut)...)
	if out, runErr := cmd.CombinedOutput(); runErr != nil {
		return nil, 0, "", fmt.Errorf("media-engine failed: %w\noutput: %s", runErr, out)
	}

	resultFile, err := os.Open(tmpOut)
	if err != nil {
		return nil, 0, "", fmt.Errorf("opening media-engine result: %w", err)
	}
	defer func() { _ = resultFile.Close() }()

	transformedKey := p.StorageKey(fileUUID, key)
	if err := c.storage.Upload(ctx, transformedKey, mimeType, resultFile); err != nil {
		return nil, 0, "", fmt.Errorf("uploading transformation: %w", err)
	}
	if err := c.insertTransformation(ctx, fileID, key, transformedKey, p); err != nil {
		return nil, 0, "", err
	}

	logger.Ctx(ctx).Info("cdn: generated transformation",
		logger.String("file_uuid", fileUUID),
		logger.String("format", p.Format),
		logger.Any("width", p.Width),
		logger.Any("height", p.Height),
	)

	// Re-download so the streamed size/headers match exactly what S3 stored.
	reader, size, err := c.storage.Download(ctx, transformedKey)
	if err != nil {
		return nil, 0, "", fmt.Errorf("streaming transformation: %w", err)
	}
	return reader, size, mimeType, nil
}

// stageOriginal downloads the original object to a temp input file and creates
// an empty temp output file for the media-engine. The returned cleanup removes
// both. Caller must call cleanup once done with the output file.
func (c *CDN) stageOriginal(ctx context.Context, originalStorageKey string) (inPath, outPath string, cleanup func(), err error) {
	origReader, _, err := c.storage.Download(ctx, originalStorageKey)
	if err != nil {
		return "", "", func() {}, fmt.Errorf("downloading original: %w", err)
	}
	defer func() { _ = origReader.Close() }()

	tmpIn, err := os.CreateTemp("", "cdn-input-*")
	if err != nil {
		return "", "", func() {}, fmt.Errorf("creating temp input: %w", err)
	}
	if _, err := io.Copy(tmpIn, origReader); err != nil {
		_ = tmpIn.Close()
		_ = os.Remove(tmpIn.Name())
		return "", "", func() {}, fmt.Errorf("writing temp input: %w", err)
	}
	_ = tmpIn.Close()

	tmpOut, err := os.CreateTemp("", "cdn-output-*")
	if err != nil {
		_ = os.Remove(tmpIn.Name())
		return "", "", func() {}, fmt.Errorf("creating temp output: %w", err)
	}
	_ = tmpOut.Close()

	cleanup = func() {
		_ = os.Remove(tmpIn.Name())
		_ = os.Remove(tmpOut.Name())
	}
	return tmpIn.Name(), tmpOut.Name(), cleanup, nil
}

// serveTransformed handles GET /images/{id}?w=&h=&fit=&format=&q=&t=
// where t= is an optional Cloudinary-style transformation chain.
func (c *CDN) serveTransformed(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "id")
	if fileUUID == "" {
		http.Error(w, "missing image id", http.StatusBadRequest)
		return
	}

	// An explicit ?format= wins; otherwise negotiate from the Accept header so
	// modern clients get AVIF/WebP. The response varies on Accept accordingly.
	format := r.URL.Query().Get("format")
	if format == "" {
		if negotiated := negotiateFormat(r.Header.Get("Accept")); negotiated != "" {
			format = negotiated
			w.Header().Set("Vary", "Accept")
		}
	}

	p := Params{
		Width:   queryInt(r, "w"),
		Height:  queryInt(r, "h"),
		Fit:     r.URL.Query().Get("fit"),
		Format:  format,
		Quality: queryInt(r, "q"),
	}

	// ?t= carries a Cloudinary-style transformation chain (e.g.
	// "w_400,c_fill,g_auto/e_blur:300"). When present it drives the transform,
	// and any f_/q_ inside it sets the output format/quality the CDN serves.
	if chain := r.URL.Query().Get("t"); chain != "" {
		p.Chain = chain
		chainFmt, chainQ := parseChainParams(chain)
		if chainFmt != "" && chainFmt != "auto" {
			p.Format = chainFmt
		} else if p.Format == "" || p.Format == "auto" {
			p.Format = "webp"
		}
		if chainQ > 0 {
			p.Quality = chainQ
		}
	}

	// Validate the request before touching the DB so malformed query params
	// fail fast with a 400 instead of a 500 from deep in transform().
	p.Defaults()
	if err := p.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	meta, err := c.fileByUUID(r.Context(), fileUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		logger.Ctx(r.Context()).Error("cdn: file lookup failed",
			logger.String("file_uuid", fileUUID), logger.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// The cache key fully determines the transformed bytes, so it is the ETag.
	// Short-circuit conditional requests before doing any transform/S3 work.
	etag := quoteETag(p.CacheKey(meta.FileUUID))
	w.Header().Set("ETag", etag)
	if etagMatches(r.Header.Get("If-None-Match"), etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	reader, size, mimeType, err := c.transform(r.Context(), meta.ID, meta.FileUUID, meta.StorageKey, p)
	if err != nil {
		logger.Ctx(r.Context()).Error("cdn: transform failed",
			logger.String("file_uuid", fileUUID), logger.Error(err))
		http.Error(w, "image transformation failed", http.StatusInternalServerError)
		return
	}
	defer func() { _ = reader.Close() }()

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", immutableCacheControl)
	if size > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	}
	if _, err := io.Copy(w, reader); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to stream transformation", logger.Error(err))
	}
}

func queryInt(r *http.Request, key string) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}
