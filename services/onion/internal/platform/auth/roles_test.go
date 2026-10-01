package auth

import (
	"testing"

	"github.com/voxire/onion/internal/platform/authorization"
)

func TestIsBusinessAdminRequiresResolvedBusinessAccess(t *testing.T) {
	for _, tc := range []struct {
		name   string
		access *authorization.EffectiveAccess
		want   bool
	}{
		{name: "owner", access: &authorization.EffectiveAccess{IsOwner: true}, want: true},
		{name: "admin", access: &authorization.EffectiveAccess{IsAdmin: true}, want: true},
		{name: "member", access: &authorization.EffectiveAccess{}, want: false},
		{name: "token role alone", access: nil, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := IsBusinessAdmin(&AuthUser{EffectiveAccess: tc.access})
			if got != tc.want {
				t.Fatalf("IsBusinessAdmin() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestIsConsoleAdminIsSeparateFromBusinessAdmin(t *testing.T) {
	user := &AuthUser{IsConsoleAdmin: true, EffectiveAccess: &authorization.EffectiveAccess{}}
	if !IsConsoleAdmin(user) {
		t.Fatal("console admin token was not recognized")
	}
	if IsBusinessAdmin(user) {
		t.Fatal("console admin must not become a business admin")
	}
}

func TestAppPermissionRequiresLicenseAccessAndPermission(t *testing.T) {
	for _, tc := range []struct {
		name        string
		licensed    bool
		hasAccess   bool
		permissions []string
		want        bool
	}{
		{name: "allowed", licensed: true, hasAccess: true, permissions: []string{"hr:employees:view"}, want: true},
		{name: "missing license", hasAccess: true, permissions: []string{"hr:employees:view"}},
		{name: "missing profile", licensed: true, hasAccess: true},
		{name: "inactive access", licensed: true, permissions: []string{"hr:employees:view"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			user := &AuthUser{EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
				AppID: authorization.AppHR, Licensed: tc.licensed, HasAccess: tc.hasAccess, Permissions: tc.permissions,
			}}}}
			if got := HasAppPermission(user, authorization.AppHR, "hr:employees:view"); got != tc.want {
				t.Fatalf("HasAppPermission() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestHasAnyAppScopeAllowsScopedAndBusinessWidePermissionGrants(t *testing.T) {
	user := &AuthUser{EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
		AppID: authorization.AppFlow, Licensed: true, HasAccess: true,
		PermissionScopes: map[string][]authorization.Scope{
			"flow:tasks:view": {{Type: "flow_project", ID: "1"}},
		},
	}}}}
	if !HasAnyAppScope(user, authorization.AppFlow) {
		t.Fatal("expected a scoped Flow assignment")
	}
	user.EffectiveAccess.Apps[0].PermissionScopes["flow:tasks:view"] = []authorization.Scope{}
	if !HasAnyAppScope(user, authorization.AppFlow) {
		t.Fatal("an empty scope list is a business-wide permission grant")
	}
	user.EffectiveAccess.Apps[0].PermissionScopes = map[string][]authorization.Scope{}
	if HasAnyAppScope(user, authorization.AppFlow) {
		t.Fatal("an app with no permission grants must fail closed")
	}
}

func TestOwnerAndAdminReceiveBusinessWideResourceAccess(t *testing.T) {
	for _, tc := range []struct {
		name         string
		owner, admin bool
	}{
		{name: "owner", owner: true}, {name: "admin", admin: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			user := &AuthUser{EffectiveAccess: &authorization.EffectiveAccess{
				IsOwner: tc.owner, IsAdmin: tc.admin,
				Apps: []authorization.EffectiveAppAccess{{AppID: authorization.AppFlow, Licensed: true, HasAccess: true}},
			}}
			if !HasAnyAppScope(user, authorization.AppFlow) {
				t.Fatal("privileged member should have business-wide access")
			}
		})
	}
}

func TestHasAnyAppScopeRequiresActiveAppAccess(t *testing.T) {
	user := &AuthUser{EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
		AppID: authorization.AppHR, Licensed: true, HasAccess: false,
	}}}}
	if HasAnyAppScope(user, authorization.AppHR) {
		t.Fatal("inactive app access must be denied")
	}
}
