package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	cacheInvalidate "github.com/voxire/onion/internal/platform/cache/invalidate"
	platformConfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/emailqueue"
	"github.com/voxire/onion/internal/platform/logger"
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

	logger.Global().Info("worker starting")

	go cacheInvalidate.SubscribeAsyncInvalidations(ctx)
	pool := emailqueue.NewPool(dbclient.Pool())
	pool.Run(ctx)

	logger.Global().Info("worker stopped")
}
