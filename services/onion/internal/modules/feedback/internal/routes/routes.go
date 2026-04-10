package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/feedback/internal/application"
	"github.com/voxire/onion/internal/modules/feedback/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListFeedbackRequest, handlers.ListFeedbackResponse](
		ctx,
		router,
		handlers.FeedbackListRoutes(),
		&handlers.ListFeedbackHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateFeedbackRequest, handlers.CreateFeedbackResponse](
		ctx,
		router,
		handlers.FeedbackCreateRoutes(),
		&handlers.CreateFeedbackHandler{Service: service},
	)
}
