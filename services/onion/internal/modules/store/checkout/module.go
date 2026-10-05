package store_checkout

import (
	"context"
	"database/sql"

	moduleinfra "github.com/voxire/onion/internal/modules/store/checkout/internal/infrastructure"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/store/checkout/internal/api"
	"github.com/voxire/onion/internal/modules/store/checkout/internal/application"
)

type Module struct {
	service    *application.Service
	repository *moduleinfra.Repository
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repo := moduleinfra.NewRepository(db)
	return &Module{service: application.NewService(ctx, repo), repository: repo}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) WithSalesProductReader(read func(context.Context, int64, int64) (string, string, string, string, error)) *Module {
	m.repository.WithSalesProductReader(read)
	return m
}
