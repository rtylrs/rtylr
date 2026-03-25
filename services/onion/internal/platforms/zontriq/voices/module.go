package voices

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/platforms/zontriq/voices/internal/routes"
)

type Module struct{}

func NewModule(_ context.Context) *Module {
	return &Module{}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...any) {
	routes.RegisterRoutes(ctx, router)
}

func (m *Module) GetAnonymousRoutes() []string {
	return nil
}
