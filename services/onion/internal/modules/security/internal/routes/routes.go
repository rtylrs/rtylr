package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/security/internal/application"
	"github.com/voxire/onion/internal/modules/security/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.GetSettingsRequest, handlers.GetSettingsResponse](
		ctx, router, handlers.SecurityRoutes(),
		&handlers.GetSettingsHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.Toggle2FARequest, handlers.Toggle2FAResponse](
		ctx, router, handlers.TwoFARoutes(),
		&handlers.Toggle2FAHandler{Service: service},
	)

	serverHttp.RegisterGET[handlers.LoginHistoryRequest, handlers.LoginHistoryResponse](
		ctx, router, handlers.LoginHistoryRoutes(),
		&handlers.LoginHistoryHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.InvalidateSessionsRequest, handlers.InvalidateSessionsResponse](
		ctx, router, handlers.InvalidateRoutes(),
		&handlers.InvalidateSessionsHandler{Service: service},
	)
}
