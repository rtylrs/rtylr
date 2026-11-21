package api

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func TestRequireCRMCustomerListPermission(t *testing.T) {
	financeCtx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID:     35,
		BusinessID: 10,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppFinance,
				Licensed:    true,
				HasAccess:   true,
				Permissions: []string{"finance:invoices:create"},
			}},
		},
	})
	if err := requireCRMCustomerListPermission(financeCtx); err != nil {
		t.Fatalf("expected finance invoice user to pass customer list permission check, got %v", err)
	}
	if err := requireCRMFeaturePermission(financeCtx, "crm:customers:create"); err == nil {
		t.Fatal("expected finance invoice user to be denied crm:customers:create")
	}

	hrOnlyCtx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID:     36,
		BusinessID: 10,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppHR,
				Licensed:    true,
				HasAccess:   true,
				Permissions: []string{"hr:employees:view"},
			}},
		},
	})
	if err := requireCRMCustomerListPermission(hrOnlyCtx); err == nil {
		t.Fatal("expected HR-only user to be denied customer list permission")
	}
}
