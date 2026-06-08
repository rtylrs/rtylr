package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/config/internal/application"
	"github.com/voxire/onion/internal/modules/config/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, configService *application.Service) {
	http.RegisterPOSTWithValidation[handlers.GetConfigByOwnerRequest, handlers.GetConfigByOwnerResponse](
		ctx,
		router,
		handlers.ConfigByOwnerRoutes(),
		&handlers.GetConfigByOwnerHandler{ConfigService: configService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateConfigRequest, handlers.CreateConfigResponse](
		ctx,
		router,
		handlers.CreateConfigRoutes(),
		&handlers.CreateConfigHandler{ConfigService: configService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateConfigRequest, handlers.UpdateConfigResponse](
		ctx,
		router,
		handlers.UpdateConfigRoutes(),
		&handlers.UpdateConfigHandler{ConfigService: configService},
	)

	http.RegisterGET[handlers.GetIndustriesRequest, handlers.GetIndustriesResponse](
		ctx,
		router,
		handlers.IndustriesRoutes(),
		&handlers.GetIndustriesHandler{},
	)
}
