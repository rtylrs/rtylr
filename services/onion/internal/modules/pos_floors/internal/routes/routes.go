package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_floors/internal/application"
	"github.com/voxire/onion/internal/modules/pos_floors/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListFloorsRequest, handlers.ListFloorsResponse](
		ctx, router, handlers.FloorListRoutes(),
		&handlers.ListFloorsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateFloorRequest, handlers.CreateFloorResponse](
		ctx, router, handlers.FloorCreateRoutes(),
		&handlers.CreateFloorHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateFloorRequest, handlers.UpdateFloorResponse](
		ctx, router, handlers.FloorUpdateRoutes(),
		&handlers.UpdateFloorHandler{Service: svc},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteFloorRequest, handlers.DeleteFloorResponse](
		ctx, router, handlers.FloorDeleteRoutes(),
		&handlers.DeleteFloorHandler{Service: svc},
	)
}
