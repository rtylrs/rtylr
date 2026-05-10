package insights

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/modules/insights/internal/routes"
)

// Module is the read-only cross-domain aggregation layer that powers the
// /insights/* namespace. It reads (never writes) across Finance, CRM, POS,
// ERP and HR schemas from the infrastructure layer only.
type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context) *Module {
	return &Module{service: application.NewService(ctx)}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
