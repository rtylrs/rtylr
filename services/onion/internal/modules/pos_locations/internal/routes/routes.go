package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/pos_locations/internal/application"
	"github.com/voxire/onion/internal/modules/pos_locations/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListLocationsRequest, handlers.ListLocationsResponse](
		ctx, router, handlers.LocationListRoutes(),
		&handlers.ListLocationsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateLocationRequest, handlers.CreateLocationResponse](
		ctx, router, handlers.LocationCreateRoutes(),
		&handlers.CreateLocationHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateLocationRequest, handlers.UpdateLocationResponse](
		ctx, router, handlers.LocationUpdateRoutes(),
		&handlers.UpdateLocationHandler{Service: svc},
	)
}
