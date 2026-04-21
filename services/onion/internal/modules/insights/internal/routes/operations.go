package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/modules/insights/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func registerOperationsRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	serverHttp.RegisterGET[handlers.GetOperationsKPIsRequest, handlers.GetOperationsKPIsResponse](
		ctx, router, handlers.GetOperationsKPIsRoutes(),
		&handlers.GetOperationsKPIsHandler{Service: svc},
	)
	serverHttp.RegisterGET[handlers.GetHeadcountRequest, handlers.GetHeadcountResponse](
		ctx, router, handlers.GetHeadcountRoutes(),
		&handlers.GetHeadcountHandler{Service: svc},
	)
	serverHttp.RegisterGET[handlers.GetShiftMetricsRequest, handlers.GetShiftMetricsResponse](
		ctx, router, handlers.GetShiftMetricsRoutes(),
		&handlers.GetShiftMetricsHandler{Service: svc},
	)
}
