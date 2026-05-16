package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/departments/internal/application"
	"github.com/voxire/onion/internal/modules/departments/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListDepartmentsRequest, handlers.ListDepartmentsResponse](
		ctx,
		router,
		handlers.ListDepartmentsRoutes(),
		&handlers.ListDepartmentsHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateDepartmentRequest, handlers.CreateDepartmentResponse](
		ctx,
		router,
		handlers.CreateDepartmentRoutes(),
		&handlers.CreateDepartmentHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateDepartmentRequest, handlers.UpdateDepartmentResponse](
		ctx,
		router,
		handlers.UpdateDepartmentRoutes(),
		&handlers.UpdateDepartmentHandler{Service: service},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteDepartmentRequest, handlers.DeleteDepartmentResponse](
		ctx,
		router,
		handlers.DeleteDepartmentRoutes(),
		&handlers.DeleteDepartmentHandler{Service: service},
	)
}
