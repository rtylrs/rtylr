package series

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/platforms/zontriq/series/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/series/internal/routes"
)

type Module struct {
	SeriesService *application.Service
}

func NewModule(ctx context.Context) *Module {
	seriesService := application.NewService(ctx)

	return &Module{
		SeriesService: seriesService,
	}
}

func (m *Module) GetService() *application.Service {
	return m.SeriesService
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.SeriesService)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
