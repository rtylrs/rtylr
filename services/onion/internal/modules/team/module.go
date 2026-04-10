package team

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/modules/team/internal/routes"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context) *Module {
	return &Module{
		service: application.NewService(ctx),
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/auth/invite/verify",
		"/auth/invite/accept",
	}
}
