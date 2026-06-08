package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/automations/internal/application"
	"github.com/voxire/onion/internal/modules/automations/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListAutomationsRequest, handlers.ListAutomationsResponse](
		ctx,
		router,
		handlers.AutomationListRoutes(),
		&handlers.ListAutomationsHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateAutomationRequest, handlers.CreateAutomationResponse](
		ctx,
		router,
		handlers.AutomationCreateRoutes(),
		&handlers.CreateAutomationHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateAutomationRequest, handlers.UpdateAutomationResponse](
		ctx,
		router,
		handlers.AutomationUpdateRoutes(),
		&handlers.UpdateAutomationHandler{Service: service},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteAutomationRequest, handlers.DeleteAutomationResponse](
		ctx,
		router,
		handlers.AutomationDeleteRoutes(),
		&handlers.DeleteAutomationHandler{Service: service},
	)
}
