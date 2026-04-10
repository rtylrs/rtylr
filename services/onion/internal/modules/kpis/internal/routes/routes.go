package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/kpis/internal/application"
	"github.com/voxire/onion/internal/modules/kpis/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.KPIRequest, handlers.KPIResponse](
		ctx, router, handlers.KPIRoutes(),
		&handlers.KPIHandler{Service: service},
	)
}
