package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_stock_movements/internal/application"
	"github.com/voxire/onion/internal/modules/pos_stock_movements/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListMovementsRequest, handlers.ListMovementsResponse](
		ctx, router, handlers.ListMovementsRoutes(),
		&handlers.ListMovementsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.ReceiveStockRequest, handlers.ReceiveStockResponse](
		ctx, router, handlers.ReceiveStockRoutes(),
		&handlers.ReceiveStockHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.AdjustStockRequest, handlers.AdjustStockResponse](
		ctx, router, handlers.AdjustStockRoutes(),
		&handlers.AdjustStockHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.WasteStockRequest, handlers.WasteStockResponse](
		ctx, router, handlers.WasteStockRoutes(),
		&handlers.WasteStockHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.TransferStockRequest, handlers.TransferStockResponse](
		ctx, router, handlers.TransferStockRoutes(),
		&handlers.TransferStockHandler{Service: svc},
	)

	http.RegisterGET[handlers.ListInventoryRequest, handlers.ListInventoryResponse](
		ctx, router, handlers.ListInventoryRoutes(),
		&handlers.ListInventoryHandler{Service: svc},
	)

	http.RegisterGET[handlers.LowStockRequest, handlers.LowStockResponse](
		ctx, router, handlers.LowStockRoutes(),
		&handlers.LowStockHandler{Service: svc},
	)
}
