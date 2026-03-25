package credits

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/platforms/zontriq/credits/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/credits/internal/routes"
)

type Module struct {
	creditsService *application.Service
}

func NewModule(ctx context.Context) *Module {
	creditsService := application.NewService(ctx)

	return &Module{
		creditsService: creditsService,
	}
}

func (m *Module) GetService() *application.Service {
	return m.creditsService
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.creditsService)
}
