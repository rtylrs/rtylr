package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/modules/insights/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func registerProductRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	serverHttp.RegisterGET[handlers.GetProductKPIsRequest, handlers.GetProductKPIsResponse](
		ctx, router, handlers.GetProductKPIsRoutes(),
		&handlers.GetProductKPIsHandler{Service: svc},
	)
	serverHttp.RegisterGET[handlers.GetBestSellersRequest, handlers.GetBestSellersResponse](
		ctx, router, handlers.GetBestSellersRoutes(),
		&handlers.GetBestSellersHandler{Service: svc},
	)
	serverHttp.RegisterGET[handlers.GetMarginDistributionRequest, handlers.GetMarginDistributionResponse](
		ctx, router, handlers.GetMarginDistributionRoutes(),
		&handlers.GetMarginDistributionHandler{Service: svc},
	)
	serverHttp.RegisterGET[handlers.GetSlowMoversRequest, handlers.GetSlowMoversResponse](
		ctx, router, handlers.GetSlowMoversRoutes(),
		&handlers.GetSlowMoversHandler{Service: svc},
	)
	serverHttp.RegisterGET[handlers.GetCategorySalesRequest, handlers.GetCategorySalesResponse](
		ctx, router, handlers.GetCategorySalesRoutes(),
		&handlers.GetCategorySalesHandler{Service: svc},
	)
}
