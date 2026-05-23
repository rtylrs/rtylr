package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/webhooks/internal/application"
	"github.com/voxire/onion/internal/modules/webhooks/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.ListRequest, handlers.ListResponse](
		ctx, router, handlers.WebhookRoutes(),
		&handlers.ListHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.CreateRequest, handlers.CreateResponse](
		ctx, router, handlers.WebhookRoutes(),
		&handlers.CreateHandler{Service: service},
	)

	serverHttp.RegisterDELETEWithValidation[handlers.DeleteRequest, handlers.DeleteResponse](
		ctx, router, handlers.WebhookRoutes(),
		&handlers.DeleteHandler{Service: service},
	)
}
