package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	careersv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/voxire/careers/v1"
	"github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platforms/voxire/careers/internal/application"
	"github.com/voxire/onion/internal/platforms/voxire/careers/internal/routes/handlers"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, careersService *application.Service) {
	http.RegisterGET[careersv1.GetActiveCareersRequest, careersv1.GetActiveCareersResponse](
		ctx,
		router,
		handlers.CareersRoutes(),
		&handlers.CareersHandler{CareersService: careersService},
	)
}
