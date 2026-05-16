package hr_shifts

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/application"
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/routes"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context) *Module {
	return &Module{
		service: application.NewService(ctx, nil),
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
