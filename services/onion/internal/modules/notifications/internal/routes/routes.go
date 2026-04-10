package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/notifications/internal/application"
	"github.com/voxire/onion/internal/modules/notifications/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.ListRequest, handlers.ListResponse](
		ctx, router, handlers.NotificationRoutes(),
		&handlers.ListHandler{Service: service},
	)

	serverHttp.RegisterPUTWithValidation[handlers.MarkReadRequest, handlers.MarkReadResponse](
		ctx, router, handlers.NotificationRoutes(),
		&handlers.MarkReadHandler{Service: service},
	)

	serverHttp.RegisterPUT[handlers.MarkAllReadRequest, handlers.MarkAllReadResponse](
		ctx, router, handlers.ReadAllRoutes(),
		&handlers.MarkAllReadHandler{Service: service},
	)
}
