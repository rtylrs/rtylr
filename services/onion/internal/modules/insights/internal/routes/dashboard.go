package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/modules/insights/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func registerDashboardRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	serverHttp.RegisterGET[handlers.GetDashboardRequest, handlers.GetDashboardResponse](
		ctx, router, handlers.GetDashboardRoutes(),
		&handlers.GetDashboardHandler{Service: svc},
	)
}
