package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_member_locations/internal/application"
	"github.com/voxire/onion/internal/modules/pos_member_locations/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.GetMyLocationsRequest, handlers.GetMyLocationsResponse](
		ctx, router, handlers.MyLocationsRoutes(), &handlers.GetMyLocationsHandler{Service: svc},
	)
	http.RegisterGET[handlers.GetLocationRosterRequest, handlers.GetLocationRosterResponse](
		ctx, router, handlers.LocationRosterRoutes(), &handlers.GetLocationRosterHandler{Service: svc},
	)
	http.RegisterPOSTWithValidation[handlers.AssignLocationRequest, handlers.AssignLocationResponse](
		ctx, router, handlers.AssignRoutes(), &handlers.AssignLocationHandler{Service: svc},
	)
	http.RegisterPOSTWithValidation[handlers.RemoveLocationRequest, handlers.RemoveLocationResponse](
		ctx, router, handlers.RemoveRoutes(), &handlers.RemoveLocationHandler{Service: svc},
	)
}
