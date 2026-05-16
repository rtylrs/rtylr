package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/support_canned/internal/application"
	"github.com/voxire/onion/internal/modules/support_canned/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListCannedRequest, handlers.ListCannedResponse](
		ctx, router, handlers.ListCannedRoutes(),
		&handlers.ListCannedHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateCannedRequest, handlers.CreateCannedResponse](
		ctx, router, handlers.CreateCannedRoutes(),
		&handlers.CreateCannedHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateCannedRequest, handlers.UpdateCannedResponse](
		ctx, router, handlers.UpdateCannedRoutes(),
		&handlers.UpdateCannedHandler{Service: service},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteCannedRequest, handlers.DeleteCannedResponse](
		ctx, router, handlers.DeleteCannedRoutes(),
		&handlers.DeleteCannedHandler{Service: service},
	)
}
