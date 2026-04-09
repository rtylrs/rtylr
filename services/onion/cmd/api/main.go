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
	"github.com/voxire/onion/internal/modules/campaigns"
	"github.com/voxire/onion/internal/modules/careers"
	"github.com/voxire/onion/internal/modules/config"
	"github.com/voxire/onion/internal/modules/customers"
	timeoff "github.com/voxire/onion/internal/modules/time_off"
	"github.com/voxire/onion/internal/modules/expenses"
	"github.com/voxire/onion/internal/modules/invoices"
	"github.com/voxire/onion/internal/modules/loyalty"
	"github.com/voxire/onion/internal/modules/echo"
	"github.com/voxire/onion/internal/modules/inventory"
	"github.com/voxire/onion/internal/modules/memory"
	poslocations "github.com/voxire/onion/internal/modules/pos_locations"
	posorders "github.com/voxire/onion/internal/modules/pos_orders"
	possessions "github.com/voxire/onion/internal/modules/pos_sessions"
	postables "github.com/voxire/onion/internal/modules/pos_tables"
	"github.com/voxire/onion/internal/modules/products"
	"github.com/voxire/onion/internal/modules/projects"
	purchaseorders "github.com/voxire/onion/internal/modules/purchase_orders"
	"github.com/voxire/onion/internal/modules/segments"
	"github.com/voxire/onion/internal/modules/departments"
	"github.com/voxire/onion/internal/modules/payroll"
	"github.com/voxire/onion/internal/modules/shifts"
	"github.com/voxire/onion/internal/modules/staff"
	"github.com/voxire/onion/internal/modules/suppliers"
	"github.com/voxire/onion/internal/modules/tasks"
	"github.com/voxire/onion/internal/modules/tax"
	"github.com/voxire/onion/internal/modules/team"
	"github.com/voxire/onion/internal/modules/user"
	"github.com/voxire/onion/internal/modules/workspaces"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
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
	User      *user.Module
	Config    *config.Module
	Careers   *careers.Module
	Echo      *echo.Module
	Memory    *memory.Module
	Products       *products.Module
	Inventory      *inventory.Module
	Suppliers      *suppliers.Module
	PurchaseOrders *purchaseorders.Module
	Staff       *staff.Module
	Departments *departments.Module
	Shifts      *shifts.Module
	Payroll     *payroll.Module
	TimeOff     *timeoff.Module
	Customers   *customers.Module
	Loyalty     *loyalty.Module
	Campaigns   *campaigns.Module
	Segments    *segments.Module
	Invoices    *invoices.Module
	Expenses    *expenses.Module
	Tax         *tax.Module
	Workspaces  *workspaces.Module
	Projects    *projects.Module
	Tasks       *tasks.Module
	PosOrders   *posorders.Module
	PosTables   *postables.Module
	PosSessions *possessions.Module
	PosLocations *poslocations.Module
	Team        *team.Module
}

func initializeModules(ctx context.Context) *apiModules {
	memoryModule := memory.NewModule(ctx)

	return &apiModules{
		User:      user.NewModule(ctx),
		Config:    config.NewModule(ctx),
		Careers:   careers.NewModule(ctx),
		Echo:      echo.NewModule(ctx, memoryModule),
		Memory:    memoryModule,
		Products:       products.NewModule(ctx),
		Inventory:      inventory.NewModule(ctx),
		Suppliers:      suppliers.NewModule(ctx),
		PurchaseOrders: purchaseorders.NewModule(ctx),
		Staff:       staff.NewModule(ctx),
		Departments: departments.NewModule(ctx),
		Shifts:      shifts.NewModule(ctx),
		Payroll:     payroll.NewModule(ctx),
		TimeOff:     timeoff.NewModule(ctx),
		Customers:   customers.NewModule(ctx),
		Loyalty:     loyalty.NewModule(ctx),
		Campaigns:   campaigns.NewModule(ctx),
		Segments:    segments.NewModule(ctx),
		Invoices:    invoices.NewModule(ctx),
		Expenses:    expenses.NewModule(ctx),
		Tax:         tax.NewModule(ctx),
		Workspaces:  workspaces.NewModule(ctx),
		Projects:    projects.NewModule(ctx),
		Tasks:       tasks.NewModule(ctx),
		PosOrders:   posorders.NewModule(ctx),
		PosTables:   postables.NewModule(ctx),
		PosSessions: possessions.NewModule(ctx),
		PosLocations: poslocations.NewModule(ctx),
		Team:        team.NewModule(ctx),
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
