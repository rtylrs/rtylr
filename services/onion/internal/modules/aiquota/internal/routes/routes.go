package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/aiquota/internal/application"
	"github.com/voxire/onion/internal/modules/aiquota/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, r *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.UsageRequest, handlers.UsageResponse](
		ctx, r, handlers.UsageRoutes(),
		&handlers.UsageHandler{Service: service},
	)

	serverHttp.RegisterGET[handlers.GetQuotaSettingsRequest, handlers.QuotaSettingsResponse](
		ctx, r, handlers.GetQuotaSettingsRoutes(),
		&handlers.GetQuotaSettingsHandler{Service: service},
	)

	serverHttp.RegisterPUTWithValidation[handlers.UpdateQuotaSettingsRequest, handlers.QuotaSettingsResponse](
		ctx, r, handlers.UpdateQuotaSettingsRoutes(),
		&handlers.UpdateQuotaSettingsHandler{Service: service},
	)
}
