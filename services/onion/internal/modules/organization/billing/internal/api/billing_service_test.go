package api

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

// The billing summary belongs to the business; members without billing access
// must not read it.
func TestBillingSummaryRequiresBillingAccess(t *testing.T) {
	api := NewBillingAPI(nil)
	member := &authPkg.AuthUser{UserID: 4, BusinessID: 10, EffectiveAccess: &authorization.EffectiveAccess{DashPermissions: []string{"hub:team:view"}}}
	ctx := authPkg.WithAuthUser(context.Background(), member)
	if _, err := api.GetBillingSummary(ctx, nil); err == nil {
		t.Error("GetBillingSummary served a member without billing access")
	}
	if _, err := api.ListInvoices(ctx, nil); err == nil {
		t.Error("ListInvoices served a member without billing access")
	}
}
