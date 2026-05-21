package memory

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/memory/internal/application"
	"github.com/voxire/onion/internal/modules/memory/internal/routes"
)

type Module struct {
	memoryService *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	memoryService := application.NewService(ctx, db)

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
