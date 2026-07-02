// Command upload is the CDN runmode of onion (cdn.voxire.dev). It serves file
// uploads, on-demand image transformations, and asset delivery, reusing the
// shared platform: config, the MySQL pool, the S3 storage client, structured
// logging, and the hardened HTTP server. The media/CDN logic lives in
// internal/cdn; this entrypoint just boots dependencies and mounts the routes.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/cdn"
	platformConfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/server/probe"
	"github.com/voxire/onion/internal/platform/server/profiling"
	"github.com/voxire/onion/internal/platform/storage"
)

func loadDotEnv() {
	platformConfig.LoadDotEnv()
}

// uploadRunMode wires the CDN routes onto the shared chi router.
type uploadRunMode struct {
	cdn *cdn.CDN
}

func (m uploadRunMode) Setup(_ context.Context, r *chi.Mux) {
	m.cdn.RegisterRoutes(r)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	loadDotEnv()
	logger.SetupLogger()

	cfg, err := platformConfig.LoadAndValidate()
	if err != nil {
		logger.Global().Fatal("Configuration validation failed", logger.Error(err))
	}

	if err := dbclient.Connect(ctx); err != nil {
		logger.Global().Fatal("database connection failed", logger.Error(err))
	}
	defer dbclient.Close()
	if err := probe.Start(ctx, probe.Config{
		ServiceName: "upload",
		Port:        probe.PortFromEnv("8888"),
		Stop:        stop,
		ReadyChecks: []probe.Check{
			func(ctx context.Context) error {
				db := dbclient.Pool()
				if db == nil {
					return fmt.Errorf("database pool is not initialized")
				}
				return db.PingContext(ctx)
			},
		},
	}); err != nil {
		logger.Global().Warn("probe server failed to start", logger.Error(err))
	}
	if err := profiling.Start(ctx, profiling.PortFromEnv("6060")); err != nil {
		logger.Global().Warn("pprof server failed to start", logger.Error(err))
	}

	store, err := storage.Default()
	if err != nil {
		// The CDN cannot serve without object storage — unlike feature code
		// that degrades gracefully, this runmode treats it as fatal.
		logger.Global().Fatal("object storage is not configured", logger.Error(err))
	}

	service := cdn.New(
		dbclient.Pool(),
		store,
		cfg.CDNBaseURL,
		cfg.CDNMaxUploadBytes,
		cfg.CDNMediaEnginePath,
		cfg.CDNAPIKey,
	)

	// Surface a misconfigured media-engine at boot rather than on the first
	// /images request. Non-fatal: uploads and original delivery still work.
	if err := service.CheckMediaEngine(); err != nil {
		logger.Ctx(ctx).Warn("media-engine unavailable; image transformations will fail", logger.Error(err))
	}

	if moderator, failClosed := cdn.NewModeratorFromConfig(cfg); moderator != nil {
		service.SetModerator(moderator, failClosed)
		logger.Ctx(ctx).Info("CDN image moderation enabled",
			logger.String("provider", cfg.CDNModerationProvider),
			logger.Bool("fail_closed", failClosed),
		)
	}

	port := platformConfig.ListenPort("8080")

	logger.Ctx(ctx).Info("CDN (upload) runmode starting",
		logger.String("port", port),
		logger.String("base_url", cfg.CDNBaseURL),
	)

	if err := http.Start(ctx, uploadRunMode{cdn: service}, port); err != nil {
		logger.Global().Fatal("CDN server failed", logger.Error(err))
	}
}
