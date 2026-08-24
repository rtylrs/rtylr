package workspaces

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	hrcore "github.com/voxire/onion/internal/modules/hr_core"
	hrleave "github.com/voxire/onion/internal/modules/hr_leave"
	"github.com/voxire/onion/internal/modules/workspaces/internal/api"
	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/realtime"
	"github.com/voxire/onion/internal/platform/logger"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/tenant"
	flowcyclev1 "github.com/voxire/voxie/gen/go/flow/cycle/v1"
	flowcyclev2 "github.com/voxire/voxie/gen/go/flow/cycle/v2"
	flowdashboardv1 "github.com/voxire/voxie/gen/go/flow/dashboard/v1"
	flowdashboardv2 "github.com/voxire/voxie/gen/go/flow/dashboard/v2"
	flowinsightsv1 "github.com/voxire/voxie/gen/go/flow/insights/v1"
	flowinsightsv2 "github.com/voxire/voxie/gen/go/flow/insights/v2"
	flowlabelv1 "github.com/voxire/voxie/gen/go/flow/label/v1"
	flowlabelv2 "github.com/voxire/voxie/gen/go/flow/label/v2"
	flowmilestonev1 "github.com/voxire/voxie/gen/go/flow/milestone/v1"
	flowmilestonev2 "github.com/voxire/voxie/gen/go/flow/milestone/v2"
	flownotificationv1 "github.com/voxire/voxie/gen/go/flow/notification/v1"
	flowprofilev1 "github.com/voxire/voxie/gen/go/flow/profile/v1"
	flowprojectv1 "github.com/voxire/voxie/gen/go/flow/project/v1"
	flowprojectv2 "github.com/voxire/voxie/gen/go/flow/project/v2"
	flowsettingv1 "github.com/voxire/voxie/gen/go/flow/setting/v1"
	flowstatusv2 "github.com/voxire/voxie/gen/go/flow/status/v2"
	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
	flowtimev1 "github.com/voxire/voxie/gen/go/flow/time/v1"
	flowtimev2 "github.com/voxire/voxie/gen/go/flow/time/v2"
	flowviewv1 "github.com/voxire/voxie/gen/go/flow/view/v1"
	flowviewv2 "github.com/voxire/voxie/gen/go/flow/view/v2"
	flowworkspacev1 "github.com/voxire/voxie/gen/go/flow/workspace/v1"
)

type Module struct {
	service *application.Service
	hrCore  *hrcore.Module
	hrLeave *hrleave.Module
}

func NewModule(ctx context.Context, db *sql.DB, hrCoreModule *hrcore.Module, hrLeaveModule *hrleave.Module) *Module {
	return &Module{
		service: application.NewService(ctx, db),
		hrCore:  hrCoreModule,
		hrLeave: hrLeaveModule,
	}
}

