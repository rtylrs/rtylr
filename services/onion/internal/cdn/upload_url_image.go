package cdn

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/voxire/onion/internal/platform/logger"
)

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
	if host == "cdn.voxire.com" || host == "cdn.voxire.dev" || host == "api.voxire.com" || host == "api.voxire.dev" {
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
	if c.storage == nil {
		http.NotFound(w, r)
		return
	}
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
