package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/purchase_orders/internal/application"
	"github.com/voxire/onion/internal/modules/purchase_orders/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListPurchaseOrdersRequest, handlers.ListPurchaseOrdersResponse](
		ctx,
		router,
		handlers.ListPurchaseOrdersRoutes(),
		&handlers.ListPurchaseOrdersHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreatePurchaseOrderRequest, handlers.CreatePurchaseOrderResponse](
		ctx,
		router,
		handlers.CreatePurchaseOrderRoutes(),
		&handlers.CreatePurchaseOrderHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdatePurchaseOrderStatusRequest, handlers.UpdatePurchaseOrderStatusResponse](
		ctx,
		router,
		handlers.UpdatePurchaseOrderStatusRoutes(),
		&handlers.UpdatePurchaseOrderStatusHandler{Service: service},
	)
}
