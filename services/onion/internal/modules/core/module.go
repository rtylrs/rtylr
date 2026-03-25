package core

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/core/internal/application"
	"github.com/voxire/onion/internal/modules/core/internal/routes"
)

type Module struct {
	Service *application.Service
}

func NewModule(ctx context.Context) *Module {
	service := application.NewService(ctx)

	module := &Module{
		Service: service,
	}

	return module
}

func (m *Module) GetService() *application.Service {
	return m.Service
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.Service)
}
