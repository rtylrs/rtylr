package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/inventory/internal/application"
	"github.com/voxire/onion/internal/modules/inventory/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterPOSTWithValidation[handlers.AdjustInventoryRequest, handlers.AdjustInventoryResponse](
		ctx,
		router,
		handlers.AdjustInventoryRoutes(),
		&handlers.AdjustInventoryHandler{Service: service},
	)

	http.RegisterGET[handlers.InventoryReportRequest, handlers.InventoryReportResponse](
		ctx,
		router,
		handlers.InventoryReportRoutes(),
		&handlers.InventoryReportHandler{Service: service},
	)

	http.RegisterGET[handlers.LowStockRequest, handlers.LowStockResponse](
		ctx,
		router,
		handlers.LowStockRoutes(),
		&handlers.LowStockHandler{Service: service},
	)

	http.RegisterGET[handlers.ValuationRequest, handlers.ValuationResponse](
		ctx,
		router,
		handlers.ValuationRoutes(),
		&handlers.ValuationHandler{Service: service},
	)
}
