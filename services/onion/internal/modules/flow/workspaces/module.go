package workspaces

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/api"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application/identity"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/realtime"
	"github.com/voxire/onion/internal/platform/logger"
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
	flowpresencev1 "github.com/voxire/voxie/gen/go/flow/presence/v1"
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
	service    *application.Service
	repository *moduleinfra.Repository
	presence   PresenceSource
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repository := moduleinfra.NewRepository(db)
	return &Module{
		service:    application.NewService(ctx, repository),
		repository: repository,
	}
}

type IdentityKey = identity.IdentityKey
type Identity = identity.Identity
type IdentityReader = identity.IdentityReader

func (m *Module) WithIdentityReaders(reader IdentityReader, email func(context.Context, string) (int64, error), username func(context.Context, int64, string) (int64, error)) *Module {
	m.repository.WithIdentityReaders(reader, email, username)
	return m
}

func (m *Module) SetWorkDayLineStore(store WorkDayLineStore) {
	if m == nil || m.repository == nil {
		return
	}
	if store == nil {
		m.repository.SetWorkDayLineStore(nil)
		return
	}
	adapter := workDayLineStoreAdapter{store: store}
	m.repository.SetWorkDayLineStore(adapter)
	m.service.SetWorkDayLineStore(adapter)
}

// WorkdayRoster and WorkdayRosterSource are the minimal directory contract
// Flow needs to resolve employee-keyed work days.
type WorkdayRoster = application.WorkdayRoster
type WorkdayRosterSource = application.WorkdayRosterSource
type InsightsMemberProfile = application.InsightsMemberProfile
type InsightsMemberProfileSource = application.InsightsMemberProfileSource

func (m *Module) SetWorkdayRosterSource(source WorkdayRosterSource) {
	m.service.SetWorkdayRosterSource(source)
}

func (m *Module) SetInsightsMemberProfileSource(source InsightsMemberProfileSource) {
	m.service.SetInsightsMemberProfileSource(source)
}

func (m *Module) SetPresenceSource(source PresenceSource) {
	m.presence = source
}

// Service exposes the module's application service for callers that need
// to cross module boundaries — specifically the realtime hub authorizer,
// which consults workspace/project membership to decide subscription
// eligibility.
func (m *Module) Service() *application.Service {
	return m.service
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	// These realtime endpoints are deliberate transport-level escape hatches
	// from the JSON API contract. The WebSocket upgrade cannot use the standard
	// JSON handler, while the ticket is its bearer-authenticated preflight.
	mux.HandleFunc("GET /flow/ws", realtime.ServeRealtime)
	mux.HandleFunc("POST /flow/ws-ticket", realtime.IssueRealtimeTicket)
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
	if err := flowpresencev1.RegisterPresenceServiceServer(
		&PresenceAPI{Module: m},
		flowpresencev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register flow presence service: %w", err))
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
