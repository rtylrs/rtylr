package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/apps/internal/application"
	"github.com/voxire/onion/internal/modules/apps/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.ListAppsRequest, handlers.ListAppsResponse](
		ctx, router, handlers.AppRoutes(),
		&handlers.ListAppsHandler{Service: service},
	)
}
