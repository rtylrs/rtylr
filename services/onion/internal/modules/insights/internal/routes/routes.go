package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/insights/internal/application"
)

// RegisterRoutes wires every /insights/* endpoint onto the given router.
// Handlers are registered endpoint-by-endpoint in sibling files in this
// package so that each HTTP surface stays small and cherry-pickable.
func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	registerDashboardRoutes(ctx, router, svc)
	registerRevenueRoutes(ctx, router, svc)
	registerCustomerRoutes(ctx, router, svc)
	registerProductRoutes(ctx, router, svc)
	registerFinancialRoutes(ctx, router, svc)
	registerOperationsRoutes(ctx, router, svc)
	registerReportsRoutes(ctx, router, svc)
	registerSettingsRoutes(ctx, router, svc)
}
