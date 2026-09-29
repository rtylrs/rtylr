package inventory

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func inventoryToolUser(permissions ...string) *authPkg.AuthUser {
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
				Scopes:      []authorization.Scope{{Type: "erp", ID: "9201"}},
			}},
		},
	}
}

func assertInventoryToolDenied(t *testing.T, toolName string, output map[string]any, handled bool, err error) {
	t.Helper()
	if !handled {
		t.Fatalf("%s was not recognized as an inventory tool", toolName)
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

// erp.set_stock is a write tool whose HTTP twin requires erp:inventory:adjust.
// A caller holding only the inventory read permission must not reach it.
func TestSetStockToolDeniesCallerWithoutAdjustPermission(t *testing.T) {
	t.Setenv("AUTHORIZATION_V2_MODE", "v2")

	readOnly := authPkg.WithAuthUser(context.Background(), inventoryToolUser("erp:inventory:view"))
	output, handled, err := (&Module{}).ExecuteCaretTool(readOnly, setStockToolName, map[string]any{"product_id": 7, "quantity": 5})
	assertInventoryToolDenied(t, setStockToolName, output, handled, err)

	unrelated := authPkg.WithAuthUser(context.Background(), inventoryToolUser("erp:products:create"))
	output, handled, err = (&Module{}).ExecuteCaretTool(unrelated, setStockToolName, map[string]any{"product_id": 7, "quantity": 5})
	assertInventoryToolDenied(t, setStockToolName, output, handled, err)
}

func TestSetStockToolAllowsCallerWithAdjustPermission(t *testing.T) {
	t.Setenv("AUTHORIZATION_V2_MODE", "v2")

	ctx := authPkg.WithAuthUser(context.Background(), inventoryToolUser("erp:inventory:view", "erp:inventory:adjust"))
	_, handled, err := (&Module{}).ExecuteCaretTool(ctx, setStockToolName, map[string]any{"product_id": 7, "quantity": 5})
	if !handled {
		t.Fatal("erp.set_stock was not recognized as an inventory tool")
	}
	// The nil service and absent database fail after the gate; a permission
	// failure here would mean the gate rejected an authorized caller.
	if err != nil && err.Error() == onekiterrors.PermissionDeniedMessage {
		t.Fatal("caller holding erp:inventory:adjust was denied")
	}
}

func TestSetStockToolDenyPrecedesTenantAndServiceWork(t *testing.T) {
	t.Setenv("AUTHORIZATION_V2_MODE", "v2")

	// A caller with no inventory grant must be rejected by the gate, not by
	// input parsing or tenant resolution, so the rejection cannot be used to
	// probe another tenant.
	ctx := authPkg.WithAuthUser(context.Background(), inventoryToolUser("erp:products:view"))
	output, handled, err := (&Module{}).ExecuteCaretTool(ctx, setStockToolName, map[string]any{})
	assertInventoryToolDenied(t, setStockToolName, output, handled, err)
}

func TestInventoryToolAuthorizationIgnoresTenantFromToolInput(t *testing.T) {
	t.Setenv("AUTHORIZATION_V2_MODE", "v2")

	spoofed := map[string]any{"product_id": 7, "quantity": 5, "business_id": 9201, "owner_id": 9101}
	readOnly := authPkg.WithAuthUser(context.Background(), inventoryToolUser("erp:inventory:view"))
	output, handled, err := (&Module{}).ExecuteCaretTool(readOnly, setStockToolName, spoofed)
	assertInventoryToolDenied(t, setStockToolName, output, handled, err)

	privileged := authPkg.WithAuthUser(context.Background(), inventoryToolUser("erp:inventory:view", "erp:inventory:adjust"))
	_, handled, err = (&Module{}).ExecuteCaretTool(privileged, setStockToolName, spoofed)
	if !handled {
		t.Fatal("erp.set_stock was not recognized as an inventory tool")
	}
	if err != nil && err.Error() == onekiterrors.PermissionDeniedMessage {
		t.Fatal("privileged caller was denied; the gate is not reading the auth context")
	}
}

// A tool dispatched without a declared permission is unreachable, and a
// declared permission without an executor is a hard error rather than a silent
// pass-through to the next executor in the chain.
func TestInventoryToolPermissionsCoverEveryDispatchedTool(t *testing.T) {
	t.Setenv("AUTHORIZATION_V2_MODE", "v2")

	dispatched := []string{listProductsToolName, setStockToolName}
	for _, toolName := range dispatched {
		if _, ok := inventoryToolPermissions[toolName]; !ok {
			t.Errorf("%s is dispatched without a declared permission", toolName)
		}
	}
	if len(inventoryToolPermissions) != len(dispatched) {
		t.Errorf("declared %d permissions for %d dispatched tools", len(inventoryToolPermissions), len(dispatched))
	}
	if got := inventoryToolPermissions[setStockToolName].permission; got == "erp:inventory:view" {
		t.Errorf("erp.set_stock is gated on the read permission %q", got)
	}

	inventoryToolPermissions["erp.not_dispatched"] = inventoryToolPermission{
		app:        authorization.AppERP,
		permission: "erp:inventory:view",
	}
	defer delete(inventoryToolPermissions, "erp.not_dispatched")
	ctx := authPkg.WithAuthUser(context.Background(), inventoryToolUser("erp:inventory:view"))
	_, handled, err := (&Module{}).ExecuteCaretTool(ctx, "erp.not_dispatched", nil)
	if !handled || err == nil {
		t.Fatalf("declared-but-unimplemented tool = (handled %v, %v), want (true, error)", handled, err)
	}
}
