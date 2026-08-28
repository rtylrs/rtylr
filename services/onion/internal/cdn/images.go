package cdn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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

	// Run the shared generation on a cancellation-detached context: if the
	// leader's client disconnects, the followers collapsed onto this call
	// would all fail with it. Each waiter still honors its own ctx below.
	genCtx := context.WithoutCancel(ctx)
	resultCh := c.transformGroup.DoChan(key, func() (any, error) {
		cached, err := c.transformationByCacheKey(genCtx, key)
		if err == nil {
			return cached.StorageKey, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("checking transformation cache: %w", err)
		}
		return c.generateTransformation(genCtx, fileID, fileUUID, originalStorageKey, key, mimeType, p)
	})

	var transformedKey string
	select {
	case result := <-resultCh:
		if result.Err != nil {
			return nil, 0, "", result.Err
		}
		var ok bool
		transformedKey, ok = result.Val.(string)
		if !ok || transformedKey == "" {
			return nil, 0, "", fmt.Errorf("media-engine returned an empty transformation key")
		}
	case <-ctx.Done():
		return nil, 0, "", ctx.Err()
	}

	reader, size, err := c.storage.Download(ctx, transformedKey)
	if err != nil {
		return nil, 0, "", fmt.Errorf("streaming transformation: %w", err)
	}
	return reader, size, mimeType, nil
}

func (c *CDN) generateTransformation(ctx context.Context, fileID int64, fileUUID, originalStorageKey, key, mimeType string, p Params) (string, error) {
	// Cache miss — pull the original to a temp file, run the media-engine,
	// upload the result, and record it.
	tmpIn, tmpOut, cleanup, err := c.stageOriginal(ctx, originalStorageKey)
	if err != nil {
		return "", err
	}
	defer cleanup()

	// Bound concurrent media-engine processes; block until a slot frees up or
	// the request is cancelled.
	select {
	case c.engineSem <- struct{}{}:
		defer func() { <-c.engineSem }()
	case <-ctx.Done():
		return "", ctx.Err()
	}

	execCtx, cancel := context.WithTimeout(ctx, engineTimeout)
	defer cancel()
	if err := c.runTransformed(execCtx, p.CLIArgs(tmpIn, tmpOut)); err != nil {
		return "", err
	}

	// #nosec G304 -- tmpOut is created by the server in its private temporary directory.
	resultFile, err := os.Open(tmpOut)
	if err != nil {
		return "", fmt.Errorf("opening media-engine result: %w", err)
	}
	defer func() { _ = resultFile.Close() }()

	transformedKey := p.StorageKey(fileUUID, key)
	if err := c.storage.Upload(ctx, transformedKey, mimeType, resultFile); err != nil {
		return "", fmt.Errorf("uploading transformation: %w", err)
	}
	if err := c.insertTransformation(ctx, fileID, key, transformedKey, p); err != nil {
		return "", err
	}

	logger.Ctx(ctx).Info("cdn: generated transformation",
		logger.String("file_uuid", fileUUID),
		logger.String("format", p.Format),
		logger.Any("width", p.Width),
		logger.Any("height", p.Height),
	)

	return transformedKey, nil
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
	if !c.validateSignedDelivery(w, r) {
		return
	}

	fileUUID := chi.URLParam(r, "id")
	if fileUUID == "" {
		http.Error(w, "missing image id", http.StatusBadRequest)
		return
	}

	parsed, err := parseImageParams(r.URL.Query(), r.Header.Get("Accept"), imageParamOptions{
		negotiateDefault: true,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if parsed.original {
		http.Redirect(w, r, c.fileURL(fileUUID), http.StatusTemporaryRedirect)
		return
	}
	p := parsed.params

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
	w.Header().Set("Content-Type", mimeForFormat(p.Format))
	w.Header().Set("Cache-Control", deliveryCacheControl(r))
	setImageVariantHeaders(w, p, parsed.varyAccept, "cdn-file-"+meta.FileUUID, parsed.surrogateKey)
	if etagMatches(r.Header.Get("If-None-Match"), etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	if r.Method == http.MethodHead {
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
	if size > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	}
	if _, err := io.Copy(w, reader); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to stream transformation", logger.Error(err))
	}
}
