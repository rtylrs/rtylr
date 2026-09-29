package cdn

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

const (
	defaultSignedURLTTL = time.Hour
	maxSignedURLTTL     = 30 * 24 * time.Hour
)

type signDeliveryRequest struct {
	URL       string            `json:"url"`
	Path      string            `json:"path"`
	ID        string            `json:"id"`
	Params    map[string]string `json:"params"`
	ExpiresIn int64             `json:"expires_in"`
	ExpiresAt int64             `json:"expires_at"`
}

type signDeliveryResponse struct {
	URL       string `json:"url"`
	ExpiresAt int64  `json:"expires_at"`
	Signature string `json:"signature"`
}

func (c *CDN) signDelivery(w http.ResponseWriter, r *http.Request) {
	secret := c.signingSecret()
	if secret == "" {
		http.Error(w, "signing is not configured", http.StatusServiceUnavailable)
		return
	}
	scope, ok := callerScope(r)
	if !ok {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var req signDeliveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid sign request", http.StatusBadRequest)
		return
	}

	if err := c.authorizeSignTarget(r, scope, req); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	rawURL, err := c.signableURL(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	expiresAt := signedURLExpiry(req)
	u, err := c.signedURL(rawURL, expiresAt, secret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, r, signDeliveryResponse{
		URL:       u.String(),
		ExpiresAt: expiresAt,
		Signature: u.Query().Get("sig"),
	}, "no-store")
}

// authorizeSignTarget holds a sign request to the caller's own assets. A
// signature is a bearer credential for the file it names, so minting one for
// another tenant's uuid would be a cross-tenant read. Targets that name no
// stored file (a proxied remote image, for example) are not tenant data and
// pass through; a target that does name one is checked against the caller.
func (c *CDN) authorizeSignTarget(r *http.Request, scope mutationScope, req signDeliveryRequest) error {
	target := strings.TrimSpace(req.ID)
	if target == "" {
		raw := strings.TrimSpace(req.URL)
		if raw == "" {
			raw = strings.TrimSpace(req.Path)
		}
		if id, found := fileUUIDFromCanonicalURL(raw); found {
			target = id
		} else if id, found := storageKeyFileUUID(uploadURLStorageKeyFromTarget(raw)); found {
			target = id
		}
	}
	if target == "" {
		return nil
	}
	// A file in another tenant is reported exactly like one that does not
	// exist, so signing cannot be used to probe for ids.
	if _, err := c.fileForRead(r.Context(), scope.read(), target); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(r.Context()).Error("cdn: sign authorization failed",
				logger.String("file_uuid", target), logger.Error(err))
		}
		return errors.New("file not found")
	}
	return nil
}

// uploadURLStorageKeyFromTarget pulls a CDN storage key out of a `?url=`
// purge/sign target, so originals/<uuid> and transformed/<uuid>/ are subject
// to the same ownership check as /files/<uuid>.
func uploadURLStorageKeyFromTarget(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if u, err := url.Parse(raw); err == nil {
		if key := u.Query().Get("url"); key != "" {
			if inner, err := url.Parse(key); err == nil && inner.Path != "" {
				return strings.TrimPrefix(inner.Path, "/")
			}
		}
		if u.Path != "" {
			return strings.TrimPrefix(u.Path, "/")
		}
	}
	return strings.TrimPrefix(raw, "/")
}

func (c *CDN) signableURL(req signDeliveryRequest) (string, error) {
	params := url.Values{}
	for key, value := range req.Params {
		if strings.TrimSpace(value) != "" {
			params.Set(key, value)
		}
	}
	switch {
	case strings.TrimSpace(req.ID) != "":
		return c.uploadImageReadURL(strings.TrimSpace(req.ID), params), nil
	case strings.TrimSpace(req.URL) != "":
		return c.absoluteCDNURLWithParams(strings.TrimSpace(req.URL), params)
	case strings.TrimSpace(req.Path) != "":
		return c.absoluteCDNURLWithParams(strings.TrimSpace(req.Path), params)
	default:
		return "", fmt.Errorf("one of url, path, or id is required")
	}
}

