package products

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func productToolUser(permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		UserID:     9101,
		BusinessID: 9201,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID:           9201,
			AuthorizationVersion: 1,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppERP,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
		},
	}
}

func assertProductToolDenied(t *testing.T, toolName string, output map[string]any, handled bool, err error) {
	t.Helper()
	if !handled {
		t.Fatalf("%s was not recognized as a products tool", toolName)
	}
	if err == nil {
		t.Fatalf("%s without its permission unexpectedly succeeded", toolName)
	}
	if err.Error() != onekiterrors.PermissionDeniedMessage {
		t.Fatalf("%s error = %q, want %q", toolName, err.Error(), onekiterrors.PermissionDeniedMessage)
	}
	if output != nil {
		t.Fatalf("%s returned output while denied: %#v", toolName, output)
	}
}

// Each ERP product write must require the same permission its HTTP twin does,
// so a caller holding only the read permission cannot mutate the catalog.
func TestProductWriteToolsDenyCallerWithoutTheMatchingPermission(t *testing.T) {

	for _, tc := range []struct {
		tool       string
		permission string
		otherWrite string
	}{
		{createProductToolName, "erp:products:create", "erp:products:edit"},
		{updateProductToolName, "erp:products:edit", "erp:products:create"},
		{deleteProductToolName, "erp:products:delete", "erp:products:edit"},
	} {
		t.Run(tc.tool, func(t *testing.T) {
			readOnly := authPkg.WithAuthUser(context.Background(), productToolUser("erp:products:view"))
			output, handled, err := (&Module{}).ExecuteCaretTool(readOnly, tc.tool, map[string]any{"product_id": 7})
			assertProductToolDenied(t, tc.tool, output, handled, err)

			// A different write permission must not stand in for this one.
			otherWrite := authPkg.WithAuthUser(context.Background(), productToolUser("erp:products:view", tc.otherWrite))
			output, handled, err = (&Module{}).ExecuteCaretTool(otherWrite, tc.tool, map[string]any{"product_id": 7})
			assertProductToolDenied(t, tc.tool, output, handled, err)
		})
	}
}

func TestProductWriteToolsAllowCallerHoldingThePermission(t *testing.T) {

	for _, tc := range []struct {
		tool       string
		permission string
	}{
		{createProductToolName, "erp:products:create"},
		{updateProductToolName, "erp:products:edit"},
		{deleteProductToolName, "erp:products:delete"},
	} {
		t.Run(tc.tool, func(t *testing.T) {
			ctx := authPkg.WithAuthUser(context.Background(), productToolUser("erp:products:view", tc.permission))
			_, handled, err := (&Module{}).ExecuteCaretTool(ctx, tc.tool, map[string]any{"product_id": 7})
			if !handled {
				t.Fatalf("%s was not recognized as a products tool", tc.tool)
			}
			if err != nil && err.Error() == onekiterrors.PermissionDeniedMessage {
				t.Fatalf("caller holding %s was denied", tc.permission)
			}
		})
	}
}

func TestProductToolsDenyCallerWithoutPermission(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), productToolUser())
	output, handled, err := (&Module{}).ExecuteCaretTool(ctx, deleteProductToolName, map[string]any{"product_id": 7})
	assertProductToolDenied(t, deleteProductToolName, output, handled, err)
}

func TestProductToolsIgnoreTenantFromToolInput(t *testing.T) {

	spoofed := map[string]any{"product_id": 7, "business_id": 9201, "owner_id": 9101}
	readOnly := authPkg.WithAuthUser(context.Background(), productToolUser("erp:products:view"))
	output, handled, err := (&Module{}).ExecuteCaretTool(readOnly, deleteProductToolName, spoofed)
	assertProductToolDenied(t, deleteProductToolName, output, handled, err)

	privileged := authPkg.WithAuthUser(context.Background(), productToolUser("erp:products:view", "erp:products:delete"))
	_, handled, err = (&Module{}).ExecuteCaretTool(privileged, deleteProductToolName, spoofed)
	if !handled {
		t.Fatal("erp.delete_product was not recognized as a products tool")
	}
	if err != nil && err.Error() == onekiterrors.PermissionDeniedMessage {
		t.Fatal("privileged caller was denied; the gate is not reading the auth context")
	}
}

// A tool dispatched without a declared permission is unreachable, and a
// declared permission without an executor is a hard error rather than a silent
// pass-through to the next executor in the chain.
func TestProductToolPermissionsCoverEveryDispatchedTool(t *testing.T) {

	dispatched := []string{
		getProductToolName,
		listReorderSuggestionsToolName,
		createProductToolName,
		updateProductToolName,
		deleteProductToolName,
	}
	for _, toolName := range dispatched {
		if _, ok := productToolPermissions[toolName]; !ok {
			t.Errorf("%s is dispatched without a declared permission", toolName)
		}
	}
	if len(productToolPermissions) != len(dispatched) {
		t.Errorf("declared %d permissions for %d dispatched tools", len(productToolPermissions), len(dispatched))
	}
	for _, toolName := range []string{createProductToolName, updateProductToolName, deleteProductToolName} {
		if got := productToolPermissions[toolName].permission; got == "erp:products:view" {
			t.Errorf("write tool %s is gated on the read permission", toolName)
		}
	}

	productToolPermissions["erp.not_dispatched"] = productToolPermission{
		app:        authorization.AppERP,
		permission: "erp:products:view",
	}
	defer delete(productToolPermissions, "erp.not_dispatched")
	ctx := authPkg.WithAuthUser(context.Background(), productToolUser("erp:products:view"))
	_, handled, err := (&Module{}).ExecuteCaretTool(ctx, "erp.not_dispatched", nil)
	if !handled || err == nil {
		t.Fatalf("declared-but-unimplemented tool = (handled %v, %v), want (true, error)", handled, err)
	}
}
