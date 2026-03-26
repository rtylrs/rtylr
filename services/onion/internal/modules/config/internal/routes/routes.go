package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/config/internal/application"
	"github.com/voxire/onion/internal/modules/config/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, configService *application.Service) {
	http.RegisterGET[handlers.GetFeaturesRequest, handlers.GetFeaturesResponse](
		ctx,
		router,
		handlers.GetFeaturesRoutes(),
		&handlers.GetFeaturesHandler{ConfigService: configService},
	)

	http.RegisterPOSTWithValidation[handlers.GetConfigByOwnerRequest, handlers.GetConfigByOwnerResponse](
		ctx,
		router,
		handlers.ConfigByOwnerRoutes(),
		&handlers.ConfigByOwnerHandler{ConfigService: configService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateConfigRequest, handlers.CreateConfigResponse](
		ctx,
		router,
		handlers.CreateConfigRoutes(),
		&handlers.CreateConfigHandler{ConfigService: configService},
	)
}
