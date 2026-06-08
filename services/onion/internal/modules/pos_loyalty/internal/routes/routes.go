package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/application"
	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.GetConfigRequest, handlers.GetConfigResponse](
		ctx, router, handlers.GetConfigRoutes(),
		&handlers.GetConfigHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateConfigRequest, handlers.UpdateConfigResponse](
		ctx, router, handlers.UpdateConfigRoutes(),
		&handlers.UpdateConfigHandler{Service: svc},
	)

	http.RegisterGET[handlers.ListTransactionsRequest, handlers.ListTransactionsResponse](
		ctx, router, handlers.ListTransactionsRoutes(),
		&handlers.ListTransactionsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.EarnPointsRequest, handlers.EarnPointsResponse](
		ctx, router, handlers.EarnPointsRoutes(),
		&handlers.EarnPointsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.RedeemPointsRequest, handlers.RedeemPointsResponse](
		ctx, router, handlers.RedeemPointsRoutes(),
		&handlers.RedeemPointsHandler{Service: svc},
	)

	http.RegisterGET[handlers.GetBalanceRequest, handlers.GetBalanceResponse](
		ctx, router, handlers.GetBalanceRoutes(),
		&handlers.GetBalanceHandler{Service: svc},
	)
}
