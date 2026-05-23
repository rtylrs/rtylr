package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/time_off/internal/application"
	"github.com/voxire/onion/internal/modules/time_off/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListTimeOffRequest, handlers.ListTimeOffResponse](
		ctx,
		router,
		handlers.ListTimeOffRoutes(),
		&handlers.ListTimeOffHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.SubmitTimeOffRequest, handlers.SubmitTimeOffResponse](
		ctx,
		router,
		handlers.SubmitTimeOffRoutes(),
		&handlers.SubmitTimeOffHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.ApproveTimeOffRequest, handlers.ApproveTimeOffResponse](
		ctx,
		router,
		handlers.ApproveTimeOffRoutes(),
		&handlers.ApproveTimeOffHandler{Service: service},
	)
}
