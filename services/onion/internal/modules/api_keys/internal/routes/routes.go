package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/api_keys/internal/application"
	"github.com/voxire/onion/internal/modules/api_keys/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.ListRequest, handlers.ListResponse](
		ctx, router, handlers.ListRoutes(),
		&handlers.ListHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.CreateRequest, handlers.CreateResponse](
		ctx, router, handlers.ListRoutes(),
		&handlers.CreateHandler{Service: service},
	)

	serverHttp.RegisterDELETEWithValidation[handlers.RevokeRequest, handlers.RevokeResponse](
		ctx, router, handlers.ListRoutes(),
		&handlers.RevokeHandler{Service: service},
	)
}
