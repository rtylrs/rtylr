// Package cdn is the media/CDN subsystem powering the `upload` runmode
// (cmd/upload). It owns file uploads, on-demand image transformation, and
// asset delivery — backed by the shared S3 storage client
// (internal/platform/storage) and the cdn_files / cdn_transformations tables.
//
// It is deliberately a self-contained subsystem rather than a tenant API
// module: it serves cross-product asset traffic on its own binary/port.
// Writes are authenticated with a service API key or a per-user credential,
// and every stored asset is owned by the business resolved for the caller
// (tenant.ResolveBusinessID). SQL is isolated in store.go; handlers call
// service methods, service methods call the store — the same inward dependency
// direction the rest of onion follows.
package cdn

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"runtime"

	"github.com/go-chi/chi/v5"
	"golang.org/x/sync/singleflight"

	"github.com/voxire/onion/internal/platform/storage"
)

// CDN bundles the dependencies shared by every CDN handler. One instance is
// built at boot in cmd/upload and its routes mounted onto the chi router.
type CDN struct {
	uploadKeyScope func(context.Context, string) (int64, int64, error)
	db             *sql.DB
	storage        *storage.Client
	baseURL        string
	maxUpload      int64
	mediaEngine    string
	apiKey         string

	// moderator, when non-nil, screens image uploads for nudity before they
	// are stored. moderationFailClosed rejects uploads when the check errors.
	moderator            ImageModerator
	moderationFailClosed bool

	// engineSem bounds how many media-engine subprocesses run at once. Each
	// invocation is CPU- and memory-heavy (libvips decode + encode), so a flood
	// of cache-miss requests must not fork an unbounded number of them.
	engineSem chan struct{}

	// engine keeps up to engineSem-capacity resident `media-engine --serve`
	// processes warm, avoiding per-request fork/exec + VIPS_INIT cost, and
	// falls back to single-shot exec when resident mode is unavailable.
	engine *enginePool

	// transformGroup collapses concurrent cache misses for the same derivative
	// inside one process. Only the first request runs media-engine; waiters
	// stream the stored object once it exists.
	transformGroup singleflight.Group
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
	concurrency := runtime.NumCPU()
	if concurrency < 1 {
		concurrency = 1
	}
	return &CDN{
		db:          db,
		storage:     store,
		baseURL:     baseURL,
		maxUpload:   maxUpload,
		mediaEngine: mediaEngine,
		apiKey:      apiKey,
		engineSem:   make(chan struct{}, concurrency),
		engine:      newEnginePool(mediaEngine, concurrency),
	}
}

// SetModerator enables image content moderation on stored uploads. failClosed
// controls behavior when the moderation check itself errors: true rejects the
// upload, false allows it with a warning log.
func (c *CDN) SetModerator(m ImageModerator, failClosed bool) {
	c.moderator = m
	c.moderationFailClosed = failClosed
}

// CheckMediaEngine verifies the configured media-engine binary exists and is
// executable. Image transformation (/images) depends on it; uploads and
// original delivery do not, so callers should treat a failure as a warning
// rather than fatal. Returns nil when the engine is usable.
func (c *CDN) CheckMediaEngine() error {
	if c.mediaEngine == "" {
		return fmt.Errorf("media-engine path is not configured")
	}
	info, err := os.Stat(c.mediaEngine)
	if err != nil {
		return fmt.Errorf("media-engine not found at %q: %w", c.mediaEngine, err)
	}
	if info.IsDir() || info.Mode().Perm()&0o111 == 0 {
		return fmt.Errorf("media-engine at %q is not an executable file", c.mediaEngine)
	}
	return nil
}

// PURGE is a non-standard HTTP method (used here for cache/derivative
// invalidation). chi panics if you route a method it doesn't know about, so it
// must be registered once at package load — before any RegisterRoutes call —
// otherwise the upload runmode crash-loops on boot.
func registerPurgeMethod() struct{} {
	chi.RegisterMethod("PURGE")
	return struct{}{}
}

var _ = registerPurgeMethod()

// RegisterRoutes mounts the CDN endpoints. Read endpoints are mounted behind
// deliveryAuth, which resolves the caller's tenant when a bearer token is
// present and leaves the request anonymous otherwise; anonymous callers only
// see assets flagged public, and a validated delivery signature is the
// credential for the rest. /health, /metrics and CORS are provided by the
// shared web server, so they are not registered here.
func (c *CDN) RegisterRoutes(r chi.Router) {
	r.Get("/capabilities", c.capabilities)

	r.Group(func(dr chi.Router) {
		dr.Use(c.deliveryAuth)
		dr.Get("/upload", c.serveUploadURL)
		dr.Get("/upload/plan", c.uploadPlan)
		dr.Get("/upload/responsive", c.uploadResponsive)
		dr.Get("/upload/variants", c.uploadVariants)
		dr.Get("/files/{id}", c.serveFile)
		dr.Get("/images/{id}/analyze", c.imageAnalyze)
		dr.Get("/images/{id}/info", c.imageInfo)
		dr.Get("/images/{id}/plan", c.imagePlan)
		dr.Get("/images/{id}/responsive", c.imageResponsive)
		dr.Get("/images/{id}/variants", c.imageVariants)
		dr.Get("/images/{id}", c.serveTransformed)
		dr.Get("/videos/info", c.videoInfo)

		// HEAD reuses the GET handlers — net/http discards the body for HEAD, so
		// callers get Content-Type/ETag/Cache-Control without transferring bytes.
		dr.Head("/upload", c.serveUploadURL)
		dr.Head("/files/{id}", c.serveFile)
		dr.Head("/images/{id}", c.serveTransformed)
	})

	r.Group(func(gr chi.Router) {
		gr.Use(c.apiKeyAuth)
		gr.Post("/upload", c.upload)
		gr.Post("/purge", c.purgeControl)
		gr.Post("/sign", c.signDelivery)
		gr.MethodFunc("PURGE", "/upload", c.purgeUploadURL)
		gr.MethodFunc("PURGE", "/files/{id}", c.purgeFileVariants)
		gr.MethodFunc("PURGE", "/images/{id}", c.purgeFileVariants)
	})
}

// WithUploadKeyReader grants only credential validation, without Organization persistence access.
func (c *CDN) WithUploadKeyReader(read func(context.Context, string) (int64, int64, error)) *CDN {
	c.uploadKeyScope = read
	return c
}
