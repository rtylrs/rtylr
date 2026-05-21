package documents

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/documents/internal/application"
	"github.com/voxire/onion/internal/modules/documents/internal/routes"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
