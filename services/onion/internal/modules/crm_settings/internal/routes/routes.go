package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/crm_settings/internal/application"
	"github.com/voxire/onion/internal/modules/crm_settings/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.GetSettingsRequest, handlers.GetSettingsResponse](
		ctx,
		router,
		handlers.SettingsGetRoutes(),
		&handlers.GetSettingsHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateSettingsRequest, handlers.UpdateSettingsResponse](
		ctx,
		router,
		handlers.SettingsUpdateRoutes(),
		&handlers.UpdateSettingsHandler{Service: service},
	)
}
