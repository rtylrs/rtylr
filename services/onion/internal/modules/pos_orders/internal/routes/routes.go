package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterPOSTWithValidation[handlers.PayOrderRequest, handlers.PayOrderResponse](
		ctx, router, handlers.PayOrderRoutes(),
		&handlers.PayOrderHandler{Service: service},
	)

	router.Post("/pos/payments/webhook", handlers.PaddlePOSWebhookHTTPHandler(service))

	router.Get("/pos/ws", handlers.ServePOSWebSocket)
}
