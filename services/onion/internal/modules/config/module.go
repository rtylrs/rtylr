package config

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/config/internal/application"
	"github.com/voxire/onion/internal/modules/config/internal/routes"
	"github.com/voxire/onion/internal/modules/config/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/utils/sliceutils"
)

type Module struct {
	ConfigService *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	configService := application.NewService(ctx, db)

	module := &Module{
		ConfigService: configService,
	}

	return module
}

func (m *Module) GetService() *application.Service {
	return m.ConfigService
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.ConfigService)
}

func (m *Module) GetAnonymousRoutes() []string {
	return sliceutils.Concat(
		handlers.IndustriesRoutes(),
	)
}
