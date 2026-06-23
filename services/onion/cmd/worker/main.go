package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	cacheInvalidate "github.com/voxire/onion/internal/platform/cache/invalidate"
	"github.com/voxire/onion/internal/platform/cache/invalidation"
	platformConfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/emailqueue"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/server/probe"
	"github.com/voxire/onion/internal/platform/server/profiling"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.SetupLogger()

	if _, err := platformConfig.LoadAndValidate(); err != nil {
		logger.Global().Error("config load failed", logger.Error(err))
		os.Exit(1)
	}

	emailPkg.Initialize()

	if err := dbclient.Connect(ctx); err != nil {
		logger.Global().Error("failed to connect to database", logger.Error(err))
		os.Exit(1)
	}
	defer dbclient.Close()
	if err := probe.Start(ctx, probe.Config{
		ServiceName: "worker",
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

	logger.Global().Info("worker starting")

	go cacheInvalidate.SubscribeAsyncInvalidations(ctx)
	go invalidation.SubscribeCacheGroupInvalidations(ctx)
	pool := emailqueue.NewPool(dbclient.Pool())
	pool.Run(ctx)

	logger.Global().Info("worker stopped")
}
