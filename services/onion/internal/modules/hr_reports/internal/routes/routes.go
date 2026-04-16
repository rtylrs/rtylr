package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/hr_reports/internal/application"
	"github.com/voxire/onion/internal/modules/hr_reports/internal/routes/handlers"
)

func RegisterRoutes(_ context.Context, router *chi.Mux, service *application.Service) {
	handler := &handlers.GenerateReportHandler{Service: service}
	for _, path := range handlers.GenerateReportRoutes() {
		router.Post(path, handler.ServeHTTP)
	}
}
