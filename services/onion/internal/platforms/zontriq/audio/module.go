package audio

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/platforms/zontriq/audio/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/audio/internal/routes"
)

type Module struct {
	audioService *application.Service
}

func NewModule(ctx context.Context) *Module {
	audioService := application.NewService(ctx)

	return &Module{
		audioService: audioService,
	}
}

func (m *Module) GetService() *application.Service {
	return m.audioService
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.audioService)
}
