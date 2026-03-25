package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/voxire/onion/internal/platform/logger"
)

const serverShutdownTimeoutSeconds = 5

func Start(ctx context.Context, runMode ChiRunMode, port string) {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware())
	r.Use(CrashNotifierMiddleware())
	r.Use(SecurityHeadersMiddleware())
	r.Use(SecureErrorMiddleware())
	r.Use(RateLimitMiddleware(100, time.Minute))

	runMode.Setup(ctx, r)

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

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Ctx(ctx).Fatal("Failed to start web server",
				logger.Error(err),
			)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(ctx, serverShutdownTimeoutSeconds*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Ctx(ctx).Error("Failed to shutdown server gracefully",
			logger.Error(err),
		)
	}
}
