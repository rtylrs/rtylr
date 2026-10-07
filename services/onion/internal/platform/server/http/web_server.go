package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/observability"
	"github.com/voxire/onion/internal/platform/server/http/httpmetrics"
)

const serverShutdownTimeoutSeconds = 20

func Start(ctx context.Context, runMode ChiRunMode, port string) error {
	// Rate limiters created during package initialization (notably the auth
	// route limiters) are process-lifetime objects. Snapshot them before setup
	// so a server restart only tears down limiters owned by this Start call.
	preexistingRateLimiters := snapshotSecurityRateLimiters()
	defer stopSecurityRateLimitersNotIn(preexistingRateLimiters)

	r := chi.NewRouter()

	runMode.Setup(ctx, r)

	// /metrics is mounted before runMode.Setup's module routes run so
	// there's no way a module accidentally shadows it. Basic-auth is
	// enforced inside the handler; the endpoint 404s when
	// METRICS_USER/METRICS_PASS aren't set, so it stays opt-in.
	r.Get("/metrics", httpmetrics.Handler())

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if usesDatabaseHealthCheck(runMode) {
			pingCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if db := dbclient.Pool(); db == nil || db.PingContext(pingCtx) != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				if _, err := w.Write([]byte(`{"status":"unhealthy"}`)); err != nil {
					logger.Ctx(r.Context()).Warn("Failed to write health response", logger.Error(err))
				}
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			logger.Ctx(ctx).Warn("Failed to write health response", logger.Error(err))
		}
	})

	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(`{"message":"Why are you here?","status":"ok"}`)); err != nil {
			logger.Ctx(ctx).Warn("Failed to write response", logger.Error(err))
		}
	})

	r.Get("/favicon.ico", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	middlewares := platformMiddlewares()
	if provider, ok := runMode.(MiddlewareProvider); ok {
		middlewares = append(middlewares, provider.Middlewares()...)
	}

	handler := composeHandler(r, middlewares)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	listenErr := make(chan error, 1)

	go func() {
		listenErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-listenErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeoutSeconds*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Ctx(ctx).Error("Failed to shutdown server gracefully",
			logger.Error(err),
		)

		return err
	}

	if err := <-listenErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// snapshotSecurityRateLimiters records limiters that already exist before a
// server instance creates its route-specific limiters. The package-global
// anonymous auth limiters are created during package initialization and must
// remain usable when a listener is restarted.
func snapshotSecurityRateLimiters() map[*SecurityRateLimiter]struct{} {
	rateLimiterMu.RLock()
	defer rateLimiterMu.RUnlock()

	limiters := make(map[*SecurityRateLimiter]struct{}, len(activeRateLimiters))
	for limiter := range activeRateLimiters {
		limiters[limiter] = struct{}{}
	}
	return limiters
}

// stopSecurityRateLimitersNotIn stops only the limiters created by the current
// server instance. It deliberately leaves preexisting/process-lifetime
// limiters running, which is required for a clean restart of the server.
func stopSecurityRateLimitersNotIn(keep map[*SecurityRateLimiter]struct{}) {
	rateLimiterMu.Lock()
	toStop := make([]*SecurityRateLimiter, 0)
	for limiter := range activeRateLimiters {
		if _, ok := keep[limiter]; ok {
			continue
		}
		delete(activeRateLimiters, limiter)
		toStop = append(toStop, limiter)
	}
	rateLimiterMu.Unlock()

	for _, limiter := range toStop {
		limiter.Stop()
	}
}

func usesDatabaseHealthCheck(runMode ChiRunMode) bool {
	mode, ok := runMode.(DatabaseHealthMode)
	if !ok {
		return true
	}
	return mode.UsesDatabaseHealthCheck()
}

// composeHandler wraps the bare router with the given middlewares (outermost
// first) instead of registering them with r.Use. Keeping the chain outside the
// router is what lets the Partials X-API-KEY bypass send a request straight to
// the router with no middleware in front of it.
func composeHandler(r *chi.Mux, middlewares []MiddlewareFunc) http.Handler {
	var protected http.Handler = r
	for i := len(middlewares) - 1; i >= 0; i-- {
		protected = middlewares[i](protected)
	}
	next := protected
	protected = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := context.WithValue(req.Context(), observability.RouteLookupContextKey{}, r)
		next.ServeHTTP(w, req.WithContext(ctx))
	})

	if partialsAPIKey() == "" {
		return protected
	}

	return PartialsBypassHandler(r, protected)
}

// platformMiddlewares returns the middlewares every runmode gets, outermost
// first. They are returned as a slice (rather than applied with r.Use) so the
// server can wrap the router with them and keep the Partials bypass outside the
// whole chain.
func platformMiddlewares() []MiddlewareFunc {
	middlewares := make([]MiddlewareFunc, 0, 16)

	if accessLogsEnabled() {
		middlewares = append(middlewares, safeRequestLogger())
	}

	middlewares = append(middlewares,
		middleware.Recoverer,
		// Sentry recover runs after chi's Recoverer so we still respond
		// with a 500, but anything that escapes the handler layer ends
		// up in Sentry with a proper stack trace and request scope.
		observability.RecoverMiddleware,
		corsMiddleware(),
		CrashNotifierMiddleware(),
		SecurityHeadersMiddleware(),
		SecureErrorMiddleware(),
		responseCompressMiddleware(),
		RequestBodyLimitMiddleware(DefaultMaxRequestBodyBytes),
		// Global per-IP guardrail against scraping/abuse only — set very high so
		// it never throttles real use (a SPA session bursts many calls/sec).
		// Per-route limits (login/register/AI) stay tight via PerRouteRateLimit.
		RateLimitMiddleware(1000, time.Second),
		// Tight per-IP budgets on the anonymous auth routes (login, register,
		// password reset, invite acceptance); no-op on every other path.
		AuthRouteRateLimit,
		// Hard 15 s request deadline — cancels in-flight DB queries when the
		// deadline fires. Also logs a warn for requests >5 s.
		RequestTimeoutMiddleware,
		// Counter-only middleware — /metrics itself skips it so a scrape doesn't
		// inflate its own numbers.
		httpmetrics.Middleware,
		// Reports any call to a Flow v1 route that the workspace-removal Phase 4
		// deletes. RequestLoggerMiddleware only reports 4xx/5xx, so a successful
		// v1 call was previously invisible and the phase's "zero v1 traffic" gate
		// could not be answered from logs. Delete with the v1 handlers.
		FlowV1UsageMiddleware,
	)

	return middlewares
}
