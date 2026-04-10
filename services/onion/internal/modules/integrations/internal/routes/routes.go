package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/integrations/internal/application"
	"github.com/voxire/onion/internal/modules/integrations/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.ListRequest, handlers.ListResponse](
		ctx, router, handlers.IntegrationRoutes(),
		&handlers.ListHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.ConnectRequest, handlers.ConnectResponse](
		ctx, router, handlers.IntegrationRoutes(),
		&handlers.ConnectHandler{Service: service},
	)

	serverHttp.RegisterDELETEWithValidation[handlers.DisconnectRequest, handlers.DisconnectResponse](
		ctx, router, handlers.IntegrationRoutes(),
		&handlers.DisconnectHandler{Service: service},
	)
}
