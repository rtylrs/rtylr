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

	http.RegisterGETWithValidation[handlers.GetStaffRequest, handlers.GetStaffResponse](
		ctx,
		router,
		handlers.GetStaffRoutes(),
		&handlers.GetStaffHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.ClockInRequest, handlers.ClockInResponse](
		ctx,
		router,
		handlers.ClockInRoutes(),
		&handlers.ClockInHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.ClockOutRequest, handlers.ClockOutResponse](
		ctx,
		router,
		handlers.ClockOutRoutes(),
		&handlers.ClockOutHandler{Service: service},
	)

	http.RegisterGET[handlers.DashboardRequest, handlers.DashboardResponse](
		ctx,
		router,
		handlers.DashboardRoutes(),
		&handlers.DashboardHandler{Service: service},
	)
}
