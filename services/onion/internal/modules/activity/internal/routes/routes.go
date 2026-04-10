package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/activity/internal/application"
	"github.com/voxire/onion/internal/modules/activity/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListActivityRequest, handlers.ListActivityResponse](
		ctx,
		router,
		handlers.ListActivityRoutes(),
		&handlers.ListActivityHandler{Service: service},
	)
}
