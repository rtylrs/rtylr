package workspaces

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/routes"
	"github.com/voxire/onion/internal/modules/workspaces/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/logger"
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

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service)
	// Realtime WebSocket endpoint. Registered as a raw chi handler — the
	// standard RegisterGET wrapper would try to decode the request body
	// and serialize a JSON response, both of which break the WS upgrade.
	router.Get("/flow/ws", handlers.ServeRealtime)
}

// GetAnonymousRoutes tells the global auth middleware to skip bearer-token
// validation for these paths. /flow/ws is "anonymous" at the middleware
// layer because browsers can't set the Authorization header on a
// WebSocket handshake; the handler itself validates the token carried in
// the `?token=` query param before upgrading.
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
