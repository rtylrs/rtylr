package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/voxire/onion/internal/modules/config"
	"github.com/voxire/onion/internal/modules/user"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/platforms/rtylr"
	"github.com/voxire/onion/internal/platforms/voxire"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for {
		if err := runServer(shutdownCtx); err != nil {
			logger.Global().Error("API server crashed, restarting in 5 seconds..", logger.Error(err))

			select {
			case <-shutdownCtx.Done():
				logger.Global().Info("Shutdown signal received during restart delay, stopping")

				return
			case <-time.After(5 * time.Second):
				continue
			}
		} else {
			return
		}
	}
}

func runServer(shutdownCtx context.Context) error {
	serverCtx, cancel := context.WithCancel(shutdownCtx)
	defer cancel()

	defer recoverAndLog(serverCtx)
	defer logger.Flush()

	// No application struct needed - setup directly in main

	serverErr := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Global().Error("API server panicked, will restart", logger.Any("panic", r))

				serverErr <- &ServerPanicError{panic: r}
			}
		}()

		run(serverCtx)

		serverErr <- nil
	}()

	select {
	case <-shutdownCtx.Done():
		cancel()
		logger.Global().Info("Shutdown signal received, waiting for API server to stop...")

		select {
		case <-serverErr:
			logger.Global().Info("API server stopped gracefully")

			return nil
		case <-time.After(10 * time.Second):
			logger.Global().Warn("API server shutdown timeout, forcing exit")

			return nil
		}
	case err := <-serverErr:
		return err
	}
}

func run(ctx context.Context) {
	logger.Ctx(ctx).Info("Starting API server...")

	startTime := time.Now()

	err := dbclient.Connect(ctx)
	if err != nil {
		logger.Ctx(ctx).Fatal("Database connection failed", logger.Error(err))
	}

	err = cache.Initialize(ctx)
	if err != nil {
		logger.Ctx(ctx).Fatal("Cache initialization failed", logger.Error(err))
	}

	ctx = jwtutils.InitializeJWT(ctx)

	readyTime := time.Now()
	startupDuration := readyTime.Sub(startTime)

	logger.Ctx(ctx).Info("API server is ready",
		logger.Any("startup_time", startupDuration.String()),
	)

	startAPIServer(ctx)
}

type apiModules struct {
	User   *user.Module
	Config *config.Module
	Rtylr  *rtylr.Platform
	Voxire *voxire.Platform
}

func initializeModules(ctx context.Context) *apiModules {
	userModule := user.NewModule(ctx)
	configModule := config.NewModule(ctx)
	rtylrPlatform := rtylr.NewPlatform(ctx)
	voxirePlatform := voxire.NewPlatform(ctx)

	return &apiModules{
		User:   userModule,
		Config: configModule,
		Rtylr:  rtylrPlatform,
		Voxire: voxirePlatform,
	}
}

func setupRoutes(ctx context.Context, r *chi.Mux) {
	modules := initializeModules(ctx)
	accessPolicy := http.NewRouteAccessPolicy()
	accessPolicy.RegisterAnonymousRoutes("/health")

	provisionerResolver := authPkg.NewStaticPlatformProvisionerResolver(
		nil,
	)
	modules.User.ProvisionerResolver = provisionerResolver

	http.AutoRegisterAnonymousRoutes(ctx, modules, accessPolicy)
	http.AutoConfigureAccessPolicy(ctx, modules, accessPolicy)

	tokenResolver := authPkg.NewStaticTokenValidatorResolver(
		modules.User.GetTokenValidator(),
		map[authPkg.ProductPlatform]authPkg.TokenValidator{
			authPkg.ProductPlatformRtylr:  modules.User.GetTokenValidator(),
			authPkg.ProductPlatformVoxire: modules.User.GetTokenValidator(),
		},
	)

	middlewares := []http.MiddlewareFunc{
		http.SecurityMiddleware(),
		http.RequestLoggerMiddleware(ctx),
		http.GetAuthMiddlewareFunc(accessPolicy, tokenResolver),
		authPkg.DeviceMiddleware(),
	}

	http.ApplyMiddlewares(r, middlewares)

	http.RegisterCacheRoutes(ctx, r)
	http.RegisterStatusRoutes(ctx, r)

	http.AutoRegisterRoutes(ctx, modules, r)
}

type apiServer struct{}

func (a apiServer) Setup(ctx context.Context, r *chi.Mux) {
	setupRoutes(ctx, r)
}

func startAPIServer(ctx context.Context) {
	server := apiServer{}
	http.Start(ctx, server, "9999")
}

func recoverAndLog(ctx context.Context) {
	if r := recover(); r != nil {
		logger.Ctx(ctx).Error("Recovered panic", logger.Any("error", r))
	}
}

type ServerPanicError struct {
	panic interface{}
}

func (e *ServerPanicError) Error() string {
	return "server panic occurred"
}

func (e *ServerPanicError) Unwrap() error {
	return nil
}
