package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/crm/loyalty/internal/application"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router http.LegacyRouter, svc *application.Service) {
	// Customer loyalty config, tiers, transactions, award, and redeem routes
	// are owned by the CRM customers module. Keeping one registration avoids
	// order-dependent handlers with different response contracts.
	http.RegisterGET[ListMembersRequest, ListMembersResponse](
		ctx, router, LoyaltyMembersRoutes(),
		&ListMembersHandler{Service: svc},
	)

	http.RegisterGET[ListRewardsRequest, ListRewardsResponse](
		ctx, router, LoyaltyRewardsListRoutes(),
		&ListRewardsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[UpsertRewardRequest, UpsertRewardResponse](
		ctx, router, LoyaltyRewardsUpsertRoutes(),
		&UpsertRewardHandler{Service: svc},
	)
}
