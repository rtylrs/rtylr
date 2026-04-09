package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/staff/internal/application"
	"github.com/voxire/onion/internal/modules/staff/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListStaffRequest, handlers.ListStaffResponse](
		ctx,
		router,
		handlers.ListStaffRoutes(),
		&handlers.ListStaffHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateStaffRequest, handlers.CreateStaffResponse](
		ctx,
		router,
		handlers.CreateStaffRoutes(),
		&handlers.CreateStaffHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateStaffRequest, handlers.UpdateStaffResponse](
		ctx,
		router,
		handlers.UpdateStaffRoutes(),
		&handlers.UpdateStaffHandler{Service: service},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteStaffRequest, handlers.DeleteStaffResponse](
		ctx,
		router,
		handlers.DeleteStaffRoutes(),
		&handlers.DeleteStaffHandler{Service: service},
	)

	http.RegisterGET[handlers.DashboardRequest, handlers.DashboardResponse](
		ctx,
		router,
		handlers.DashboardRoutes(),
		&handlers.DashboardHandler{Service: service},
	)
}
