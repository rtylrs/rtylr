package oauth

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/oauth/internal/application"
	"github.com/voxire/onion/internal/modules/oauth/internal/routes"
)

type Module struct {
	service    *application.Service
	authAppURL string
}

func NewModule(ctx context.Context, db *sql.DB, authAppURL string) *Module {
	return &Module{
		service:    application.NewService(ctx, db),
		authAppURL: authAppURL,
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service, m.authAppURL)
}

func (m *Module) GetAnonymousRoutes() []string {
	return routes.AnonymousRoutes()
}
