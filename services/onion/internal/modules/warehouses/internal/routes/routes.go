package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/warehouses/internal/application"
	"github.com/voxire/onion/internal/modules/warehouses/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListWarehousesRequest, handlers.ListWarehousesResponse](
		ctx,
		router,
		handlers.ListWarehousesRoutes(),
		&handlers.ListWarehousesHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateWarehouseRequest, handlers.CreateWarehouseResponse](
		ctx,
		router,
		handlers.CreateWarehouseRoutes(),
		&handlers.CreateWarehouseHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateWarehouseRequest, handlers.UpdateWarehouseResponse](
		ctx,
		router,
		handlers.UpdateWarehouseRoutes(),
		&handlers.UpdateWarehouseHandler{Service: service},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteWarehouseRequest, handlers.DeleteWarehouseResponse](
		ctx,
		router,
		handlers.DeleteWarehouseRoutes(),
		&handlers.DeleteWarehouseHandler{Service: service},
	)
}
