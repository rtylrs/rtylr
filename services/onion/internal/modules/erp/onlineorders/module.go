package onlineorders

import (
	"context"
	"database/sql"

	moduleinfra "github.com/voxire/onion/internal/modules/erp/onlineorders/internal/infrastructure"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/erp/onlineorders/internal/api"
	"github.com/voxire/onion/internal/modules/erp/onlineorders/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