// Service exposes the module's application service for callers that need
// to cross module boundaries — specifically the realtime hub authorizer,
// which consults workspace/project membership to decide subscription
// eligibility.
func (m *Module) Service() *application.Service {
	return m.service
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := flowworkspacev1.RegisterWorkspaceServiceServer(
		api.NewWorkspaceAPI(m.service),
		flowworkspacev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow workspace service: %w", err))
	}
	if err := flowprojectv1.RegisterProjectServiceServer(
		api.NewProjectAPI(m.service),
		flowprojectv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow project service: %w", err))
	}
	if err := flowprojectv2.RegisterProjectServiceServer(
		api.NewProjectAPIV2(m.service),
		flowprojectv2.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow project v2 service: %w", err))
	}
	if err := flowlabelv1.RegisterLabelServiceServer(
		api.NewLabelAPI(m.service),
		flowlabelv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow label service: %w", err))
	}
	if err := flowtaskv1.RegisterTaskServiceServer(
		api.NewTaskAPI(m.service),
		flowtaskv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow task service: %w", err))
	}
	if err := flowtaskv2.RegisterTaskServiceServer(
		api.NewTaskAPIV2(m.service),
		flowtaskv2.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow task v2 service: %w", err))
	}
	if err := flowdashboardv2.RegisterDashboardServiceServer(
		api.NewDashboardAPIV2(m.service),
		flowdashboardv2.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow dashboard v2: %w", err))
	}
	if err := flowinsightsv2.RegisterInsightsServiceServer(
		api.NewInsightsAPIV2(m.service),
		flowinsightsv2.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow insights v2: %w", err))
	}
	if err := flowlabelv2.RegisterLabelServiceServer(
		api.NewLabelAPIV2(m.service),
		flowlabelv2.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow label v2: %w", err))
	}
	if err := flowstatusv2.RegisterStatusServiceServer(
		api.NewStatusAPIV2(m.service),
		flowstatusv2.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow status v2: %w", err))
	}
	if err := flowcyclev2.RegisterCycleServiceServer(
		api.NewCycleAPIV2(m.service),
		flowcyclev2.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow cycle v2: %w", err))
	}
	if err := flowmilestonev2.RegisterMilestoneServiceServer(
		api.NewMilestoneAPIV2(m.service),
		flowmilestonev2.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow milestone v2: %w", err))
	}
	if err := flowviewv2.RegisterViewServiceServer(
		api.NewViewAPIV2(m.service),
		flowviewv2.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow view v2: %w", err))
	}
	if err := flowdashboardv1.RegisterDashboardServiceServer(
		api.NewDashboardAPI(m.service),
		flowdashboardv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow dashboard service: %w", err))
	}
	if err := flowinsightsv1.RegisterInsightsServiceServer(
		api.NewInsightsAPI(m.service),
		flowinsightsv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow insights service: %w", err))
	}
	if err := flowmilestonev1.RegisterMilestoneServiceServer(
		api.NewMilestoneAPI(m.service),
		flowmilestonev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow milestone service: %w", err))
	}
	if err := flowcyclev1.RegisterCycleServiceServer(
		api.NewCycleAPI(m.service),
		flowcyclev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow cycle service: %w", err))
	}
	if err := flowviewv1.RegisterViewServiceServer(
		api.NewViewAPI(m.service),
		flowviewv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow view service: %w", err))
	}
	if err := flownotificationv1.RegisterNotificationServiceServer(
		api.NewNotificationAPI(m.service),
		flownotificationv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow notification service: %w", err))
	}
	if err := flowsettingv1.RegisterSettingServiceServer(
		api.NewSettingAPI(m.service),
		flowsettingv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow setting service: %w", err))
	}
	if err := flowprofilev1.RegisterProfileServiceServer(
		api.NewProfileAPI(m.service),
		flowprofilev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow profile service: %w", err))
	}
	if err := flowtimev2.RegisterTimeServiceServer(
		api.NewWorkTimeAPIV2(m.service),
		flowtimev2.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow time v2 service: %w", err))
	}
	if err := flowtimev1.RegisterTimeServiceServer(
		api.NewWorkTimeAPI(m.service),
		flowtimev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow time service: %w", err))
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	// Realtime WebSocket endpoint. Registered as a raw chi handler — the
	// standard RegisterGET wrapper would try to decode the request body
	// and serialize a JSON response, both of which break the WS upgrade.
	router.Get("/flow/ws", realtime.ServeRealtime)
	router.Post("/flow/ws-ticket", realtime.IssueRealtimeTicket)

	serverHttp.RegisterGET[PresenceRequest, PresenceResponse](
		ctx, router, PresenceRoutes(), &PresenceHandler{Module: m},
	)
}

// GetAnonymousRoutes tells the global auth middleware to skip bearer-token
// validation for the browser WebSocket path. Browsers cannot set the
// Authorization header on a WebSocket handshake, so /flow/ws authenticates
// its short-lived ticket inside the handler. /flow/ws-ticket remains protected
// by the normal Bearer middleware.
func (m *Module) GetAnonymousRoutes() []string {
	return []string{"/flow/ws"}
}

// CreatePersonalWorkspace creates a personal workspace for a newly registered user.
// Safe to call multiple times — idempotent. The hook this is wired to
// (user.OnUserCreated/OnUserLogin) only carries a userID, so businessID is
// resolved here from the user's own default business record.
func (m *Module) CreatePersonalWorkspace(ctx context.Context, userID int64) {
	businessID, bizErr := tenant.ResolveBusinessID(ctx, 0, userID)
	if bizErr != nil {
		logger.Ctx(ctx).Warn("failed to resolve business for personal workspace",
			logger.Any("user_id", userID),
			logger.Error(bizErr),
		)
		return
	}
	if _, err := m.service.CreatePersonalWorkspace(ctx, userID, businessID); err != nil {
		logger.Ctx(ctx).Warn("failed to create personal workspace",
			logger.Any("user_id", userID),
			logger.Error(err),
		)
	}
}
