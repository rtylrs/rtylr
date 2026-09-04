package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterPOSTWithValidation[handlers.VoidOrderApprovalRequest, handlers.VoidOrderApprovalResponse](
		ctx, router, handlers.ApproveVoidOrderRoutes(), &handlers.ApproveVoidOrderHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.VoidOrderApprovalRequest, handlers.VoidOrderApprovalResponse](
		ctx, router, handlers.OverrideVoidOrderRoutes(), &handlers.OverrideVoidOrderHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.ProcessRefundRequest, handlers.ProcessRefundResponse](
		ctx, router, handlers.ApproveRefundOrderRoutes(), &handlers.ApproveRefundOrderHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.ProcessRefundRequest, handlers.ProcessRefundResponse](
		ctx, router, handlers.OverrideRefundOrderRoutes(), &handlers.OverrideRefundOrderHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.PayOrderRequest, handlers.PayOrderResponse](
		ctx, router, handlers.PayOrderRoutes(),
		&handlers.PayOrderHandler{Service: service},
	)

	// Paddle signs the raw request body, so this webhook must own body reads.
	router.Post("/pos/payments/webhook", handlers.PaddlePOSWebhookHTTPHandler(service))

	router.Get("/pos/ws", handlers.ServePOSWebSocket)
	// The ticket endpoint is a websocket transport preflight, not a JSON API
	// operation, and is authenticated by the websocket ticket handler.
	router.Post("/pos/ws-ticket", handlers.IssuePOSWebSocketTicket)
}
