package main

import (
	"context"
	stdhttp "net/http"

	"github.com/voxire/onion/internal/modules/organization/scopes"

	"github.com/go-chi/chi/v5"

	hrdashboardstats "github.com/voxire/onion/internal/modules/hr/hr_dashboard/stats"
	"github.com/voxire/onion/internal/modules/organization/addons"
	"github.com/voxire/onion/internal/modules/organization/team"
	"github.com/voxire/onion/internal/modules/organization/user"
	"github.com/voxire/onion/internal/modules/webhooks"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache"
	"github.com/voxire/onion/internal/platform/cdn"
	platformConfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/observability"
	"github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/storage"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/platform/wsbus"
)

// setupRoutes registers every API route on r and returns the middlewares the
// server should wrap it with. It does not apply them itself: the HTTP server
// builds one chain so the Partials X-API-KEY bypass can sit outside it.
func setupRoutes(ctx context.Context, r *chi.Mux) []http.MiddlewareFunc {
	modules := initializeModules(ctx)

	// Start the websocket bus with topic-level authorization backed by
	// workspace/project membership checks. The rules:
	//
	//   user:<id>        → caller must be that user
	//   owner:<id>       → caller must own or belong to that owner account
	//   business:<id>    → caller must belong to that business
	//   workspace:<id>   → caller must belong to the workspace
	//   project:<id>     → caller must be a project or workspace member
	//   task:<id>        → caller must be able to access the task's project
	//
	// Unknown topic schemes are denied by default.
	workspacesSvc := modules.Workspaces.Service()
	teamSvc := modules.Team.Service()
	// Employment-driven access revocation (#834): when hr_core terminates an
	// employee, team (the entitlement owner) drops that user's permission rows
	// for the business. team's daily reconciliation job is the durability net.
	modules.HRCore.SetTerminationRevoker(teamSvc)
	hubResolver := team.NewHubProfileResolver(modules.Team.Service())
	wsbus.SetDefault(ctx, wsbus.New(func(ctx context.Context, userID int64, topic string) bool {
		scope, id, ok := wsbus.ParseTopic(topic)
		if !ok {
			return false
		}
		switch scope {
		case "user":
			return id == userID
		case "owner":
			return teamSvc.CanAccessOwner(ctx, userID, id)
		case "business":
			// Workspace-removal topic: project lifecycle events fan out here.
			// The authorizer only gets a user id, so membership resolves via
			// the hub profile: a resolvable profile means the caller belongs
			// to that business.
			_, profileErr := hubResolver.GetProfile(ctx, userID, id)
			return profileErr == nil
		case "workspace":
			return workspacesSvc.IsWorkspaceMember(ctx, userID, id)
		case "project":
			return workspacesSvc.IsProjectMember(ctx, userID, id)
		case "task":
			return workspacesSvc.IsTaskMember(ctx, userID, id)
		default:
			return false
		}
	}))

	// Fan realtime events across replicas when Redis is available; otherwise the
	// hub stays in-process, which is correct for a single replica.
	if platformConfig.IsRedisEnabled() && cache.RedisAvailable() {
		wsbus.EnableRedisBackplane(ctx)
	} else if platformConfig.IsRedisEnabled() {
		logger.Ctx(ctx).Warn("Redis backplane disabled; using in-process websocket fan-out")
	}

	accessPolicy := http.NewRouteAccessPolicy()
	accessPolicy.RegisterAnonymousRoutes("/health")
	accessPolicy.RegisterAnonymousRoutes("/version")
	accessPolicy.RegisterAnonymousRoutes("/flow/ws")
	accessPolicy.RegisterAnonymousRoutes("/invalidate-cache")

	// CDN routes (write endpoints protected by CDN API key auth)
	accessPolicy.RegisterAnonymousRoutes(
		"/upload",
		"/upload/plan",
		"/upload/responsive",
		"/upload/variants",
		"/capabilities",
		"/files/{id}",
		"/images/{id}",
		"/images/{id}/analyze",
		"/images/{id}/info",
		"/images/{id}/plan",
		"/images/{id}/responsive",
		"/images/{id}/variants",
		"/videos/info",
		"/purge",
		"/sign",
	)

	provisionerResolver := authPkg.NewStaticPlatformProvisionerResolver(
		nil,
	)
	modules.User.ProvisionerResolver = provisionerResolver
	modules.User.ProfileResolver = team.NewHubProfileResolver(modules.Team.Service())
	modules.User.WithProvisioningStore(teamSvc)
	modules.User.WithInviteAccepter(modules.Team.Service())
	// A business-wide Owner/Admin/Operator has no path through Flow's own
	// project/workspace membership tables, so RequireProjectRole/
	// RequireTaskRole/RequireWorkspaceRole consult this resolver first and
	// admit them unconditionally before falling back to Flow's own roles.
	modules.Workspaces.Service().SetHubProfileResolver(team.NewHubProfileResolver(modules.Team.Service()))
	// Dept-manager visibility proxy (M3): Flow asks HR who reports to a caller
	// and grants them member-equivalent access to those people's projects. Until
	// this is wired the proxy is inert, so a worker or test without it sees
	// exactly the pre-proxy behaviour.
	modules.Workspaces.Service().SetOrgchartResolver(modules.Staff.OrgchartResolver())
	modules.HRDashboard.WithStatsLoader(func(ctx context.Context, businessID int64) (hrdashboardstats.DashboardStats, error) {
		totalStaff, onShiftToday, pendingTimeOff, payrollDraft, err := modules.Staff.LoadDashboardStats(ctx, businessID)
		if err != nil {
			return hrdashboardstats.DashboardStats{}, err
		}
		return hrdashboardstats.DashboardStats{
			TotalStaff:     totalStaff,
			OnShiftToday:   onShiftToday,
			PendingTimeOff: pendingTimeOff,
			PayrollDraft:   payrollDraft,
		}, nil
	})
	modules.Business.WithUserService(modules.User.Onboarding())
	modules.Apps.SetAddonProvider(addons.NewHubAddonProvider(modules.Addons.GetService()))
	modules.APIKeys.SetWebhooksProvider(webhooks.NewHubWebhooksProvider(modules.Webhooks.GetService()))

	// Wire personal workspace creation into user registration and login (idempotent).
	ensurePersonalWorkspace := func(ctx context.Context, userID int64) {
		modules.Workspaces.CreatePersonalWorkspace(ctx, userID)
	}
	user.OnUserCreated = ensurePersonalWorkspace
	user.OnUserLogin = ensurePersonalWorkspace

	http.AutoRegisterAnonymousRoutes(ctx, accessPolicy, modules.anonymousRouteProviders()...)

	tokenResolver := authPkg.NewStaticTokenValidatorResolver(
		modules.User.GetTokenValidator(),
		map[authPkg.ProductPlatform]authPkg.TokenValidator{
			authPkg.ProductPlatformRtylr:  modules.User.GetTokenValidator(),
			authPkg.ProductPlatformVoxire: modules.User.GetTokenValidator(),
		},
	)

	middlewares := []http.MiddlewareFunc{
		http.SecurityMiddleware(),
		jwtutils.JWTMiddleware(),
		http.RequestLoggerMiddleware(ctx),
		http.GetAuthMiddlewareFunc(accessPolicy, tokenResolver, teamSvc),
		authPkg.DeviceMiddleware(),
		// Keep subscription state authoritative for every authenticated tenant
		// request. The middleware leaves billing recovery and console routes open.
		http.SubscriptionEnforcementMiddleware(scopes.NewReader(dbclient.Pool()).SubscriptionAccess, accessPolicy),
		// After auth so idempotency keys are scoped per authenticated user.
		http.IdempotencyMiddleware,
	}

	http.RegisterCacheRoutes(ctx, r)
	http.RegisterStatusRoutes(ctx, r)
	http.RegisterVersionRoutes(r)

	// Consolidated CDN / Upload routes
	cfg := platformConfig.GlobalConfig
	store, storeErr := storage.Default()
	if storeErr != nil {
		logger.Ctx(ctx).Warn("object storage is not configured; CDN routes will degrade", logger.Error(storeErr))
	}
	cdnService := cdn.New(
		dbclient.Pool(),
		store,
		cfg.CDNBaseURL,
		cfg.CDNMaxUploadBytes,
		cfg.CDNMediaEnginePath,
		cfg.CDNAPIKey,
	).WithUploadKeyReader(scopes.NewReader(dbclient.Pool()).UploadKeyScope)
	if err := cdnService.CheckMediaEngine(); err != nil {
		logger.Ctx(ctx).Warn("media-engine unavailable; image transformations will fail", logger.Error(err))
	}
	if moderator, failClosed := cdn.NewModeratorFromConfig(cfg); moderator != nil {
		cdnService.SetModerator(moderator, failClosed)
		logger.Ctx(ctx).Info("CDN image moderation enabled",
			logger.String("provider", cfg.CDNModerationProvider),
			logger.Bool("fail_closed", failClosed),
		)
	}
	cdnService.RegisterRoutes(r)

	// oneKit contract services: every module's generated service servers
	// register on one shared mux; the generated code owns every contract path.
	// Runs before RegisterRoutes so module escape hatches (rate-limited or
	// special routes) can reference the mux captured in RegisterAPI.
	oneKitMux := stdhttp.NewServeMux()
	http.AutoRegisterAPIs(ctx, oneKitMux, modules.apiRegistrars()...)

	// Modules that still use the Register* helpers fill a route table instead
	// of a router. Routes no contract serves yet are mounted on the same mux.
	legacyRoutes := http.NewLegacyRoutes()
	http.AutoRegisterRoutes(ctx, legacyRoutes, modules.routeRegistrars()...)
	registerCompatContracts(oneKitMux, legacyRoutes)
	legacyRoutes.Mount(oneKitMux, compatRoutes().Contains)
	observability.SetRouteMux(oneKitMux)
	http.MountOneKit(r, oneKitMux)

	return middlewares
}

type apiServer struct {
	middlewares []http.MiddlewareFunc
}

func (a *apiServer) Setup(ctx context.Context, r *chi.Mux) {
	a.middlewares = setupRoutes(ctx, r)
}

func (a *apiServer) Middlewares() []http.MiddlewareFunc {
	return a.middlewares
}

func startAPIServer(ctx context.Context) error {
	server := &apiServer{}
	return http.Start(ctx, server, platformConfig.ListenPort("9999"))
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
