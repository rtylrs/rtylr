package api

import (
	"context"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/websitelead/internal/application"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterPOSTWithValidationAndMiddleware[SubmitRequest, SubmitResponse](
		ctx, router, SubmitRoutes(), &SubmitHandler{Service: service},
		http.PerRouteRateLimit(10, time.Minute),
	)
	http.RegisterPOSTWithValidationAndMiddleware[ContactRequest, ContactResponse](
		ctx, router, ContactRoutes(), &ContactHandler{Service: service},
		http.PerRouteRateLimit(10, time.Minute),
	)
	http.RegisterPOSTWithValidationAndMiddleware[NewsletterRequest, NewsletterResponse](
		ctx, router, NewsletterRoutes(), &NewsletterHandler{Service: service},
		http.PerRouteRateLimit(10, time.Minute),
	)
	http.RegisterPOSTWithValidationAndMiddleware[AnalyticsEventRequest, AnalyticsEventResponse](
		ctx, router, AnalyticsEventRoutes(), &AnalyticsEventHandler{},
		http.PerRouteRateLimit(60, time.Minute),
	)
}
