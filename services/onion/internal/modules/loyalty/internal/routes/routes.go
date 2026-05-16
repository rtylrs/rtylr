package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/loyalty/internal/application"
	"github.com/voxire/onion/internal/modules/loyalty/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListTransactionsRequest, handlers.ListTransactionsResponse](
		ctx, router, handlers.LoyaltyTransactionsRoutes(),
		&handlers.ListTransactionsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.AwardPointsRequest, handlers.AwardPointsResponse](
		ctx, router, handlers.LoyaltyAwardRoutes(),
		&handlers.AwardPointsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.RedeemPointsRequest, handlers.RedeemPointsResponse](
		ctx, router, handlers.LoyaltyRedeemRoutes(),
		&handlers.RedeemPointsHandler{Service: svc},
	)

	http.RegisterGET[handlers.GetLoyaltyConfigRequest, handlers.GetLoyaltyConfigResponse](
		ctx, router, handlers.LoyaltyConfigGetRoutes(),
		&handlers.GetLoyaltyConfigHandler{Service: svc},
	)

	http.RegisterPUT[handlers.UpdateLoyaltyConfigRequest, handlers.UpdateLoyaltyConfigResponse](
		ctx, router, handlers.LoyaltyConfigUpdateRoutes(),
		&handlers.UpdateLoyaltyConfigHandler{Service: svc},
	)

	http.RegisterGET[handlers.ListMembersRequest, handlers.ListMembersResponse](
		ctx, router, handlers.LoyaltyMembersRoutes(),
		&handlers.ListMembersHandler{Service: svc},
	)

	http.RegisterGET[handlers.ListRewardsRequest, handlers.ListRewardsResponse](
		ctx, router, handlers.LoyaltyRewardsListRoutes(),
		&handlers.ListRewardsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.UpsertRewardRequest, handlers.UpsertRewardResponse](
		ctx, router, handlers.LoyaltyRewardsUpsertRoutes(),
		&handlers.UpsertRewardHandler{Service: svc},
	)
}
