package invoices

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func invoiceToolUser(permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		UserID:     9101,
		BusinessID: 9201,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID:           9201,
			AuthorizationVersion: 1,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppFinance,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
				Scopes:      []authorization.Scope{{Type: "finance", ID: "9201"}},
			}},
		},
	}
}

// The module maps tenant-resolution failure to the same error as a permission
// failure, so the gate is asserted directly rather than through the executor.
func TestRequireInvoiceToolPermission(t *testing.T) {
	t.Setenv("AUTHORIZATION_V2_MODE", "v2")

	for _, tc := range []struct {
		tool       string
		permission string
	}{
		{listInvoicesToolName, "finance:invoices:view"},
		{getInvoiceToolName, "finance:invoices:view"},
		{createInvoiceToolName, "finance:invoices:create"},
		{recordPaymentToolName, "finance:invoices:create"},
		{sendInvoiceToolName, "finance:invoices:send"},
	} {
		t.Run(tc.tool, func(t *testing.T) {
			required := invoiceToolPermissions[tc.tool]
			if required.permission != tc.permission {
				t.Fatalf("permission = %q, want %q", required.permission, tc.permission)
			}

			allowed := authPkg.WithAuthUser(context.Background(), invoiceToolUser("finance:invoices:view", tc.permission))
			if err := requireInvoiceToolPermission(allowed, required); err != nil {
				t.Fatalf("caller holding %s was denied: %v", tc.permission, err)
			}

			denied := authPkg.WithAuthUser(context.Background(), invoiceToolUser("finance:expenses:view"))
			if err := requireInvoiceToolPermission(denied, required); err == nil {
				t.Fatal("caller without the permission was allowed")
			} else if err.Error() != onekiterrors.PermissionDeniedMessage {
				t.Fatalf("error = %q, want %q", err.Error(), onekiterrors.PermissionDeniedMessage)
			}
		})
	}
}

func TestRequireInvoiceToolPermissionRejectsLegacyRoleAndMissingAuth(t *testing.T) {
	t.Setenv("AUTHORIZATION_V2_MODE", "v2")

	required := invoiceToolPermissions[sendInvoiceToolName]

	legacy := invoiceToolUser()
	legacy.Role = "owner"
	if err := requireInvoiceToolPermission(authPkg.WithAuthUser(context.Background(), legacy), required); err == nil {
		t.Fatal("a legacy owner role authorized an invoice send")
	}

	// Assignment, not licensing. HasAppPermission and HasAppAccess both gate on
	// HasAccess and neither reads EffectiveAppAccess.Licensed, so an unassigned
	// tenant is denied here exactly as the HTTP twin denies it. The Licensed flag
	// is surfaced to clients but enforced by no authorization gate in the
	// codebase, HTTP or AI; asserting it here would assert a product change
	// rather than this fix.
	unassigned := invoiceToolUser("finance:invoices:send")
	unassigned.EffectiveAccess.Apps[0].HasAccess = false
	if err := requireInvoiceToolPermission(authPkg.WithAuthUser(context.Background(), unassigned), required); err == nil {
		t.Fatal("an unassigned finance app authorized an invoice send")
	}

	suspended := invoiceToolUser("finance:invoices:send")
	suspended.ProvisioningStatus = &authorization.ProvisioningStatus{State: authorization.ProvisioningSuspended}
	if err := requireInvoiceToolPermission(authPkg.WithAuthUser(context.Background(), suspended), required); err == nil {
		t.Fatal("a suspended user authorized an invoice send")
	}

	if err := requireInvoiceToolPermission(context.Background(), required); err == nil {
		t.Fatal("an unauthenticated context was authorized")
	}
}

// Every tool the module dispatches must declare a permission, and no write tool
// may fall back to the read permission.
func TestInvoiceToolPermissionsCoverEveryDispatchedTool(t *testing.T) {
	dispatched := []string{
		listInvoicesToolName,
		getInvoiceToolName,
		createInvoiceToolName,
		recordPaymentToolName,
		sendInvoiceToolName,
	}
	for _, toolName := range dispatched {
		if _, ok := invoiceToolPermissions[toolName]; !ok {
			t.Errorf("%s is dispatched without a declared permission", toolName)
		}
	}
	if len(invoiceToolPermissions) != len(dispatched) {
		t.Errorf("declared %d permissions for %d dispatched tools", len(invoiceToolPermissions), len(dispatched))
	}
	for _, toolName := range []string{createInvoiceToolName, recordPaymentToolName, sendInvoiceToolName} {
		if got := invoiceToolPermissions[toolName].permission; got == "finance:invoices:view" {
			t.Errorf("write tool %s is gated on the read permission", toolName)
		}
	}
}
