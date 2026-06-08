package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/campaigns/internal/application"
	"github.com/voxire/onion/internal/modules/campaigns/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListCampaignsRequest, handlers.ListCampaignsResponse](
		ctx, router, handlers.CampaignListRoutes(),
		&handlers.ListCampaignsHandler{Service: svc},
	)

	http.RegisterGET[handlers.GetCampaignRequest, handlers.GetCampaignResponse](
		ctx, router, handlers.CampaignGetRoutes(),
		&handlers.GetCampaignHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateCampaignRequest, handlers.CreateCampaignResponse](
		ctx, router, handlers.CampaignCreateRoutes(),
		&handlers.CreateCampaignHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateCampaignRequest, handlers.UpdateCampaignResponse](
		ctx, router, handlers.CampaignUpdateRoutes(),
		&handlers.UpdateCampaignHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateCampaignStatusRequest, handlers.UpdateCampaignStatusResponse](
		ctx, router, handlers.CampaignStatusRoutes(),
		&handlers.UpdateCampaignStatusHandler{Service: svc},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteCampaignRequest, handlers.DeleteCampaignResponse](
		ctx, router, handlers.CampaignDeleteRoutes(),
		&handlers.DeleteCampaignHandler{Service: svc},
	)
}
