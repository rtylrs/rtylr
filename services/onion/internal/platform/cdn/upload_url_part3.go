package cdn

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/voxire/onion/internal/platform/logger"
)

type uploadURLImageSource struct {
	body      io.Reader
	mimeType  string
	size      int64
	cacheSeed string
}

type stagedUploadURLImage struct {
	inputPath  string
	outputPath string
	sourceHash string
	size       int64
	cleanup    func()
}

func (c *CDN) serveUploadURLWebP(w http.ResponseWriter, r *http.Request, source uploadURLImageSource) {
	p := Params{Chain: "f_webp", Format: "webp", Quality: uploadURLWebPQuality}
	p.Defaults()
	c.serveUploadURLTransform(w, r, source, imageParamResult{
		params:       p,
		explicit:     true,
		surrogateKey: imageVariantSurrogateKey(p),
	})
}

func (c *CDN) serveUploadURLTransform(w http.ResponseWriter, r *http.Request, source uploadURLImageSource, parsed imageParamResult) {
	p := parsed.params
	if r.Method == http.MethodHead {
		w.Header().Set("Content-Type", mimeForFormat(p.Format))
		w.Header().Set("Cache-Control", deliveryCacheControl(r))
		setImageVariantHeaders(w, p, parsed.varyAccept, uploadURLSurrogateKey(source.cacheSeed), parsed.surrogateKey)
		return
	}

	staged, err := stageUploadURLImage(source.body, c.maxUpload)
	if err != nil {
		if errors.Is(err, errUploadURLImageTooLarge) {
			http.Error(w, "image too large", http.StatusRequestEntityTooLarge)
			return
		}
		logger.Ctx(r.Context()).Warn("cdn: failed to stage upload-url image", logger.Error(err))
		http.Error(w, "image unavailable", http.StatusBadGateway)
		return
	}
	defer staged.cleanup()

	key := uploadURLTransformCacheKey(source.cacheSeed, staged.sourceHash, p)
	etag := quoteETag(key)
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", mimeForFormat(p.Format))
	w.Header().Set("Cache-Control", deliveryCacheControl(r))
	w.Header().Set("Voxire-CDN-Info", uploadURLInfoHeader(source, p))
	setImageVariantHeaders(w, p, parsed.varyAccept, uploadURLSurrogateKey(source.cacheSeed), parsed.surrogateKey)
	if etagMatches(r.Header.Get("If-None-Match"), etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if err := c.writeUploadURLTransform(w, r, source, staged, key, p); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: upload-url transform failed; falling back to original",
			logger.String("mime_type", source.mimeType),
			logger.String("format", p.Format),
			logger.Error(err))
		serveStagedUploadURLOriginal(w, r, staged, source)
	}
}

var errUploadURLImageTooLarge = errors.New("upload url image exceeds max size")

func stageUploadURLImage(body io.Reader, maxBytes int64) (*stagedUploadURLImage, error) {
	tmpIn, err := os.CreateTemp("", "cdn-upload-url-input-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp input: %w", err)
	}
	inputPath := tmpIn.Name()
	cleanupInput := true
	defer func() {
		if cleanupInput {
			_ = os.Remove(inputPath)
		}
	}()

	hash := sha256.New()
	n, err := copyUploadURLSource(io.MultiWriter(tmpIn, hash), body, maxBytes)
	closeErr := tmpIn.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, fmt.Errorf("closing temp input: %w", closeErr)
	}

	tmpOut, err := os.CreateTemp("", "cdn-upload-url-output-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp output: %w", err)
	}
	outputPath := tmpOut.Name()
	if err := tmpOut.Close(); err != nil {
		_ = os.Remove(outputPath)
		return nil, fmt.Errorf("closing temp output: %w", err)
	}

	cleanupInput = false
	return &stagedUploadURLImage{
		inputPath:  inputPath,
		outputPath: outputPath,
		sourceHash: hex.EncodeToString(hash.Sum(nil)),
		size:       n,
		cleanup: func() {
			_ = os.Remove(inputPath)
			_ = os.Remove(outputPath)
		},
	}, nil
}

func copyUploadURLSource(dst io.Writer, src io.Reader, maxBytes int64) (int64, error) {
	reader := src
	if maxBytes > 0 {
		reader = io.LimitReader(src, maxBytes+1)
	}
	n, err := io.Copy(dst, reader)
	if err != nil {
		return n, fmt.Errorf("copying source image: %w", err)
	}
	if maxBytes > 0 && n > maxBytes {
		return n, errUploadURLImageTooLarge
	}
	return n, nil
}

func uploadURLWebPCacheKey(seed, sourceHash string) string {
	p := Params{Chain: "f_webp", Format: "webp", Quality: uploadURLWebPQuality}
	p.Defaults()
	return uploadURLTransformCacheKey(seed, sourceHash, p)
}

func uploadURLTransformCacheKey(seed, sourceHash string, p Params) string {
	return cacheKey("upload-url-transform", seed, sourceHash, p.CacheKey("params"))
}

func uploadURLTransformSeedKey(seed string) string {
	return cacheKey("upload-url-seed", seed)
}

func uploadURLTransformStoragePrefix(seed string) string {
	return "upload-url-transformed/" + uploadURLTransformSeedKey(seed) + "/"
}

func uploadURLTransformStorageKey(seed, key string, p Params) string {
	return uploadURLTransformStoragePrefix(seed) + key + extensionForFormat(p.Format)
}

func uploadURLSurrogateKey(seed string) string {
	return "cdn-upload-url-" + uploadURLTransformSeedKey(seed)
}

func uploadURLInfoHeader(source uploadURLImageSource, p Params) string {
	return fmt.Sprintf("ifsz=%d ifmt=%s ofmt=%s", source.size, mimeTypeToFormat(source.mimeType), strings.ToLower(p.Format))
}

func (c *CDN) writeUploadURLTransform(w http.ResponseWriter, r *http.Request, source uploadURLImageSource, staged *stagedUploadURLImage, key string, p Params) error {
	transformedKey := uploadURLTransformStorageKey(source.cacheSeed, key, p)
	if c.storage != nil {
		if ok, _ := c.storage.Exists(r.Context(), transformedKey); ok {
			reader, size, err := c.storage.Download(r.Context(), transformedKey)
			if err != nil {
				return fmt.Errorf("downloading cached upload-url transform: %w", err)
			}
			defer func() { _ = reader.Close() }()
			if size > 0 {
				w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
			}
			if _, err := io.Copy(w, reader); err != nil {
				logger.Ctx(r.Context()).Warn("cdn: failed to stream cached upload-url transform", logger.Error(err))
			}
			return nil
		}
	}

	if err := c.generateUploadURLTransform(r.Context(), staged, p); err != nil {
		return err
	}

	resultFile, err := os.Open(staged.outputPath)
	if err != nil {
		return fmt.Errorf("opening upload-url transform: %w", err)
	}
	defer func() { _ = resultFile.Close() }()

	if info, statErr := resultFile.Stat(); statErr == nil && info.Size() > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	}

	if c.storage != nil {
		if err := c.storage.Upload(r.Context(), transformedKey, mimeForFormat(p.Format), resultFile); err != nil {
			logger.Ctx(r.Context()).Warn("cdn: failed to cache upload-url transform",
				logger.String("storage_key", transformedKey), logger.Error(err))
		}
		if _, err := resultFile.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("rewinding upload-url transform: %w", err)
		}
	}

	if _, err = io.Copy(w, resultFile); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to stream upload-url transform", logger.Error(err))
	}
	return nil
}
