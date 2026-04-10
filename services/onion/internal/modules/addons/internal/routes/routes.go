package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/addons/internal/application"
	"github.com/voxire/onion/internal/modules/addons/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.ListAddonsRequest, handlers.ListAddonsResponse](
		ctx, router, handlers.ListAddonsRoutes(),
		&handlers.ListAddonsHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.EnableAddonRequest, handlers.EnableAddonResponse](
		ctx, router, handlers.EnableAddonRoutes(),
		&handlers.EnableAddonHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.DisableAddonRequest, handlers.DisableAddonResponse](
		ctx, router, handlers.DisableAddonRoutes(),
		&handlers.DisableAddonHandler{Service: service},
	)

	serverHttp.RegisterGET[handlers.AddonStatusRequest, handlers.AddonStatusResponse](
		ctx, router, handlers.AddonStatusRoutes(),
		&handlers.AddonStatusHandler{Service: service},
	)
}
