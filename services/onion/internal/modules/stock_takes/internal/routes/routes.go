package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/stock_takes/internal/application"
	"github.com/voxire/onion/internal/modules/stock_takes/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListStockTakesRequest, handlers.ListStockTakesResponse](
		ctx,
		router,
		handlers.ListStockTakesRoutes(),
		&handlers.ListStockTakesHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateStockTakeRequest, handlers.CreateStockTakeResponse](
		ctx,
		router,
		handlers.CreateStockTakeRoutes(),
		&handlers.CreateStockTakeHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateStockTakeRequest, handlers.UpdateStockTakeResponse](
		ctx,
		router,
		handlers.UpdateStockTakeRoutes(),
		&handlers.UpdateStockTakeHandler{Service: service},
	)
}
