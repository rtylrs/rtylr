package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// Sessions
	serverHttp.RegisterGET[handlers.GetCurrentSessionRequest, handlers.GetCurrentSessionResponse](
		ctx, router, handlers.GetCurrentSessionRoutes(),
		&handlers.GetCurrentSessionHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.OpenSessionRequest, handlers.OpenSessionResponse](
		ctx, router, handlers.OpenSessionRoutes(),
		&handlers.OpenSessionHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.CloseSessionRequest, handlers.CloseSessionResponse](
		ctx, router, handlers.CloseSessionRoutes(),
		&handlers.CloseSessionHandler{Service: service},
	)

	serverHttp.RegisterGETWithValidation[handlers.ListSessionsRequest, handlers.ListSessionsResponse](
		ctx, router, handlers.ListSessionsRoutes(),
		&handlers.ListSessionsHandler{Service: service},
	)

	// Orders
	serverHttp.RegisterGETWithValidation[handlers.ListOrdersRequest, handlers.ListOrdersResponse](
		ctx, router, handlers.ListOrdersRoutes(),
		&handlers.ListOrdersHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.CreateOrderRequest, handlers.CreateOrderResponse](
		ctx, router, handlers.ListOrdersRoutes(),
		&handlers.CreateOrderHandler{Service: service},
	)

	serverHttp.RegisterGET[handlers.GetOrderDetailRequest, handlers.GetOrderDetailResponse](
		ctx, router, handlers.GetOrderDetailRoutes(),
		&handlers.GetOrderDetailHandler{Service: service},
	)

	serverHttp.RegisterPUTWithValidation[handlers.UpdateOrderStatusRequest, handlers.UpdateOrderStatusResponse](
		ctx, router, handlers.UpdateOrderStatusRoutes(),
		&handlers.UpdateOrderStatusHandler{Service: service},
	)

	serverHttp.RegisterPUTWithValidation[handlers.UpdateOrderItemStatusRequest, handlers.UpdateOrderItemStatusResponse](
		ctx, router, handlers.UpdateOrderItemStatusRoutes(),
		&handlers.UpdateOrderItemStatusHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.ProcessPaymentRequest, handlers.ProcessPaymentResponse](
		ctx, router, handlers.ProcessPaymentRoutes(),
		&handlers.ProcessPaymentHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.ProcessRefundRequest, handlers.ProcessRefundResponse](
		ctx, router, handlers.ProcessRefundRoutes(),
		&handlers.ProcessRefundHandler{Service: service},
	)

	// Locations
	serverHttp.RegisterGET[handlers.ListLocationsRequest, handlers.ListLocationsResponse](
		ctx, router, handlers.ListLocationsRoutes(),
		&handlers.ListLocationsHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.CreateLocationRequest, handlers.CreateLocationResponse](
		ctx, router, handlers.ListLocationsRoutes(),
		&handlers.CreateLocationHandler{Service: service},
	)

	// Dashboard
	serverHttp.RegisterGET[handlers.DashboardRequest, handlers.DashboardResponse](
		ctx, router, handlers.DashboardRoutes(),
		&handlers.DashboardHandler{Service: service},
	)

	// Order extensions
	serverHttp.RegisterPOSTWithValidation[handlers.SplitOrderRequest, handlers.SplitOrderResponse](
		ctx, router, handlers.SplitOrderRoutes(),
		&handlers.SplitOrderHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.DiscountOrderRequest, handlers.DiscountOrderResponse](
		ctx, router, handlers.DiscountOrderRoutes(),
		&handlers.DiscountOrderHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.VoidOrderRequest, handlers.VoidOrderResponse](
		ctx, router, handlers.VoidOrderRoutes(),
		&handlers.VoidOrderHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.HoldOrderRequest, handlers.HoldOrderResponse](
		ctx, router, handlers.HoldOrderRoutes(),
		&handlers.HoldOrderHandler{Service: service},
	)

	serverHttp.RegisterGET[handlers.ListHeldOrdersRequest, handlers.ListHeldOrdersResponse](
		ctx, router, handlers.ListHeldOrdersRoutes(),
		&handlers.ListHeldOrdersHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.TransferTableRequest, handlers.TransferTableResponse](
		ctx, router, handlers.TransferTableRoutes(),
		&handlers.TransferTableHandler{Service: service},
	)

	serverHttp.RegisterDELETEWithValidation[handlers.VoidOrderItemRequest, handlers.VoidOrderItemResponse](
		ctx, router, handlers.VoidOrderItemRoutes(),
		&handlers.VoidOrderItemHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.SendToKitchenRequest, handlers.SendToKitchenResponse](
		ctx, router, handlers.SendToKitchenRoutes(),
		&handlers.SendToKitchenHandler{Service: service},
	)

	// Paddle billing
	serverHttp.RegisterPOSTWithValidation[handlers.PayOrderRequest, handlers.PayOrderResponse](
		ctx, router, handlers.PayOrderRoutes(),
		&handlers.PayOrderHandler{Service: service},
	)

	// Paddle webhook (anonymous, raw body for HMAC verification)
	router.Post("/pos/payments/webhook", handlers.PaddlePOSWebhookHTTPHandler(service))

	// Public online ordering (anonymous)
	serverHttp.RegisterPOSTWithValidation[handlers.CreatePublicOrderRequest, handlers.CreatePublicOrderResponse](
		ctx, router, handlers.CreatePublicOrderRoutes(),
		&handlers.CreatePublicOrderHandler{Service: service},
	)

	serverHttp.RegisterGET[handlers.GetPublicOrderStatusRequest, handlers.GetPublicOrderStatusResponse](
		ctx, router, handlers.GetPublicOrderStatusRoutes(),
		&handlers.GetPublicOrderStatusHandler{Service: service},
	)

	// POS WebSocket realtime endpoint (anonymous at middleware layer; token validated in handler)
	router.Get("/pos/ws", handlers.ServePOSWebSocket)
}
