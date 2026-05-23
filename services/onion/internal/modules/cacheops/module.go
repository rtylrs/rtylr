package cacheops

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/cacheops/internal/routes"
)

// Module exposes the cache invalidation public API.
type Module struct{}

func NewModule() *Module { return &Module{} }

func (m *Module) RegisterRoutes(ctx context.Context, r *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, r)
}
