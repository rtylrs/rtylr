package authorization

import "testing"

func TestResolveIndependentAppRolesAndLicenseRevocation(t *testing.T) {
	flow := AppID("flow")
	finance := AppID("finance")
	flowRole := int64(10)
	financeRole := int64(20)
	access := Resolve(Snapshot{
		BusinessID:       7,
		DashRole:         "manager",
		MembershipActive: true,
		SubscribedApps:   []AppID{flow, finance},
		Licenses: []License{
			{AppID: flow, Source: SourceExplicit, Active: true},
			{AppID: finance, Source: SourceExplicit, Active: false},
		},
		Assignments: []RoleAssignment{
			{AppID: flow, RoleIDs: []int64{flowRole}, PrimaryRoleID: &flowRole, Source: SourceExplicit, Active: true},
			{AppID: finance, RoleIDs: []int64{financeRole}, PrimaryRoleID: &financeRole, Source: SourceExplicit, Active: true},
		},
		Roles: []Role{
			{ID: flowRole, AppID: &flow, Key: "flow.manager", DisplayName: "Manager", Permissions: []string{"flow:tasks:edit"}},
			{ID: financeRole, AppID: &finance, Key: "finance.viewer", DisplayName: "Viewer", Permissions: []string{"finance:reports:view"}},
		},
	})
	financeAccess := appAccessByID(t, access, finance)
	if financeAccess.Licensed || financeAccess.HasAccess {
		t.Fatalf("expected finance role to remain history-only after license removal: %#v", access.Apps)
	}
	flowAccess := appAccessByID(t, access, flow)
	if !flowAccess.HasAccess || flowAccess.DisplayRole != "Manager" {
		t.Fatalf("expected independent Flow access: %#v", flowAccess)
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

func TestResolveGroupAndDirectAssignmentsUnionPermissionsAndScopes(t *testing.T) {
	flow := AppID("flow")
	managerID := int64(1)
	contributorID := int64(2)
	access := Resolve(Snapshot{
		BusinessID:       9,
		DashRole:         "member",
		MembershipActive: true,
		SubscribedApps:   []AppID{flow},
		Licenses:         []License{{AppID: flow, Source: SourceGroup, GroupID: int64Pointer(22), Active: true}},
		Assignments: []RoleAssignment{
			{AppID: flow, RoleIDs: []int64{managerID}, Source: SourceExplicit, Scopes: []Scope{{Type: "project", ID: "1"}}, Active: true},
			{AppID: flow, GroupID: int64Pointer(22), RoleIDs: []int64{contributorID}, Source: SourceGroup, Scopes: []Scope{{Type: "project", ID: "2"}}, Active: true},
		},
		Roles: []Role{
			{ID: managerID, AppID: &flow, Key: "flow.manager", Permissions: []string{"flow:projects:manage"}},
			{ID: contributorID, AppID: &flow, Key: "flow.contributor", Permissions: []string{"flow:tasks:edit"}},
		},
	})
	app := access.Apps[0]
	if !app.HasAccess || len(app.RoleIDs) != 2 || len(app.Scopes) != 2 {
		t.Fatalf("expected union of direct and group assignments: %#v", app)
	}
	if len(app.PermissionSources["flow:tasks:edit"]) != 1 || len(app.PermissionSources["flow:projects:manage"]) != 1 {
		t.Fatalf("expected permission source attribution: %#v", app.PermissionSources)
	}
}

func TestResolveOwnerAndAdminRequireSubscriptionEntitlement(t *testing.T) {
	flow := AppID("flow")
	access := Resolve(Snapshot{
		BusinessID:       11,
		DashRole:         "admin",
		MembershipActive: true,
		SubscribedApps:   []AppID{flow},
		AllPermissions:   []string{"flow:tasks:edit"},
	})
	if !access.Apps[0].HasAccess || !access.Apps[0].Licensed {
		t.Fatalf("expected subscribed app override for admin: %#v", access.Apps[0])
	}

	noSubscription := Resolve(Snapshot{BusinessID: 11, DashRole: "admin", MembershipActive: true})
	if len(noSubscription.Apps) != 0 {
		t.Fatalf("expected no app access without entitlement: %#v", noSubscription.Apps)
	}
}

func TestResolveArchivedRoleAndInactiveMembershipFailClosed(t *testing.T) {
	flow := AppID("flow")
	roleID := int64(3)
	access := Resolve(Snapshot{
		BusinessID:       12,
		DashRole:         "member",
		MembershipActive: false,
		SubscribedApps:   []AppID{flow},
		Licenses:         []License{{AppID: flow, Source: SourceExplicit, Active: true}},
		Assignments:      []RoleAssignment{{AppID: flow, RoleIDs: []int64{roleID}, Source: SourceExplicit, Active: true}},
		Roles:            []Role{{ID: roleID, AppID: &flow, IsArchived: true, Permissions: []string{"flow:tasks:edit"}}},
	})
	if access.Apps[0].HasAccess {
		t.Fatalf("expected inactive membership to revoke access: %#v", access.Apps[0])
	}
}

func TestResolveRejectsInvalidSourceTargets(t *testing.T) {
	flow := AppID("flow")
	roleID := int64(4)
	cases := []struct {
		name      string
		source    AssignmentSource
		licenseID *int64
		roleID    *int64
	}{
		{name: "group license without group", source: SourceGroup, licenseID: nil, roleID: int64Pointer(22)},
		{name: "group role without group", source: SourceGroup, licenseID: int64Pointer(22), roleID: nil},
		{name: "unknown source", source: AssignmentSource("invalid"), licenseID: nil, roleID: nil},
		{name: "explicit license with group", source: SourceExplicit, licenseID: int64Pointer(22), roleID: nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			licenses := []License{{AppID: flow, Source: tc.source, GroupID: tc.licenseID, Active: true}}
			assignments := []RoleAssignment{{AppID: flow, Source: tc.source, GroupID: tc.roleID, RoleIDs: []int64{roleID}, Active: true}}
			access := Resolve(Snapshot{
				BusinessID: 13, MembershipActive: true, SubscribedApps: []AppID{flow},
				Licenses: licenses, Assignments: assignments,
				Roles: []Role{{ID: roleID, AppID: &flow, Permissions: []string{"flow:tasks:edit"}}},
			})
			if access.Apps[0].HasAccess {
				t.Fatalf("invalid source target retained app access: %#v", access.Apps[0])
			}
		})
	}
}

func TestLegacyAccessPreservesOnlyLegacyAppGrants(t *testing.T) {
	access := LegacyAccess(12, []string{"flow", "finance", "unknown"}, "operator")
	if len(access.Apps) != 2 || len(access.Licenses) != 2 {
		t.Fatalf("expected only known legacy app grants: %#v", access)
	}
	if access.Apps[0].Permissions == nil || access.Apps[1].Permissions == nil {
		t.Fatal("expected legacy access to expose empty feature permissions")
	}
	if access.IsOwner || access.IsAdmin {
		t.Fatal("operator must not become an owner or admin during rollback")
	}
}
