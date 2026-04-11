package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/booking_dashboard/internal/application"
	"github.com/voxire/onion/internal/modules/booking_dashboard/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.DashboardRequest, handlers.DashboardResponse](ctx, router, handlers.DashboardRoutes(), &handlers.DashboardHandler{Service: service})
}
