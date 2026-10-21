package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	posordersdiscountsv1 "github.com/voxire/voxie/gen/go/pos/orders/discounts/v1"
	posordersrefundsv1 "github.com/voxire/voxie/gen/go/pos/orders/refunds/v1"
	posordersv1 "github.com/voxire/voxie/gen/go/pos/orders/v1"
)

func posOrdersUser(permissions ...string) *auth.AuthUser {
	permissionScopes := make(map[string][]authorization.Scope, len(permissions))
	for _, permission := range permissions {
		permissionScopes[permission] = []authorization.Scope{{Type: "pos_location", ID: "3001"}}
	}
	return &auth.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:            authorization.AppPOS,
				Licensed:         true,
				HasAccess:        true,
				Permissions:      permissions,
				PermissionScopes: permissionScopes,
			}},
		},
	}
}

func posOrdersStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestPOSOrderOperationsRequireExactPermissionBeforeServiceAccess(t *testing.T) {
	cases := []struct {
		name       string
		permission string
		call       func(context.Context) error
	}{
		{"list", "pos:orders:view", func(ctx context.Context) error {
			_, err := (&ListOrdersHandler{}).Serve(ctx, &ListOrdersRequest{})
			return err
		}},
		{"create", "pos:orders:create", func(ctx context.Context) error {
			_, err := (&CreateOrderHandler{}).Serve(ctx, &CreateOrderRequest{})
			return err
		}},
		{"canonical list", "pos:orders:view", func(ctx context.Context) error {
			_, err := (&OrderAPI{}).ListOrders(ctx, &posordersv1.ListOrdersRequest{})
			return err
		}},
		{"canonical create", "pos:orders:create", func(ctx context.Context) error {
			_, err := (&OrderAPI{}).CreateOrder(ctx, &posordersv1.CreateOrderRequest{
				Items: []*posordersv1.CreateOrderItemInput{{MenuItemId: 1, Name: "item", Quantity: 1}},
			})
			return err
		}},
		{"void", "pos:orders:void", func(ctx context.Context) error {
			_, err := (&VoidOrderHandler{}).Serve(ctx, &VoidOrderRequest{})
			return err
		}},
		{"refund", "pos:orders:refund", func(ctx context.Context) error {
			_, err := (&RefundAPI{}).ProcessRefund(ctx, &posordersrefundsv1.ProcessRefundRequest{})
			return err
		}},
		{"discount", "pos:orders:discount", func(ctx context.Context) error {
			_, err := (&DiscountAPI{}).DiscountOrder(ctx, &posordersdiscountsv1.DiscountOrderRequest{})
			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := auth.WithAuthUser(context.Background(), posOrdersUser())
			if err := tc.call(ctx); posOrdersStatus(err) != http.StatusForbidden {
				t.Fatalf("permission %s returned %v, want 403 before service access", tc.permission, err)
			}
		})
	}
}

func TestPOSOrderPermissionHelperRejectsWrongAppAndMembersWithoutAccess(t *testing.T) {
	wrongApp := &auth.AuthUser{BusinessID: 9, UserID: 7, EffectiveAccess: &authorization.EffectiveAccess{
		Apps: []authorization.EffectiveAppAccess{{AppID: authorization.AppFinance, HasAccess: true, Permissions: []string{"pos:orders:view"}}},
	}}
	if _, err := requirePOSPermission(auth.WithAuthUser(context.Background(), wrongApp), "pos:orders:view"); posOrdersStatus(err) != http.StatusForbidden {
		t.Fatalf("wrong app returned %v, want 403", err)
	}
	noAccess := posOrdersUser("pos:orders:view")
	noAccess.EffectiveAccess.Apps[0].HasAccess = false
	if _, err := requirePOSPermission(auth.WithAuthUser(context.Background(), noAccess), "pos:orders:view"); posOrdersStatus(err) != http.StatusForbidden {
		t.Fatalf("app without access returned %v, want 403", err)
	}
	// A permission assignment cannot replace the app's active license.
	unlicensed := posOrdersUser("pos:orders:view")
	unlicensed.EffectiveAccess.Apps[0].Licensed = false
	if _, err := requirePOSPermission(auth.WithAuthUser(context.Background(), unlicensed), "pos:orders:view"); posOrdersStatus(err) != http.StatusForbidden {
		t.Fatalf("an unlicensed app returned %v, want 403", err)
	}
}
