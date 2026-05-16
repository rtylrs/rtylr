package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/subscriptions/internal/application"
	"github.com/voxire/onion/internal/modules/subscriptions/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.ListPlansRequest, handlers.ListPlansResponse](
		ctx, router, handlers.ListPlansRoutes(),
		&handlers.ListPlansHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.MySubscriptionRequest, handlers.MySubscriptionResponse](
		ctx, router, handlers.MySubscriptionRoutes(),
		&handlers.MySubscriptionHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.BillingDetailRequest, handlers.BillingDetailResponse](
		ctx, router, handlers.BillingDetailRoutes(),
		&handlers.BillingDetailHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.BillingHistoryRequest, handlers.BillingHistoryResponse](
		ctx, router, handlers.BillingHistoryRoutes(),
		&handlers.BillingHistoryHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.UpsertSubscriptionRequest, handlers.UpsertSubscriptionResponse](
		ctx, router, handlers.UpsertSubscriptionRoutes(),
		&handlers.UpsertSubscriptionHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CheckoutRequest, handlers.CheckoutResponse](
		ctx, router, handlers.CheckoutRoutes(),
		&handlers.CheckoutHandler{Service: service},
	)

	// Webhook uses a raw handler — body must be read before any JSON decoding for HMAC verification.
	router.Post("/subscriptions/paddle-webhook", handlers.PaddleWebhookHTTPHandler(service))
}
