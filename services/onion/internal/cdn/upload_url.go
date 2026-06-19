package cdn

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
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
)

type uploadURLSource struct {
	redirectURL string
	remoteURL   *url.URL
	storageKey  string
}

// serveUploadURL handles GET /upload?url=... as a compatibility read path for
// image URLs that predate /files/{uuid}. New uploads still use POST /upload.
func (c *CDN) serveUploadURL(w http.ResponseWriter, r *http.Request) {
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
		http.Redirect(w, r, source.redirectURL, http.StatusTemporaryRedirect)
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
	http.Redirect(w, r, c.fileURL(id), http.StatusTemporaryRedirect)
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
			return uploadURLSource{redirectURL: c.fileURL(id)}, nil
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
	uri := u.RequestURI()
	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}
	return c.baseURL + uri
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

	etag := quoteETag(cacheKey("upload-url-storage", key))
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", uploadURLCacheControl)
	if etagMatches(r.Header.Get("If-None-Match"), etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	if r.Method == http.MethodHead {
		return
	}

	reader, size, err := c.storage.Download(r.Context(), key)
	if err != nil {
		logger.Ctx(r.Context()).Warn("cdn: storage-key image lookup failed",
			logger.String("storage_key", key), logger.Error(err))
		http.Error(w, "image not found", http.StatusNotFound)
		return
	}
	defer func() { _ = reader.Close() }()

	if size > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	}
	if _, err := io.Copy(w, reader); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to stream storage-key image", logger.Error(err))
	}
}

func (c *CDN) proxyRemoteImage(w http.ResponseWriter, r *http.Request, target *url.URL) {
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
	if v := r.Header.Get("If-None-Match"); v != "" {
		req.Header.Set("If-None-Match", v)
	}
	if v := r.Header.Get("If-Modified-Since"); v != "" {
		req.Header.Set("If-Modified-Since", v)
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

	copyProxyCacheHeaders(w.Header(), resp.Header)
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", uploadURLCacheControl)
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
	case ".jpg", ".jpeg":
		return "image/jpeg"
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
	case "image/avif", "image/gif", "image/jpeg", "image/png", "image/tiff", "image/webp":
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
