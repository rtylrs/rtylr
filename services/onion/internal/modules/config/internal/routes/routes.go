package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/config/internal/application"
	"github.com/voxire/onion/internal/modules/config/internal/routes/handlers"
	configv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/shared/config/v1"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, configService *application.Service) {
	http.RegisterGET[configv1.GetFeaturesRequest, configv1.GetFeaturesResponse](
		ctx,
		router,
		handlers.GetFeaturesRoutes(),
		&handlers.GetFeaturesHandler{ConfigService: configService},
	)

	http.RegisterPOSTWithValidation[configv1.GetConfigByOwnerRequest, configv1.GetConfigByOwnerResponse](
		ctx,
		router,
		handlers.ConfigByOwnerRoutes(),
		&handlers.ConfigByOwnerHandler{ConfigService: configService},
	)

	http.RegisterPOSTWithValidation[configv1.CreateConfigRequest, configv1.CreateConfigResponse](
		ctx,
		router,
		handlers.CreateConfigRoutes(),
		&handlers.CreateConfigHandler{ConfigService: configService},
	)
}
