package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_reports/internal/application"
	"github.com/voxire/onion/internal/modules/hr_reports/internal/routes/handlers"
)

func RegisterRoutes(_ context.Context, router *chi.Mux, service *application.Service, permissions handlers.PermissionChecker) {
	handler := &handlers.GenerateReportHandler{Service: service, Permissions: permissions}
	for _, path := range handlers.GenerateReportRoutes() {
		// The handler supports both JSON and CSV responses and therefore owns
		// response content negotiation in addition to request validation.
		router.Post(path, handler.ServeHTTP)
	}
}
