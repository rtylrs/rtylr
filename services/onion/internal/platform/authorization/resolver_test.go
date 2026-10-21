package authorization

import "testing"

func TestResolveKeepsPermissionKeysAndScopesIndependent(t *testing.T) {
	flow := AppID("flow")
	readProfileID, writeProfileID := int64(10), int64(11)
	access := Resolve(Snapshot{
		BusinessID:       7,
		MembershipActive: true,
		SubscribedApps:   []AppID{flow},
		Licenses:         []License{{AppID: flow, Source: SourceExplicit, Active: true}},
		Assignments: []RoleAssignment{
			{AppID: flow, RoleIDs: []int64{readProfileID}, Source: SourceExplicit, Scopes: []Scope{{Type: "project", ID: "1"}}, Active: true},
			{AppID: flow, RoleIDs: []int64{writeProfileID}, Source: SourceGroup, GroupID: int64Pointer(22), Scopes: []Scope{{Type: "project", ID: "2"}}, Active: true},
		},
		Roles: []Role{
			{ID: readProfileID, AppID: &flow, Key: "flow.reader", Permissions: []string{"flow:tasks:read"}},
			{ID: writeProfileID, AppID: &flow, Key: "flow.writer", Permissions: []string{"flow:tasks:write"}},
		},
	})
	app := appAccessByID(t, access, flow)
	if !app.HasAccess || !app.Licensed {
		t.Fatalf("expected licensed app access: %#v", app)
	}
	if len(app.Permissions) != 2 || app.PermissionScopes["flow:tasks:read"][0].ID != "1" || app.PermissionScopes["flow:tasks:write"][0].ID != "2" {
		t.Fatalf("read/write scopes must remain independent: %#v", app)
	}
}

func TestResolveRegularMemberNeedsLicenseAndPermissionProfile(t *testing.T) {
	flow := AppID("flow")
	roleID := int64(11)
	base := Snapshot{
		BusinessID:       7,
		MembershipActive: true,
		SubscribedApps:   []AppID{flow},
		Assignments:      []RoleAssignment{{AppID: flow, RoleIDs: []int64{roleID}, Source: SourceExplicit, Active: true}},
		Roles:            []Role{{ID: roleID, AppID: &flow, Permissions: []string{"flow:tasks:read"}}},
	}
	withoutLicense := Resolve(base)
	if appAccessByID(t, withoutLicense, flow).HasAccess {
		t.Fatal("profile assignment without an app license granted access")
	}
	base.Licenses = []License{{AppID: flow, Source: SourceExplicit, Active: true}}
	withoutProfile := base
	withoutProfile.Assignments = nil
	if appAccessByID(t, Resolve(withoutProfile), flow).HasAccess {
		t.Fatal("app license without a permission profile granted access")
	}
	if !appAccessByID(t, Resolve(base), flow).HasAccess {
		t.Fatal("active license and profile assignment should grant access")
	}
}

func TestResolveActiveMemberCanEnterDashWithoutManagementPermissions(t *testing.T) {
	access := Resolve(Snapshot{BusinessID: 7, MembershipActive: true})
	if !containsPermission(access.DashPermissions, "hub:dashboard:view") {
		t.Fatalf("active member cannot enter Dash: %#v", access.DashPermissions)
	}
	if access.IsAdmin || access.IsOwner || containsPermission(access.DashPermissions, "hub:team:edit") {
		t.Fatalf("Dash entry granted management authority: %#v", access)
	}
}

func TestResolveOwnerAndAdminInheritFullCatalogForPlanApps(t *testing.T) {
	flow, finance := AppID("flow"), AppID("finance")
	allPermissions := []string{"flow:tasks:read", "flow:tasks:write", "finance:reports:read"}
	for _, snapshot := range []Snapshot{
		{BusinessID: 11, IsOwner: true, MembershipActive: true, SubscribedApps: []AppID{flow, finance}, AllPermissions: allPermissions},
		{BusinessID: 11, IsAdmin: true, MembershipActive: true, SubscribedApps: []AppID{flow, finance}, AllPermissions: allPermissions},
	} {
		access := Resolve(snapshot)
		for _, app := range access.Apps {
			if !app.HasAccess || !app.Licensed {
				t.Fatalf("owner/admin app is not fully unlocked: %#v", app)
			}
		}
	}
}

func TestResolveInactiveMembershipFailsClosed(t *testing.T) {
	flow := AppID("flow")
	access := Resolve(Snapshot{
		BusinessID:       12,
		MembershipActive: false,
		IsAdmin:          true,
		SubscribedApps:   []AppID{flow},
		AllPermissions:   []string{"flow:tasks:read"},
	})
	if access.IsAdmin || appAccessByID(t, access, flow).HasAccess {
		t.Fatalf("inactive membership retained privileges: %#v", access)
	}
}

func TestResolveRejectsInvalidAssignmentSourceTarget(t *testing.T) {
	flow := AppID("flow")
	roleID := int64(4)
	access := Resolve(Snapshot{
		BusinessID: 13, MembershipActive: true, SubscribedApps: []AppID{flow},
		Licenses:    []License{{AppID: flow, Source: SourceExplicit, Active: true}},
		Assignments: []RoleAssignment{{AppID: flow, Source: SourceGroup, RoleIDs: []int64{roleID}, Active: true}},
		Roles:       []Role{{ID: roleID, AppID: &flow, Permissions: []string{"flow:tasks:read"}}},
	})
	if appAccessByID(t, access, flow).HasAccess {
		t.Fatal("group permission assignment without a group id granted access")
	}
}

func appAccessByID(t *testing.T, access EffectiveAccess, appID AppID) EffectiveAppAccess {
	t.Helper()
	for _, app := range access.Apps {
		if app.AppID == appID {
			return app
		}
	}
	t.Fatalf("missing app access for %s: %#v", appID, access.Apps)
	return EffectiveAppAccess{}
}

func int64Pointer(value int64) *int64 { return &value }
