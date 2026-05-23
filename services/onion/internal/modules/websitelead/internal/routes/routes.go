package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/websitelead/internal/application"
	"github.com/voxire/onion/internal/modules/websitelead/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterPOSTWithValidation[handlers.SubmitRequest, handlers.SubmitResponse](
		ctx, router, handlers.SubmitRoutes(), &handlers.SubmitHandler{Service: service},
	)
	http.RegisterPOSTWithValidation[handlers.ContactRequest, handlers.ContactResponse](
		ctx, router, handlers.ContactRoutes(), &handlers.ContactHandler{Service: service},
	)
	http.RegisterPOST[handlers.AnalyticsEventRequest, handlers.AnalyticsEventResponse](
		ctx, router, handlers.AnalyticsEventRoutes(), &handlers.AnalyticsEventHandler{},
	)
}
