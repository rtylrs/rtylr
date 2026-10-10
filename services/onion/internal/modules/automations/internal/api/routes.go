package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/automations/internal/application"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router http.LegacyRouter, service *application.Service) {
	http.RegisterGET[ListAutomationsRequest, ListAutomationsResponse](
		ctx,
		router,
		AutomationListRoutes(),
		&ListAutomationsHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[CreateAutomationRequest, CreateAutomationResponse](
		ctx,
		router,
		AutomationCreateRoutes(),
		&CreateAutomationHandler{Service: service},
	)

	http.RegisterPUTWithValidation[UpdateAutomationRequest, UpdateAutomationResponse](
		ctx,
		router,
		AutomationUpdateRoutes(),
		&UpdateAutomationHandler{Service: service},
	)

	http.RegisterDELETEWithValidation[DeleteAutomationRequest, DeleteAutomationResponse](
		ctx,
		router,
		AutomationDeleteRoutes(),
		&DeleteAutomationHandler{Service: service},
	)
}
