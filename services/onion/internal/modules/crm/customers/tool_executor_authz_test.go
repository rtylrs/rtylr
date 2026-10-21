package customers

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func customerToolUser(permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		UserID:     9101,
		BusinessID: 9201,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID:           9201,
			AuthorizationVersion: 1,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppCRM,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
		},
	}
}

func assertPermissionDenied(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected permission denied")
	}
	if denied := onekiterrors.PermissionDenied(); err.Error() != denied.Error() {
		t.Fatalf("error = %q, want %q", err.Error(), denied.Error())
	}
}

// A user holding only the CRM read permission must not create a customer
// through Caret, even though crm.create_customer is a registered write tool.
func TestCreateCustomerToolDeniesCallerWithoutCreatePermission(t *testing.T) {

	ctx := authPkg.WithAuthUser(context.Background(), customerToolUser("crm:customers:view"))
	output, handled, err := (&Module{}).ExecuteCaretTool(ctx, createCustomerToolName, map[string]any{"first_name": "Ada"})
	if !handled {
		t.Fatal("crm.create_customer was not recognized as a customers tool")
	}
	assertPermissionDenied(t, err)
	if output != nil {
		t.Fatalf("denied create_customer returned output: %#v", output)
	}
}

func TestCreateCustomerToolAllowsCallerWithCreatePermission(t *testing.T) {

	// The gate admits the caller; the executor then fails on the nil service,
	// which proves authorization passed rather than short-circuiting.
	ctx := authPkg.WithAuthUser(context.Background(), customerToolUser("crm:customers:create"))
	output, handled, err := (&Module{}).ExecuteCaretTool(ctx, createCustomerToolName, map[string]any{"first_name": "Ada"})
	if !handled {
		t.Fatal("crm.create_customer was not recognized as a customers tool")
	}
	if err != nil && err.Error() == onekiterrors.PermissionDeniedMessage {
		t.Fatal("caller holding crm:customers:create was denied")
	}
	if output != nil {
		t.Fatalf("nil-service module produced output: %#v", output)
	}
}

func TestCreateCustomerToolDeniesCallerWithoutPermission(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), customerToolUser())
	_, handled, err := (&Module{}).ExecuteCaretTool(ctx, createCustomerToolName, map[string]any{"first_name": "Ada"})
	if !handled {
		t.Fatal("crm.create_customer was not recognized as a customers tool")
	}
	assertPermissionDenied(t, err)
}

// Tool input must not be able to influence the authorization decision: the
// gate resolves the caller from the auth context, so a caller that supplies the
// "right" business in its input is still denied, and a privileged caller stays
// admitted even when the input names another tenant.
func TestCustomerToolAuthorizationIgnoresTenantFromToolInput(t *testing.T) {

	spoofed := map[string]any{"first_name": "Ada", "business_id": 9201, "owner_id": 9101}
	underPrivileged := authPkg.WithAuthUser(context.Background(), customerToolUser("crm:customers:view"))
	_, handled, err := (&Module{}).ExecuteCaretTool(underPrivileged, createCustomerToolName, spoofed)
	if !handled {
		t.Fatal("crm.create_customer was not recognized as a customers tool")
	}
	assertPermissionDenied(t, err)

	privileged := authPkg.WithAuthUser(context.Background(), customerToolUser("crm:customers:create"))
	_, handled, err = (&Module{}).ExecuteCaretTool(privileged, createCustomerToolName, spoofed)
	if !handled {
		t.Fatal("crm.create_customer was not recognized as a customers tool")
	}
	if err != nil && err.Error() == onekiterrors.PermissionDeniedMessage {
		t.Fatal("privileged caller was denied; the gate is not reading the auth context")
	}
}

func TestCustomerToolsDenyUnknownAndUnauthenticatedCallers(t *testing.T) {

	module := &Module{}
	output, handled, err := module.ExecuteCaretTool(context.Background(), "crm.not_a_tool", nil)
	if handled || err != nil || output != nil {
		t.Fatalf("unregistered tool = (%#v, %v, %v), want (nil, false, nil)", output, handled, err)
	}

	ctx := authPkg.WithAuthUser(context.Background(), customerToolUser("crm:customers:view"))
	_, handled, err = module.ExecuteCaretTool(ctx, "crm.not_a_tool", nil)
	if handled || err != nil {
		t.Fatalf("unregistered tool was claimed: handled=%v err=%v", handled, err)
	}
}

// A tool dispatched without a declared permission is unreachable, and a
// declared permission without an executor is a hard error rather than a silent
// pass-through to the next executor in the chain.
func TestCustomerToolPermissionsCoverEveryDispatchedTool(t *testing.T) {

	dispatched := []string{listCustomersToolName, getCustomerToolName, createCustomerToolName}
	for _, toolName := range dispatched {
		if _, ok := customerToolPermissions[toolName]; !ok {
			t.Errorf("%s is dispatched without a declared permission", toolName)
		}
	}
	if len(customerToolPermissions) != len(dispatched) {
		t.Errorf("declared %d permissions for %d dispatched tools", len(customerToolPermissions), len(dispatched))
	}
	if got := customerToolPermissions[createCustomerToolName].permission; got == "crm:customers:view" {
		t.Errorf("crm.create_customer is gated on the read permission %q", got)
	}

	// A permission declared for a tool with no executor fails loudly.
	customerToolPermissions["crm.not_dispatched"] = customerToolPermission{
		app:        authorization.AppCRM,
		permission: "crm:customers:view",
	}
	defer delete(customerToolPermissions, "crm.not_dispatched")
	ctx := authPkg.WithAuthUser(context.Background(), customerToolUser("crm:customers:view"))
	_, handled, err := (&Module{}).ExecuteCaretTool(ctx, "crm.not_dispatched", nil)
	if !handled || err == nil {
		t.Fatalf("declared-but-unimplemented tool = (handled %v, %v), want (true, error)", handled, err)
	}
}
