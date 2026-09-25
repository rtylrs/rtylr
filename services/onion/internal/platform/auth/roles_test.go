package auth

import (
	"testing"

	"github.com/voxire/onion/internal/platform/authorization"
)

func TestHasAnyRole(t *testing.T) {
	user := &AuthUser{Role: " Admin "}
	if !HasAnyRole(user, "owner", "admin") {
		t.Fatal("expected normalized admin role to match")
	}
	if HasAnyRole(user, "billing", "staff") {
		t.Fatal("unexpected role match")
	}
}

func TestIsBusinessAdmin(t *testing.T) {
	for _, role := range []string{"owner", "admin", "console_admin"} {
		if !IsBusinessAdmin(&AuthUser{Role: role}) {
			t.Errorf("role %q should be a business admin", role)
		}
	}
	if IsBusinessAdmin(&AuthUser{Role: "operator"}) {
		t.Fatal("operator should not be a business admin")
	}
}

func TestHasAnyAppScopeFailsClosedForNormalizedAssignments(t *testing.T) {
	user := &AuthUser{EffectiveAccess: &authorization.EffectiveAccess{
		AuthorizationVersion: 1,
		Apps: []authorization.EffectiveAppAccess{{
			AppID:     authorization.AppPOS,
			Licensed:  true,
			HasAccess: true,
			Scopes:    []authorization.Scope{{Type: "location", ID: "3001"}},
		}},
	}}
	if !HasAnyAppScope(user, authorization.AppPOS) {
		t.Fatal("expected normalized POS assignment with a location scope")
	}
	user.EffectiveAccess.Apps[0].Scopes = nil
	if HasAnyAppScope(user, authorization.AppPOS) {
		t.Fatal("normalized POS assignment without a resource scope must fail closed")
	}
}

func TestHasAnyAppScopeFailsClosedForUnscopedNormalizedFlowAssignments(t *testing.T) {
	user := &AuthUser{EffectiveAccess: &authorization.EffectiveAccess{
		AuthorizationVersion: 1,
		Apps: []authorization.EffectiveAppAccess{{
			AppID:     authorization.AppFlow,
			Licensed:  true,
			HasAccess: true,
		}},
	}}
	if HasAnyAppScope(user, authorization.AppFlow) {
		t.Fatal("normalized Flow assignment without a resource scope must fail closed")
	}
}

func TestHasAnyAppScopeAllowsOwnerAndAdminWithoutScopedAssignments(t *testing.T) {
	for _, tc := range []struct {
		name    string
		isOwner bool
		isAdmin bool
	}{
		{name: "owner", isOwner: true},
		{name: "admin", isAdmin: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			user := &AuthUser{EffectiveAccess: &authorization.EffectiveAccess{
				AuthorizationVersion: 1,
				IsOwner:              tc.isOwner,
				IsAdmin:              tc.isAdmin,
				Apps: []authorization.EffectiveAppAccess{{
					AppID:     authorization.AppFlow,
					Licensed:  true,
					HasAccess: true,
				}},
			}}
			if !HasAnyAppScope(user, authorization.AppFlow) {
				t.Fatalf("%s with system access must not need a resource scope", tc.name)
			}
		})
	}
}

func TestHasAnyAppScopeStillRequiresAccessForOwner(t *testing.T) {
	user := &AuthUser{EffectiveAccess: &authorization.EffectiveAccess{
		AuthorizationVersion: 1,
		IsOwner:              true,
		Apps: []authorization.EffectiveAppAccess{{
			AppID:     authorization.AppFlow,
			Licensed:  true,
			HasAccess: false,
		}},
	}}
	if HasAnyAppScope(user, authorization.AppFlow) {
		t.Fatal("owner without app access (inactive membership) must fail closed")
	}
}

func TestHasAnyAppScopePreservesLegacyMigrationShape(t *testing.T) {
	user := &AuthUser{EffectiveAccess: &authorization.EffectiveAccess{
		Apps: []authorization.EffectiveAppAccess{{
			AppID:     authorization.AppHR,
			Licensed:  true,
			HasAccess: true,
		}},
	}}
	if !HasAnyAppScope(user, authorization.AppHR) {
		t.Fatal("legacy access without a normalized version should retain migration compatibility")
	}
}
