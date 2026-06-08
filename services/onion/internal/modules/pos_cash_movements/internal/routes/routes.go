package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_cash_movements/internal/application"
	"github.com/voxire/onion/internal/modules/pos_cash_movements/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListCashMovementsRequest, handlers.ListCashMovementsResponse](
		ctx, router, handlers.ListCashMovementsRoutes(),
		&handlers.ListCashMovementsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CashDropRequest, handlers.CashDropResponse](
		ctx, router, handlers.CashDropRoutes(),
		&handlers.CashDropHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CashPickupRequest, handlers.CashPickupResponse](
		ctx, router, handlers.CashPickupRoutes(),
		&handlers.CashPickupHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CashAdjustRequest, handlers.CashAdjustResponse](
		ctx, router, handlers.CashAdjustRoutes(),
		&handlers.CashAdjustHandler{Service: svc},
	)
}
