package workspaces

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/workspaces/internal/api"
	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/realtime"
	"github.com/voxire/onion/internal/platform/logger"
	flowcyclev1 "github.com/voxire/voxie/gen/go/flow/cycle/v1"
	flowdashboardv1 "github.com/voxire/voxie/gen/go/flow/dashboard/v1"
	flowinsightsv1 "github.com/voxire/voxie/gen/go/flow/insights/v1"
	flowlabelv1 "github.com/voxire/voxie/gen/go/flow/label/v1"
	flowmilestonev1 "github.com/voxire/voxie/gen/go/flow/milestone/v1"
	flownotificationv1 "github.com/voxire/voxie/gen/go/flow/notification/v1"
	flowprojectv1 "github.com/voxire/voxie/gen/go/flow/project/v1"
	flowsettingv1 "github.com/voxire/voxie/gen/go/flow/setting/v1"
	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
	flowviewv1 "github.com/voxire/voxie/gen/go/flow/view/v1"
	flowworkspacev1 "github.com/voxire/voxie/gen/go/flow/workspace/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
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
}

func (m *Module) RegisterRoutes(_ context.Context, router *chi.Mux, _ ...interface{}) {
	// Realtime WebSocket endpoint. Registered as a raw chi handler — the
	// standard RegisterGET wrapper would try to decode the request body
	// and serialize a JSON response, both of which break the WS upgrade.
	router.Get("/flow/ws", realtime.ServeRealtime)
	router.Post("/flow/ws-ticket", realtime.IssueRealtimeTicket)
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
// Safe to call multiple times — idempotent.
func (m *Module) CreatePersonalWorkspace(ctx context.Context, userID int64) {
	if _, err := m.service.CreatePersonalWorkspace(ctx, userID); err != nil {
		logger.Ctx(ctx).Warn("failed to create personal workspace",
			logger.Any("user_id", userID),
			logger.Error(err),
		)
	}
}
