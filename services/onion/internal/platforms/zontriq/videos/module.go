package videos

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/platforms/zontriq/videos/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/videos/internal/routes"
)

type Module struct {
	videoService *application.Service
}

func NewModule(ctx context.Context) *Module {
	videoService := application.NewService(ctx)

	return &Module{
		videoService: videoService,
	}
}

func (m *Module) GetService() *application.Service {
	return m.videoService
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.videoService)
}
