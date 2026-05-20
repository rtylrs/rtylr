package business

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/business/internal/application"
	"github.com/voxire/onion/internal/modules/business/internal/routes"
	"github.com/voxire/onion/internal/modules/business/internal/routes/handlers"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	// Public business header for the customer-facing menu app.
	// Every other route in this module requires a JWT.
	return handlers.PublicBusinessRoutes()
}
