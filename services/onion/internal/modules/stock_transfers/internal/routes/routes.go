package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/stock_transfers/internal/application"
	"github.com/voxire/onion/internal/modules/stock_transfers/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListStockTransfersRequest, handlers.ListStockTransfersResponse](
		ctx,
		router,
		handlers.ListStockTransfersRoutes(),
		&handlers.ListStockTransfersHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateStockTransferRequest, handlers.CreateStockTransferResponse](
		ctx,
		router,
		handlers.CreateStockTransferRoutes(),
		&handlers.CreateStockTransferHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateStockTransferStatusRequest, handlers.UpdateStockTransferStatusResponse](
		ctx,
		router,
		handlers.UpdateStockTransferStatusRoutes(),
		&handlers.UpdateStockTransferStatusHandler{Service: service},
	)
}
