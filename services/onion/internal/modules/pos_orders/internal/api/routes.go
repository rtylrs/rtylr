package api

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGETWithValidation[ListOrdersRequest, ListOrdersResponse](
		ctx, router, ListOrdersRoutes(), &ListOrdersHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[VoidOrderApprovalRequest, VoidOrderApprovalResponse](
		ctx, router, ApproveVoidOrderRoutes(), &ApproveVoidOrderHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[VoidOrderApprovalRequest, VoidOrderApprovalResponse](
		ctx, router, OverrideVoidOrderRoutes(), &OverrideVoidOrderHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[ProcessRefundRequest, ProcessRefundResponse](
		ctx, router, ApproveRefundOrderRoutes(), &ApproveRefundOrderHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[ProcessRefundRequest, ProcessRefundResponse](
		ctx, router, OverrideRefundOrderRoutes(), &OverrideRefundOrderHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[PayOrderRequest, PayOrderResponse](
		ctx, router, PayOrderRoutes(),
		&PayOrderHandler{Service: service},
	)

	// Paddle signs the raw request body, so this webhook must own body reads.
	router.Post("/pos/payments/webhook", PaddlePOSWebhookHTTPHandler(service))

	router.Get("/pos/ws", ServePOSWebSocket)
	// The ticket endpoint is a websocket transport preflight, not a JSON API
	// operation, and is authenticated by the websocket ticket handler.
	router.Post("/pos/ws-ticket", IssuePOSWebSocketTicket)
}
