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
			Scopes:    []authorization.Scope{{Type: "pos_location", ID: "3001"}},
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

// everyApp is the full AppID set, so a future app constant is covered by
// construction and cannot silently inherit the wrong branch.
var everyApp = []authorization.AppID{
	authorization.AppHub, authorization.AppPOS, authorization.AppERP,
	authorization.AppHR, authorization.AppCRM, authorization.AppFinance,
	authorization.AppFlow, authorization.AppInsights, authorization.AppMenu,
	authorization.AppRecruit,
}

func TestHasAnyAppScopeAllowsUnscopedNormalizedHRMembers(t *testing.T) {
	// The regression that locked every non-owner out of HR: the resolver grants
	// the app, but no product path can create an HR scope, so requiring one
	// denied every member.
	user := &AuthUser{EffectiveAccess: &authorization.EffectiveAccess{
		AuthorizationVersion: 1,
		Apps: []authorization.EffectiveAppAccess{{
			AppID:     authorization.AppHR,
			HasAccess: true,
		}},
	}}
	if !HasAnyAppScope(user, authorization.AppHR) {
		t.Fatal("an unscoped normalized HR assignment must be allowed; HR models no resource scope")
	}
}

func TestHasAnyAppScopeOnlyRequiresScopesForPOSAndFlow(t *testing.T) {
	for _, appID := range everyApp {
		scoped := appRequiresScope(appID)
		t.Run(string(appID), func(t *testing.T) {
			user := &AuthUser{EffectiveAccess: &authorization.EffectiveAccess{
				AuthorizationVersion: 1,
				Apps: []authorization.EffectiveAppAccess{{
					AppID:     appID,
					HasAccess: true,
				}},
			}}
			got := HasAnyAppScope(user, appID)
			if scoped && got {
				t.Fatalf("%s is a resource scoped app and must fail closed without a scope", appID)
			}
			if !scoped && !got {
				t.Fatalf("%s models no resource scope and must not require one", appID)
			}
		})
	}
}

func TestHasAnyAppScopeStillFailsClosedWithoutAppAccess(t *testing.T) {
	// Proves the unscoped-app early return sits behind the HasAccess filter and
	// cannot widen access for an app the member cannot reach.
	for _, appID := range everyApp {
		user := &AuthUser{EffectiveAccess: &authorization.EffectiveAccess{
			AuthorizationVersion: 1,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:     appID,
				HasAccess: false,
			}},
		}}
		if HasAnyAppScope(user, appID) {
			t.Fatalf("%s must fail closed when the member has no app access", appID)
		}
	}
}

func TestHasAnyAppScopeKeepsLegacyMigrationShapeForEveryApp(t *testing.T) {
	// Guards the ordering: the version-0 escape hatch must still run before the
	// scope allowlist, so a migration can re-tighten any app.
	for _, appID := range everyApp {
		user := &AuthUser{EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:     appID,
				HasAccess: true,
			}},
		}}
		if !HasAnyAppScope(user, appID) {
			t.Fatalf("%s must retain legacy migration compatibility at version 0", appID)
		}
	}
}

func TestHasAnyAppScopeRejectsUnknownAppID(t *testing.T) {
	user := &AuthUser{EffectiveAccess: &authorization.EffectiveAccess{
		AuthorizationVersion: 1,
		Apps: []authorization.EffectiveAppAccess{{
			AppID:     authorization.AppHR,
			HasAccess: true,
		}},
	}}
	if HasAnyAppScope(user, authorization.AppID("not-an-app")) {
		t.Fatal("an app id with no matching resolved entry must fail closed")
	}
}

// Licensing is plan metadata. A member's access comes from role assignments, so
// an absent license row must neither grant nor deny it.
func TestAccessDoesNotDependOnLicenseRows(t *testing.T) {
	unlicensed := &AuthUser{EffectiveAccess: &authorization.EffectiveAccess{
		AuthorizationVersion: 1,
		Apps: []authorization.EffectiveAppAccess{{
			AppID:     authorization.AppHR,
			Licensed:  false,
			HasAccess: true,
			Permissions: []string{
				"hr:employees:view",
			},
		}},
	}}
	if !HasAppAccess(unlicensed, authorization.AppHR) {
		t.Fatal("app access must not require a license row")
	}
	if !HasAppPermission(unlicensed, authorization.AppHR, "hr:employees:view") {
		t.Fatal("app permissions must not require a license row")
	}
	if !HasAnyAppScope(unlicensed, authorization.AppHR) {
		t.Fatal("an unlicensed but assigned HR app must still resolve its scope rule")
	}
}
