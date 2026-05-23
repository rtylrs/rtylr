package echo

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/echo/internal/application"
	"github.com/voxire/onion/internal/modules/echo/internal/handlers"
	"github.com/voxire/onion/internal/modules/echo/internal/routes"
)

type Module struct {
	snipsService  *application.Service
	memoryService handlers.MemoryService
}

func NewModule(ctx context.Context, memoryService handlers.MemoryService) *Module {
	snipsService, err := application.NewService(ctx)
	if err != nil {
		snipsService = nil
	}

	return &Module{
		snipsService:  snipsService,
		memoryService: memoryService,
	}
}

func (m *Module) GetSnipsService() *application.Service {
	return m.snipsService
}

func (m *Module) RegisterRoutes(ctx context.Context, r *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, r, m.snipsService, m.memoryService)
}
