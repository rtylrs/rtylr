package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	poscashv1 "github.com/voxire/voxie/gen/go/pos/cash/v1"
)

func cashStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestCashOperationsRequireExactPermissionBeforeServiceAccess(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:     authorization.AppPOS,
				Licensed:  true,
				HasAccess: true,
			}},
		},
	})
	api := &CashAPI{}
	cases := []func() error{
		func() error { _, err := api.ListCashMovements(ctx, &poscashv1.ListCashMovementsRequest{}); return err },
		func() error { _, err := api.CashDrop(ctx, &poscashv1.CashDropRequest{}); return err },
		func() error { _, err := api.CashPickup(ctx, &poscashv1.CashPickupRequest{}); return err },
		func() error { _, err := api.CashAdjust(ctx, &poscashv1.CashAdjustRequest{}); return err },
	}
	for i, call := range cases {
		if err := call(); cashStatus(err) != http.StatusForbidden {
			t.Fatalf("case %d returned %v, want 403 before service access", i, err)
		}
	}
}
