package memory

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/platforms/rtylr/memory/internal/application"
	"github.com/voxire/onion/internal/platforms/rtylr/memory/internal/routes"
)

type Module struct {
	memoryService *application.Service
}

func NewModule(ctx context.Context) *Module {
	memoryService := application.NewService(ctx)

	return &Module{
		memoryService: memoryService,
	}
}

func (m *Module) GetService() *application.Service {
	return m.memoryService
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.memoryService)
}
