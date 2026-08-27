package cdn

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/platform/logger"
)

const (
	maxUploadURLLength      = 4096
	uploadURLProxyTimeout   = 15 * time.Second
	uploadURLProxyRedirects = 3
	uploadURLCacheControl   = "public, max-age=86400, stale-while-revalidate=604800"
	uploadURLWebPQuality    = 85
)

type uploadURLSource struct {
	redirectURL string
	remoteURL   *url.URL
	storageKey  string
}

// serveUploadURL handles GET /upload?url=... as a compatibility read path for
// image URLs that predate /files/{uuid}. New uploads still use POST /upload.
func (c *CDN) serveUploadURL(w http.ResponseWriter, r *http.Request) {
	if !c.validateSignedDelivery(w, r) {
		return
	}

	if id := strings.TrimSpace(r.URL.Query().Get("id")); id != "" {
		c.serveUploadID(w, r, id)
		return
	}

	source, err := c.resolveUploadURLSource(r.URL.Query().Get("url"), r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if source.redirectURL != "" {
		c.redirectToCanonical(w, source.redirectURL)
		return
	}
	if source.storageKey != "" {
		c.serveStorageKeyImage(w, r, source.storageKey)
		return
	}
	if source.remoteURL != nil {
		c.proxyRemoteImage(w, r, source.remoteURL)
		return
	}

	http.Error(w, "invalid image url", http.StatusBadRequest)
}

func (c *CDN) serveUploadID(w http.ResponseWriter, r *http.Request, id string) {
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "invalid file id", http.StatusBadRequest)
		return
	}

	parsed, err := parseImageParams(r.URL.Query(), r.Header.Get("Accept"), imageParamOptions{
		defaultFullImage: true,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if parsed.original {
		c.redirectToCanonical(w, c.fileURL(id))
		return
	}

	meta, err := c.fileByUUID(r.Context(), id)
	if err != nil {
		c.redirectToCanonical(w, c.fileURL(id))
		return
	}
	if !isUploadURLTransformableMIME(meta.MIMEType) {
		c.redirectToCanonical(w, c.fileURL(id))
		return
	}
	c.redirectToCanonical(w, c.uploadImageReadURL(id, r.URL.Query()))
}

func (c *CDN) redirectToCanonical(w http.ResponseWriter, destination string) {
	base, baseErr := url.Parse(c.baseURL)
	target, targetErr := url.Parse(destination)
	if baseErr != nil || targetErr != nil || base.Scheme == "" || base.Host == "" ||
		target.Scheme != base.Scheme || target.Host != base.Host || target.User != nil {
		http.Error(w, "invalid canonical image url", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", target.String())
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (c *CDN) resolveUploadURLSource(raw, requestHost string) (uploadURLSource, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return uploadURLSource{}, errors.New("missing url")
	}
	if len(raw) > maxUploadURLLength {
		return uploadURLSource{}, errors.New("url is too long")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return uploadURLSource{}, errors.New("invalid url")
	}
	if u.Scheme == "" && u.Host != "" {
		return uploadURLSource{}, errors.New("protocol-relative urls are not supported")
	}

	if u.Scheme == "" && u.Host == "" {
		return c.resolveUploadPathSource(u)
	}

	host := hostnameOnly(u.Host)
	if c.isCDNHost(host, requestHost) {
		return c.resolveUploadCDNSource(u)
	}

	if err := validateRemoteImageURL(u); err != nil {
		return uploadURLSource{}, err
	}
	return uploadURLSource{remoteURL: u}, nil
}

func (c *CDN) resolveUploadPathSource(u *url.URL) (uploadURLSource, error) {
	if id := strings.TrimPrefix(u.Path, "/"); !strings.Contains(id, "/") {
		if _, err := uuid.Parse(id); err == nil {
			if uploadURLWantsOriginal(u.Query()) {
				return uploadURLSource{redirectURL: c.fileURL(id)}, nil
			}
			return uploadURLSource{redirectURL: c.uploadImageReadURL(id, u.Query())}, nil
		}
	}
	if isCanonicalReadPath(u.Path) {
		return uploadURLSource{redirectURL: c.canonicalReadURL(u)}, nil
	}
	if strings.HasPrefix(u.Path, "/upload") {
		return uploadURLSource{}, errors.New("nested upload urls are not supported")
	}

	key, err := storageKeyFromPath(u.Path)
	if err != nil {
		return uploadURLSource{}, err
	}
	return uploadURLSource{storageKey: key}, nil
}

func (c *CDN) resolveUploadCDNSource(u *url.URL) (uploadURLSource, error) {
	if isCanonicalReadPath(u.Path) {
		return uploadURLSource{redirectURL: c.canonicalReadURL(u)}, nil
	}
	if strings.HasPrefix(u.Path, "/upload") {
		return uploadURLSource{}, errors.New("nested upload urls are not supported")
	}

	key, err := storageKeyFromPath(u.Path)
	if err != nil {
		return uploadURLSource{}, err
	}
	return uploadURLSource{storageKey: key}, nil
}

func isCanonicalReadPath(p string) bool {
	return strings.HasPrefix(p, "/files/") ||
		strings.HasPrefix(p, "/images/") ||
		strings.HasPrefix(p, "files/") ||
		strings.HasPrefix(p, "images/")
}

func (c *CDN) canonicalReadURL(u *url.URL) string {
	if id, ok := uploadFileUUIDFromPath(u.Path); ok {
		if uploadURLWantsOriginal(u.Query()) {
			return c.fileURL(id)
		}
		return c.uploadImageReadURL(id, u.Query())
	}

	uri := u.RequestURI()
	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}
	return c.baseURL + uri
}

func uploadFileUUIDFromPath(rawPath string) (string, bool) {
	p := strings.TrimPrefix(rawPath, "/")
	if !strings.HasPrefix(p, "files/") {
		return "", false
	}
	id := strings.TrimPrefix(p, "files/")
	if id == "" || strings.Contains(id, "/") {
		return "", false
	}
	if _, err := uuid.Parse(id); err != nil {
		return "", false
	}
	return id, true
}

func (c *CDN) uploadImageReadURL(fileUUID string, query url.Values) string {
	out := imageTransformQuery(query)
	if !hasImageTransformQuery(out) {
		out.Set("t", "f_webp")
	} else if out.Get("format") == "" && out.Get("f") == "" && out.Get("t") == "" && out.Get("auto") == "" {
		out.Set("format", "webp")
	}
	return fmt.Sprintf("%s/images/%s?%s", c.baseURL, fileUUID, out.Encode())
}

func imageTransformQuery(query url.Values) url.Values {
	out := url.Values{}
	for key, values := range query {
		if !imageTransformQueryKeys[strings.ToLower(key)] {
			continue
		}
		for _, value := range values {
			if strings.TrimSpace(value) != "" {
				out.Add(key, value)
			}
		}
	}
	return out
}

func storageKeyFromPath(rawPath string) (string, error) {
	key := strings.TrimPrefix(rawPath, "/")
	key = strings.TrimSpace(key)
	if key == "" {
		return "", errors.New("missing storage key")
	}
	if strings.ContainsAny(key, "\\\x00\r\n") {
		return "", errors.New("invalid storage key")
	}
	for _, part := range strings.Split(key, "/") {
		if part == "" || part == "." || part == ".." {
			return "", errors.New("invalid storage key")
		}
	}
	return key, nil
}

func (c *CDN) isCDNHost(host, requestHost string) bool {
	host = strings.ToLower(host)
	if host == "" {
		return false
	}
	if host == "cdn.voxire.com" || host == "cdn.voxire.dev" {
		return true
	}
	if base, err := url.Parse(c.baseURL); err == nil && hostnameOnly(base.Host) == host {
		return true
	}
	return requestHost != "" && hostnameOnly(requestHost) == host
}

func validateRemoteImageURL(u *url.URL) error {
	if u == nil {
		return errors.New("invalid url")
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return errors.New("only http and https urls are supported")
	}
	if u.Host == "" {
		return errors.New("missing url host")
	}
	if u.User != nil {
		return errors.New("url credentials are not supported")
	}
	if strings.ContainsAny(u.Host, " \t\r\n") {
		return errors.New("invalid url host")
	}
	if ip := net.ParseIP(hostnameOnly(u.Host)); ip != nil && !isPublicProxyIP(ip) {
		return errors.New("url host is not public")
	}
	return nil
}

func (c *CDN) serveStorageKeyImage(w http.ResponseWriter, r *http.Request, key string) {
	mimeType := imageMIMEForStorageKey(key)
	if mimeType == "" {
		http.Error(w, "unsupported image type", http.StatusUnsupportedMediaType)
		return
	}

	parsed, err := parseImageParams(r.URL.Query(), r.Header.Get("Accept"), imageParamOptions{
		defaultFullImage: true,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if uploadURLShouldTransform(parsed, mimeType) && r.Method == http.MethodHead {
		w.Header().Set("Content-Type", mimeForFormat(parsed.params.Format))
		w.Header().Set("Cache-Control", deliveryCacheControl(r))
		setImageVariantHeaders(w, parsed.params, parsed.varyAccept, uploadURLSurrogateKey("storage:"+key), parsed.surrogateKey)
		return
	}

	etag := quoteETag(cacheKey("upload-url-storage", key))
	if !uploadURLShouldTransform(parsed, mimeType) {
		w.Header().Set("ETag", etag)
		w.Header().Set("Content-Type", mimeType)
		w.Header().Set("Cache-Control", deliveryCacheControl(r))
		if etagMatches(r.Header.Get("If-None-Match"), etag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		if r.Method == http.MethodHead {
			return
		}
	}

	reader, size, err := c.storage.Download(r.Context(), key)
	if err != nil {
		logger.Ctx(r.Context()).Warn("cdn: storage-key image lookup failed",
			logger.String("storage_key", key), logger.Error(err))
		http.Error(w, "image not found", http.StatusNotFound)
		return
	}
	defer func() { _ = reader.Close() }()

	if uploadURLShouldTransform(parsed, mimeType) {
		c.serveUploadURLTransform(w, r, uploadURLImageSource{
			body:      reader,
			mimeType:  mimeType,
			size:      size,
			cacheSeed: "storage:" + key,
		}, parsed)
		return
	}

	if size > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	}
	if _, err := io.Copy(w, reader); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to stream storage-key image", logger.Error(err))
	}
}

func (c *CDN) proxyRemoteImage(w http.ResponseWriter, r *http.Request, target *url.URL) {
	parsed, err := parseImageParams(r.URL.Query(), r.Header.Get("Accept"), imageParamOptions{
		defaultFullImage: true,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client := imageProxyHTTPClient()
	ctx, cancel := context.WithTimeout(r.Context(), uploadURLProxyTimeout)
	defer cancel()

	method := http.MethodGet
	if r.Method == http.MethodHead {
		method = http.MethodHead
	}
	req, err := http.NewRequestWithContext(ctx, method, target.String(), nil)
	if err != nil {
		http.Error(w, "invalid image url", http.StatusBadRequest)
		return
	}
	req.Header.Set("Accept", "image/avif,image/webp,image/*,*/*;q=0.8")
	if parsed.original {
		if v := r.Header.Get("If-None-Match"); v != "" {
			req.Header.Set("If-None-Match", v)
		}
		if v := r.Header.Get("If-Modified-Since"); v != "" {
			req.Header.Set("If-Modified-Since", v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Ctx(r.Context()).Warn("cdn: remote image fetch failed",
			logger.String("url", target.Redacted()), logger.Error(err))
		http.Error(w, "remote image unavailable", http.StatusBadGateway)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotModified {
		copyProxyCacheHeaders(w.Header(), resp.Header)
		w.WriteHeader(http.StatusNotModified)
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		http.Error(w, "remote image unavailable", http.StatusBadGateway)
		return
	}
	if c.maxUpload > 0 && resp.ContentLength > c.maxUpload {
		http.Error(w, "remote image too large", http.StatusRequestEntityTooLarge)
		return
	}

	var head []byte
	body := io.Reader(resp.Body)
	if method != http.MethodHead {
		head, err = readProxyHead(resp.Body)
		if err != nil {
			http.Error(w, "remote image unavailable", http.StatusBadGateway)
			return
		}
		body = io.MultiReader(bytes.NewReader(head), resp.Body)
	}

	mimeType := responseImageMIME(resp.Header.Get("Content-Type"), head)
	if mimeType == "" {
		http.Error(w, "remote url is not a supported image", http.StatusUnsupportedMediaType)
		return
	}

	if uploadURLShouldTransform(parsed, mimeType) {
		c.serveUploadURLTransform(w, r, uploadURLImageSource{
			body:      body,
			mimeType:  mimeType,
			size:      resp.ContentLength,
			cacheSeed: "remote:" + target.String(),
		}, parsed)
		return
	}

	copyProxyCacheHeaders(w.Header(), resp.Header)
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", deliveryCacheControl(r))
	if resp.ContentLength > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", resp.ContentLength))
	}
	if method == http.MethodHead {
		return
	}

	limited := body
	if c.maxUpload > 0 {
		limited = io.LimitReader(body, c.maxUpload)
	}
	if _, err := io.Copy(w, limited); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to stream remote image", logger.Error(err))
	}
}

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

func (c *CDN) generateUploadURLTransform(ctx context.Context, staged *stagedUploadURLImage, p Params) error {
	if c.mediaEngine == "" {
		return errors.New("media-engine path is not configured")
	}
	if c.engineSem == nil {
		return errors.New("media-engine semaphore is not configured")
	}

	p.Defaults()
	if err := p.Validate(); err != nil {
		return err
	}

	select {
	case c.engineSem <- struct{}{}:
		defer func() { <-c.engineSem }()
	case <-ctx.Done():
		return ctx.Err()
	}

	execCtx, cancel := context.WithTimeout(ctx, engineTimeout)
	defer cancel()
	// #nosec G204 G702 -- mediaEngine is a server-side configured binary path; CLI args are built from validated params and temp paths.
	// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
	cmd := exec.CommandContext(execCtx, c.mediaEngine, p.CLIArgs(staged.inputPath, staged.outputPath)...)
	if out, runErr := cmd.CombinedOutput(); runErr != nil {
		return fmt.Errorf("media-engine failed: %w\noutput: %s", runErr, out)
	}
	return nil
}

func serveStagedUploadURLOriginal(w http.ResponseWriter, r *http.Request, staged *stagedUploadURLImage, source uploadURLImageSource) {
	etag := quoteETag(cacheKey("upload-url-original", source.cacheSeed, staged.sourceHash))
	w.Header().Del("Content-Length")
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", source.mimeType)
	w.Header().Set("Cache-Control", deliveryCacheControl(r))
	if etagMatches(r.Header.Get("If-None-Match"), etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if staged.size > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", staged.size))
	}
	input, err := os.Open(staged.inputPath)
	if err != nil {
		http.Error(w, "image unavailable", http.StatusBadGateway)
		return
	}
	defer func() { _ = input.Close() }()
	if _, err := io.Copy(w, input); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to stream fallback upload-url image", logger.Error(err))
	}
}

func uploadURLWantsOriginal(query url.Values) bool {
	format := strings.ToLower(strings.TrimSpace(query.Get("format")))
	if format == "" {
		format = strings.ToLower(strings.TrimSpace(query.Get("f")))
	}
	return format == "original" || format == "source"
}

func uploadURLShouldTransform(parsed imageParamResult, mimeType string) bool {
	if parsed.original || !isUploadURLTransformableMIME(mimeType) {
		return false
	}
	if parsed.explicit {
		return true
	}
	return strings.ToLower(parsed.params.Format) != mimeTypeToFormat(strings.ToLower(strings.TrimSpace(mimeType)))
}

func isUploadURLTransformableMIME(mimeType string) bool {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/avif", "image/gif", "image/heic", "image/heif", "image/jpeg", "image/jxl", "image/png", "image/tiff", "image/webp":
		return true
	default:
		return false
	}
}

func imageProxyHTTPClient() *http.Client {
	transport := &http.Transport{
		DialContext:           safeImageProxyDialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   uploadURLProxyTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= uploadURLProxyRedirects {
				return errors.New("too many redirects")
			}
			return validateRemoteImageURL(req.URL)
		},
	}
}

func safeImageProxyDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	if ip := net.ParseIP(host); ip != nil {
		if !isPublicProxyIP(ip) {
			return nil, fmt.Errorf("refusing private proxy address %s", ip.String())
		}
		return dialer.DialContext(ctx, network, address)
	}

	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		if isPublicProxyIP(addr.IP) {
			return dialer.DialContext(ctx, network, net.JoinHostPort(addr.IP.String(), port))
		}
	}
	return nil, fmt.Errorf("no public addresses for %s", host)
}

func isPublicProxyIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	return ip.IsGlobalUnicast() &&
		!ip.IsPrivate() &&
		!ip.IsLoopback() &&
		!ip.IsLinkLocalUnicast() &&
		!ip.IsLinkLocalMulticast() &&
		!ip.IsUnspecified()
}

