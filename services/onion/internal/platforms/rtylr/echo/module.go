package echo

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/platforms/rtylr/echo/internal/application"
	"github.com/voxire/onion/internal/platforms/rtylr/echo/internal/routes"
	"github.com/voxire/onion/internal/platforms/rtylr/memory"
)

type Module struct {
	snipsService  *application.Service
	memoryService *memory.Module
}

func NewModule(ctx context.Context, memoryService *memory.Module) *Module {
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
