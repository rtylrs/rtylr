package automations

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/platforms/zontriq/automations/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/automations/internal/routes"
)

type Module struct {
	automationsService *application.Service
}

func NewModule(ctx context.Context) *Module {
	automationsService := application.NewService(ctx)

	return &Module{
		automationsService: automationsService,
	}
}

func (m *Module) GetService() *application.Service {
	return m.automationsService
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.automationsService)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
