package workspaces

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/routes"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context) *Module {
	return &Module{
		service: application.NewService(ctx),
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

// CreatePersonalWorkspace creates a personal workspace for a newly registered user.
// Safe to call multiple times — idempotent.
func (m *Module) CreatePersonalWorkspace(ctx context.Context, userID int64) {
	if _, err := m.service.CreatePersonalWorkspace(ctx, userID); err != nil {
		// Non-fatal; log but don't block registration
		_ = err
	}
}
