package routes

import (
	"context"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/websitelead/internal/application"
	"github.com/voxire/onion/internal/modules/websitelead/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterPOSTWithValidationAndMiddleware[handlers.SubmitRequest, handlers.SubmitResponse](
		ctx, router, handlers.SubmitRoutes(), &handlers.SubmitHandler{Service: service},
		http.PerRouteRateLimit(10, time.Minute),
	)
	http.RegisterPOSTWithValidationAndMiddleware[handlers.ContactRequest, handlers.ContactResponse](
		ctx, router, handlers.ContactRoutes(), &handlers.ContactHandler{Service: service},
		http.PerRouteRateLimit(10, time.Minute),
	)
	http.RegisterPOSTWithValidationAndMiddleware[handlers.AnalyticsEventRequest, handlers.AnalyticsEventResponse](
		ctx, router, handlers.AnalyticsEventRoutes(), &handlers.AnalyticsEventHandler{},
		http.PerRouteRateLimit(60, time.Minute),
	)
}
