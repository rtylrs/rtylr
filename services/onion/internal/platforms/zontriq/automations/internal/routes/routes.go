package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	automationsv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/automations/v1"
	"github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platforms/zontriq/automations/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/automations/internal/routes/handlers"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, automationsService *application.Service) {
	http.RegisterGET[automationsv1.ListAutomationsRequest, automationsv1.ListAutomationsResponse](
		ctx,
		router,
		handlers.ListAutomationsRoutes(),
		&handlers.ListAutomationsHandler{AutomationsService: automationsService},
	)

	http.RegisterGET[automationsv1.GetAutomationRequest, automationsv1.GetAutomationResponse](
		ctx,
		router,
		handlers.GetAutomationRoutes(),
		&handlers.GetAutomationHandler{AutomationsService: automationsService},
	)

	http.RegisterPOSTWithValidation[automationsv1.CreateAutomationRequest, automationsv1.CreateAutomationResponse](
		ctx,
		router,
		handlers.CreateAutomationRoutes(),
		&handlers.CreateAutomationHandler{AutomationsService: automationsService},
	)

	http.RegisterPOSTWithValidation[automationsv1.UpdateAutomationRequest, automationsv1.UpdateAutomationResponse](
		ctx,
		router,
		handlers.UpdateAutomationRoutes(),
		&handlers.UpdateAutomationHandler{AutomationsService: automationsService},
	)

	http.RegisterDELETE[automationsv1.DeleteAutomationRequest, automationsv1.DeleteAutomationResponse](
		ctx,
		router,
		handlers.DeleteAutomationRoutes(),
		&handlers.DeleteAutomationHandler{AutomationsService: automationsService},
	)
}