func signedURLExpiry(req signDeliveryRequest) int64 {
	now := time.Now()
	if req.ExpiresAt > 0 {
		at := time.Unix(req.ExpiresAt, 0)
		if at.After(now.Add(maxSignedURLTTL)) {
			return now.Add(maxSignedURLTTL).Unix()
		}
		return req.ExpiresAt
	}
	ttl := time.Duration(req.ExpiresIn) * time.Second
	if ttl <= 0 {
		ttl = defaultSignedURLTTL
	}
	if ttl > maxSignedURLTTL {
		ttl = maxSignedURLTTL
	}
	return now.Add(ttl).Unix()
}

func (c *CDN) absoluteCDNURLWithParams(raw string, params url.Values) (string, error) {
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid cdn base url")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid url")
	}
	if !u.IsAbs() {
		u = base.ResolveReference(u)
	}
	if hostnameOnly(u.Host) != hostnameOnly(base.Host) {
		return "", fmt.Errorf("signed urls must use the cdn host")
	}
	if len(params) > 0 {
		query := u.Query()
		for key, values := range params {
			query.Del(key)
			for _, value := range values {
				query.Add(key, value)
			}
		}
		u.RawQuery = query.Encode()
	}
	return u.String(), nil
}

func (c *CDN) signedURL(raw string, expiresAt int64, secret string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid url")
	}
	q := u.Query()
	q.Set("expires", strconv.FormatInt(expiresAt, 10))
	q.Del("sig")
	q.Set("sig", signDeliveryPath(secret, u.Path, q))
	u.RawQuery = q.Encode()
	return u, nil
}

// validateSignedDelivery checks a delivery signature and returns the read scope
// it grants. An unsigned request keeps the scope the caller already had — the
// anonymous/public surface or their own tenant — so signing stays an optional
// cache and expiry control rather than the only way to reach a private asset.
func (c *CDN) validateSignedDelivery(w http.ResponseWriter, r *http.Request) (readScope, bool) {
	base := deliveryScopeFrom(r.Context())
	query := r.URL.Query()
	sig := strings.TrimSpace(query.Get("sig"))
	expires := strings.TrimSpace(query.Get("expires"))
	if sig == "" && expires == "" {
		return base, true
	}
	if sig == "" || expires == "" {
		http.Error(w, "invalid signed url", http.StatusForbidden)
		return base, false
	}
	secret := c.signingSecret()
	if secret == "" {
		http.Error(w, "signed delivery is not configured", http.StatusForbidden)
		return base, false
	}
	exp, err := strconv.ParseInt(expires, 10, 64)
	if err != nil || exp <= 0 {
		http.Error(w, "invalid signed url expiry", http.StatusForbidden)
		return base, false
	}
	if time.Now().Unix() > exp {
		http.Error(w, "signed url expired", http.StatusForbidden)
		return base, false
	}
	expected := signDeliveryPath(secret, r.URL.Path, query)
	if subtle.ConstantTimeCompare([]byte(sig), []byte(expected)) != 1 {
		http.Error(w, "invalid signed url", http.StatusForbidden)
		return base, false
	}
	return signedRead(), true
}

func deliveryCacheControl(r *http.Request) string {
	if strings.TrimSpace(r.URL.Query().Get("sig")) != "" || strings.TrimSpace(r.URL.Query().Get("expires")) != "" {
		return "private, no-store"
	}
	return immutableCacheControl
}

func (c *CDN) signingSecret() string {
	if c.apiKey != "" {
		return c.apiKey
	}
	return uploadJWTSecret()
}

func signDeliveryPath(secret, path string, query url.Values) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(path))
	mac.Write([]byte{'\n'})
	mac.Write([]byte(canonicalSignedQuery(query)))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func canonicalSignedQuery(query url.Values) string {
	keys := make([]string, 0, len(query))
	for key := range query {
		if key == "sig" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		values := append([]string(nil), query[key]...)
		sort.Strings(values)
		for _, value := range values {
			parts = append(parts, url.QueryEscape(key)+"="+url.QueryEscape(value))
		}
	}
	return strings.Join(parts, "&")
}
