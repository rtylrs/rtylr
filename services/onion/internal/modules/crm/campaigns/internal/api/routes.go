package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/crm/campaigns/internal/application"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router http.LegacyRouter, svc *application.Service) {
	http.RegisterGET[ListCampaignsRequest, ListCampaignsResponse](
		ctx, router, CampaignListRoutes(),
		&ListCampaignsHandler{Service: svc},
	)

	http.RegisterGET[GetCampaignRequest, GetCampaignResponse](
		ctx, router, CampaignGetRoutes(),
		&GetCampaignHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[CreateCampaignRequest, CreateCampaignResponse](
		ctx, router, CampaignCreateRoutes(),
		&CreateCampaignHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[UpdateCampaignRequest, UpdateCampaignResponse](
		ctx, router, CampaignUpdateRoutes(),
		&UpdateCampaignHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[UpdateCampaignStatusRequest, UpdateCampaignStatusResponse](
		ctx, router, CampaignStatusRoutes(),
		&UpdateCampaignStatusHandler{Service: svc},
	)

	http.RegisterDELETEWithValidation[DeleteCampaignRequest, DeleteCampaignResponse](
		ctx, router, CampaignDeleteRoutes(),
		&DeleteCampaignHandler{Service: svc},
	)
}
