package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.ListTemplatesRequest, handlers.ListTemplatesResponse](
		ctx, router, handlers.ListTemplatesRoutes(), &handlers.ListTemplatesHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateTemplateRequest, handlers.CreateTemplateResponse](
		ctx, router, handlers.CreateTemplateRoutes(), &handlers.CreateTemplateHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateTemplateRequest, handlers.UpdateTemplateResponse](
		ctx, router, handlers.UpdateTemplateRoutes(), &handlers.UpdateTemplateHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.AddItemToTemplateRequest, handlers.AddItemToTemplateResponse](
		ctx, router, handlers.AddItemToTemplateRoutes(), &handlers.AddItemToTemplateHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.TriggerChecklistRequest, handlers.TriggerChecklistResponse](
		ctx, router, handlers.TriggerChecklistRoutes(), &handlers.TriggerChecklistHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.ListEmployeeChecklistsRequest, handlers.ListEmployeeChecklistsResponse](
		ctx, router, handlers.ListEmployeeChecklistsRoutes(), &handlers.ListEmployeeChecklistsHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.GetEmployeeChecklistRequest, handlers.GetEmployeeChecklistResponse](
		ctx, router, handlers.GetEmployeeChecklistRoutes(), &handlers.GetEmployeeChecklistHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CompleteChecklistItemRequest, handlers.CompleteChecklistItemResponse](
		ctx, router, handlers.CompleteChecklistItemRoutes(), &handlers.CompleteChecklistItemHandler{Service: service},
	)
}
