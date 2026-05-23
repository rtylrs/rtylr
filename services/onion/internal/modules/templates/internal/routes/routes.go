package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/templates/internal/application"
	"github.com/voxire/onion/internal/modules/templates/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListTemplatesRequest, handlers.ListTemplatesResponse](
		ctx,
		router,
		handlers.TemplateListRoutes(),
		&handlers.ListTemplatesHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateTemplateRequest, handlers.CreateTemplateResponse](
		ctx,
		router,
		handlers.TemplateCreateRoutes(),
		&handlers.CreateTemplateHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateTemplateRequest, handlers.UpdateTemplateResponse](
		ctx,
		router,
		handlers.TemplateUpdateRoutes(),
		&handlers.UpdateTemplateHandler{Service: service},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteTemplateRequest, handlers.DeleteTemplateResponse](
		ctx,
		router,
		handlers.TemplateDeleteRoutes(),
		&handlers.DeleteTemplateHandler{Service: service},
	)
}
