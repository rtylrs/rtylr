package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/shifts/internal/application"
	"github.com/voxire/onion/internal/modules/shifts/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListShiftsRequest, handlers.ListShiftsResponse](
		ctx,
		router,
		handlers.ListShiftsRoutes(),
		&handlers.ListShiftsHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateShiftRequest, handlers.CreateShiftResponse](
		ctx,
		router,
		handlers.CreateShiftRoutes(),
		&handlers.CreateShiftHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateShiftRequest, handlers.UpdateShiftResponse](
		ctx,
		router,
		handlers.UpdateShiftRoutes(),
		&handlers.UpdateShiftHandler{Service: service},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteShiftRequest, handlers.DeleteShiftResponse](
		ctx,
		router,
		handlers.DeleteShiftRoutes(),
		&handlers.DeleteShiftHandler{Service: service},
	)
}
