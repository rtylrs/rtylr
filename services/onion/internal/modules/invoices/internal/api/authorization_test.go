package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func financeUser(permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppFinance,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
		},
	}
}

func financePermissionStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestListInvoicesRequiresFeaturePermission(t *testing.T) {
	h := &ListInvoicesHandler{}
	ctx := authPkg.WithAuthUser(context.Background(), financeUser())

	_, err := h.Serve(ctx, &ListInvoicesRequest{})
	if got := financePermissionStatus(err); got != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", got, http.StatusForbidden)
	}
}

func TestRequireFinancePermissionAllowsGrantedPermission(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), financeUser("finance:invoices:view"))

	if err := requireFinancePermission(ctx, "finance:invoices:view"); err != nil {
		t.Fatalf("expected granted permission to pass, got %v", err)
	}
}
