// Package observability centralises the production telemetry wiring
// that main.go bootstraps. Right now it's just Sentry; metrics and
// tracing slot in here without widening main.go's import list.
package observability

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

// InitSentry boots sentry-go. A missing SENTRY_DSN is not a fatal
// condition — local development and preview builds run without it.
// Returns a `flush` function that MUST be deferred in main so pending
// events get shipped before the process exits; calling flush on a
// disabled hub is a no-op.
func InitSentry(ctx context.Context) (flush func()) {
	var dsn string
	if config.GlobalConfig != nil {
		dsn = config.GlobalConfig.SentryDSN
	} else {
		dsn = os.Getenv("SENTRY_DSN")
	}

	if dsn == "" {
		logger.Ctx(ctx).Info("sentry disabled: SENTRY_DSN not set")
		return func() {}
	}

	var env string
	if config.GlobalConfig != nil {
		env = config.GlobalConfig.AppEnv
	} else {
		env = os.Getenv("APP_ENV")
	}
	if env == "" {
		env = "production"
	}

	// Release string lets Sentry group events across deploys. The
	// build pipeline can stamp this with the git SHA; if absent, we
	// leave it empty and Sentry groups by version-in-event only.
	var release string
	if config.GlobalConfig != nil {
		release = config.GlobalConfig.AppRelease
	} else {
		release = os.Getenv("APP_RELEASE")
	}

	var hostname string
	if config.GlobalConfig != nil {
		hostname = config.GlobalConfig.Hostname
	} else {
		hostname = os.Getenv("HOSTNAME")
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      env,
		Release:          release,
		AttachStacktrace: true,
		// Tune these in production once we have a feel for volume.
		// 20 % traces is a safe starting sample rate — enough to
		// spot latency regressions, not enough to blow the quota.
		TracesSampleRate: 0.2,
		BeforeSend:       scrubRequestData,
		// Enrich events with the runtime host — helps when a single
		// region is misbehaving.
		ServerName: hostname,
	})
	if err != nil {
		logger.Ctx(ctx).Error("sentry init failed", logger.Error(err))
		return func() {}
	}

	logger.Ctx(ctx).Info("sentry initialised",
		logger.String("environment", env),
		logger.String("release", release),
	)

	return func() {
		// 2s is the default-ish flush window. Longer and we stall
		// shutdown; shorter and we drop events on a crashy deploy.
		sentry.Flush(2 * time.Second)
	}
}

// scrubRequestData removes credentials and customer content from request
// metadata before an event leaves the process. sentry.NewRequest already
// omits the query from URL, but QueryString is populated when the SDK is
// configured to capture request details, so clear both representations.
func scrubRequestData(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
	if event == nil || event.Request == nil {
		return event
	}
	event.Request.Data = ""
	event.Request.QueryString = ""
	event.Request.Cookies = ""
	if queryStart := strings.IndexByte(event.Request.URL, '?'); queryStart >= 0 {
		event.Request.URL = event.Request.URL[:queryStart]
	}
	return event
}

// RecoverMiddleware turns a panic in any handler into a Sentry event
// AND a 500 response. Without this, a panic would take down the
// request without ever reaching our error reporter — the worst
// possible combination (user sees crash, we never know).
//
// Chain this AFTER chi's middleware.Recoverer so the chi recoverer
// catches lower-level crashes, while this one captures everything
// that reaches the handler layer.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub := sentry.CurrentHub().Clone()
		hub.Scope().SetRequest(r)

		defer func() {
			if rec := recover(); rec != nil {
				hub.RecoverWithContext(r.Context(), rec)
				// Give the event a best-effort flush so a crash
				// doesn't beat our ship-to-Sentry window.
				hub.Flush(500 * time.Millisecond)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
