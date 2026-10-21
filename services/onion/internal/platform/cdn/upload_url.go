package cdn

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
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
	scope, ok := c.validateSignedDelivery(w, r)
	if !ok {
		return
	}

	if id := strings.TrimSpace(r.URL.Query().Get("id")); id != "" {
		c.serveUploadID(w, r, scope, id)
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

func (c *CDN) serveUploadID(w http.ResponseWriter, r *http.Request, scope readScope, id string) {
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

	meta, err := c.fileForRead(r.Context(), scope, id)
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
