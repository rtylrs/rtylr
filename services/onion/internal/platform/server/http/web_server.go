package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/observability"
)

const serverShutdownTimeoutSeconds = 5

func Start(ctx context.Context, runMode ChiRunMode, port string) error {
	// Register the server context so rate-limiter cleanup goroutines stop
	// when this server instance shuts down (avoids accumulation on restart).
	serverCtx = ctx

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// Sentry recover runs after chi's Recoverer so we still respond
	// with a 500, but anything that escapes the handler layer ends
	// up in Sentry with a proper stack trace and request scope.
	r.Use(observability.RecoverMiddleware)
	r.Use(corsMiddleware())
	r.Use(CrashNotifierMiddleware())
	r.Use(SecurityHeadersMiddleware())
	r.Use(SecureErrorMiddleware())
	r.Use(RateLimitMiddleware(100, time.Minute))
	// Hard 15 s request deadline — cancels in-flight DB queries when
	// the deadline fires. Also logs a warn for requests >5 s.
	r.Use(RequestTimeoutMiddleware)
	// Deduplicates mutating requests that carry Idempotency-Key header.
	// Runs after auth so user-scoped keys work correctly.
	r.Use(IdempotencyMiddleware)
	// Counter-only middleware — /metrics itself skips it so a scrape
	// doesn't inflate its own numbers.
	r.Use(MetricsMiddleware)

	runMode.Setup(ctx, r)

	// /metrics is mounted before runMode.Setup's module routes run so
	// there's no way a module accidentally shadows it. Basic-auth is
	// enforced inside the handler; the endpoint 404s when
	// METRICS_USER/METRICS_PASS aren't set, so it stays opt-in.
	r.Get("/metrics", MetricsHandler())

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			logger.Ctx(ctx).Warn("Failed to write response", logger.Error(err))
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

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
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
