package pos_menu

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_menu/internal/application"
	"github.com/voxire/onion/internal/modules/pos_menu/internal/routes"
	"github.com/voxire/onion/internal/modules/pos_menu/internal/routes/handlers"
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
	// Public-menu read is the one anonymous surface in pos_menu.
	// Every other route is tenant-scoped and requires a JWT.
	return handlers.PublicListMenuItemsRoutes()
}
