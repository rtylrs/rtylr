package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/modules/careers/internal/application"
	"github.com/voxire/onion/internal/modules/careers/internal/routes/handlers"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, careersService *application.Service) {
	http.RegisterGET[handlers.GetActiveCareersRequest, handlers.GetActiveCareersResponse](
		ctx,
		router,
		handlers.CareersRoutes(),
		&handlers.CareersHandler{CareersService: careersService},
	)
}
