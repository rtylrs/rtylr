// Package cdn is the media/CDN subsystem powering the `upload` runmode
// (cmd/upload). It owns file uploads, on-demand image transformation, and
// asset delivery — backed by the shared S3 storage client
// (internal/platform/storage) and the cdn_files / cdn_transformations tables.
//
// It is deliberately a self-contained subsystem rather than a tenant API
// module: it serves cross-product asset traffic on its own binary/port and
// authenticates writes with a service API key, not per-user JWTs. SQL is
// isolated in store.go; handlers call service methods, service methods call
// the store — the same inward dependency direction the rest of onion follows.
package cdn

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/platform/storage"
)

// CDN bundles the dependencies shared by every CDN handler. One instance is
// built at boot in cmd/upload and its routes mounted onto the chi router.
type CDN struct {
	db          *sql.DB
	storage     *storage.Client
	baseURL     string
	maxUpload   int64
	mediaEngine string
	apiKey      string
}

// immutableCacheControl is sent on every served asset. Originals are keyed by
// UUID and transformations by a content hash, so a given URL's bytes never
// change — clients and edge caches may keep them forever.
const immutableCacheControl = "public, max-age=31536000, immutable"

// New constructs a CDN. baseURL is the public prefix for generated file URLs,
// maxUpload the per-request byte ceiling, mediaEngine the path to the C++
// media-engine binary, and apiKey the bearer token guarding write endpoints
// (empty disables write auth — local dev only).
func New(db *sql.DB, store *storage.Client, baseURL string, maxUpload int64, mediaEngine, apiKey string) *CDN {
	return &CDN{
		db:          db,
		storage:     store,
		baseURL:     baseURL,
		maxUpload:   maxUpload,
		mediaEngine: mediaEngine,
		apiKey:      apiKey,
	}
}

// RegisterRoutes mounts the CDN endpoints. Read endpoints are public; the
// upload endpoint is wrapped in the API-key middleware. /health, /metrics and
// CORS are provided by the shared web server, so they are not registered here.
func (c *CDN) RegisterRoutes(r chi.Router) {
	r.Get("/files/{id}", c.serveFile)
	r.Get("/images/{id}", c.serveTransformed)
	r.Get("/videos/info", c.videoInfo)

	// HEAD reuses the GET handlers — net/http discards the body for HEAD, so
	// callers get Content-Type/ETag/Cache-Control without transferring bytes.
	r.Head("/files/{id}", c.serveFile)
	r.Head("/images/{id}", c.serveTransformed)

	r.Group(func(gr chi.Router) {
		gr.Use(c.apiKeyAuth)
		gr.Post("/upload", c.upload)
	})
}
