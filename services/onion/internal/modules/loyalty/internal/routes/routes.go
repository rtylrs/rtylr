package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/loyalty/internal/application"
	"github.com/voxire/onion/internal/modules/loyalty/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	// Customer loyalty config, tiers, transactions, award, and redeem routes
	// are owned by the CRM customers module. Keeping one registration avoids
	// order-dependent handlers with different response contracts.
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
