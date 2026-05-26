package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/booking_calendar/internal/application"
	"github.com/voxire/onion/internal/modules/booking_calendar/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGETWithValidation[handlers.CalendarRequest, handlers.CalendarResponse](ctx, router, handlers.CalendarRoutes(), &handlers.CalendarHandler{Service: service})
}