func imageMIMEForStorageKey(key string) string {
	switch strings.ToLower(path.Ext(key)) {
	case ".avif":
		return "image/avif"
	case ".gif":
		return "image/gif"
	case ".heic":
		return "image/heic"
	case ".heif":
		return "image/heif"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".jxl":
		return "image/jxl"
	case ".png":
		return "image/png"
	case ".tif", ".tiff":
		return "image/tiff"
	case ".webp":
		return "image/webp"
	default:
		return ""
	}
}

func responseImageMIME(header string, head []byte) string {
	mimeType := ""
	if header != "" {
		if parsed, _, err := mime.ParseMediaType(header); err == nil {
			mimeType = strings.ToLower(parsed)
		}
	}
	if mimeType == "" && len(head) > 0 {
		mimeType = strings.ToLower(http.DetectContentType(head))
	}
	switch mimeType {
	case "image/avif", "image/gif", "image/heic", "image/heif", "image/jpeg", "image/jxl", "image/png", "image/tiff", "image/webp":
		return mimeType
	default:
		return ""
	}
}

func readProxyHead(body io.Reader) ([]byte, error) {
	buf := make([]byte, 512)
	n, err := io.ReadFull(body, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return buf[:n], nil
}

func copyProxyCacheHeaders(dst, src http.Header) {
	for _, key := range []string{"ETag", "Last-Modified"} {
		if value := src.Get(key); value != "" {
			dst.Set(key, value)
		}
	}
}

func hostnameOnly(hostport string) string {
	host := strings.TrimSpace(strings.ToLower(hostport))
	if host == "" {
		return ""
	}
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		return strings.Trim(parsed, "[]")
	}
	return strings.Trim(host, "[]")
}
