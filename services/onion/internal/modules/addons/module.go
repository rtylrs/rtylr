package addons

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/addons/internal/application"
	"github.com/voxire/onion/internal/modules/addons/internal/routes"
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

// GetService exposes the service for other modules to check addon status.
func (m *Module) GetService() *application.Service {
	return m.service
}
