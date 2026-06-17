package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_settings/internal/application"
	"github.com/voxire/onion/internal/modules/pos_settings/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.GetSettingsRequest, handlers.GetSettingsResponse](
		ctx, router, handlers.GetSettingsRoutes(),
		&handlers.GetSettingsHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateSettingsRequest, handlers.UpdateSettingsResponse](
		ctx, router, handlers.UpdateSettingsRoutes(),
		&handlers.UpdateSettingsHandler{Service: svc},
	)
}
