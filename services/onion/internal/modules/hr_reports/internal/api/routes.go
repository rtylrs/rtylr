package api

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_reports/internal/application"
)

func RegisterRoutes(_ context.Context, router *chi.Mux, service *application.Service, permissions PermissionChecker) {
	handler := &GenerateReportHandler{Service: service, Permissions: permissions}
	for _, path := range GenerateReportRoutes() {
		// The handler supports both JSON and CSV responses and therefore owns
		// response content negotiation in addition to request validation.
		router.Post(path, handler.ServeHTTP)
	}
}
