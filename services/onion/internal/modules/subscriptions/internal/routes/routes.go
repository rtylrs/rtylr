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
	serverHttp.RegisterPOSTWithValidation[handlers.UpsertSubscriptionRequest, handlers.UpsertSubscriptionResponse](
		ctx, router, handlers.UpsertSubscriptionRoutes(),
		&handlers.UpsertSubscriptionHandler{Service: service},
	)
}
