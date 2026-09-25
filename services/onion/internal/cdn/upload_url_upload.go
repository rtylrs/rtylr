package cdn

import (
	"context"
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

	"github.com/voxire/onion/internal/platform/logger"
)

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
