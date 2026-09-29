package authorization

import "testing"

func TestResolveIndependentAppRolesAndLicenseReporting(t *testing.T) {
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
	// Licensing is plan metadata, so an inactive license row is reported but no
	// longer removes the access its role assignment grants. Revoking the
	// assignment is the revocation mechanism.
	financeAccess := appAccessByID(t, access, finance)
	if !financeAccess.HasAccess {
		t.Fatalf("expected the finance role assignment to grant access: %#v", financeAccess)
	}
	if financeAccess.Licensed {
		t.Fatalf("expected finance to still be reported as unlicensed: %#v", financeAccess)
	}
	flowAccess := appAccessByID(t, access, flow)
	if !flowAccess.HasAccess || flowAccess.DisplayRole != "Manager" {
		t.Fatalf("expected independent Flow access: %#v", flowAccess)
	}
}

func TestResolveRevokedAssignmentRemovesAccess(t *testing.T) {
	flow := AppID("flow")
	roleID := int64(11)
	access := Resolve(Snapshot{
		BusinessID:       7,
		DashRole:         "manager",
		MembershipActive: true,
		SubscribedApps:   []AppID{flow},
		Licenses:         []License{{AppID: flow, Source: SourceExplicit, Active: true}},
		Assignments:      []RoleAssignment{{AppID: flow, RoleIDs: []int64{roleID}, Source: SourceExplicit, Active: false}},
		Roles:            []Role{{ID: roleID, AppID: &flow, Permissions: []string{"flow:tasks:edit"}}},
	})
	if appAccessByID(t, access, flow).HasAccess {
		t.Fatalf("a revoked role assignment must remove access: %#v", access.Apps)
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
			{AppID: flow, RoleIDs: []int64{managerID}, Source: SourceExplicit, Scopes: []Scope{{Type: "flow_project", ID: "1"}}, Active: true},
			{AppID: flow, GroupID: int64Pointer(22), RoleIDs: []int64{contributorID}, Source: SourceGroup, Scopes: []Scope{{Type: "flow_project", ID: "2"}}, Active: true},
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

func TestResolveOwnerAndAdminAccessFollowsTheAppSetNotLicensing(t *testing.T) {
	flow := AppID("flow")
	// An admin takes the system path for the apps it is already entitled to.
	// Entitlement is the caller's app set (an owner's plan, a member's role
	// assignments), so a missing license row neither blocks nor is waived: it is
	// reported as plan metadata.
	access := Resolve(Snapshot{
		BusinessID:       11,
		DashRole:         "admin",
		MembershipActive: true,
		SubscribedApps:   []AppID{flow},
		AllPermissions:   []string{"flow:tasks:edit"},
	})
	if !access.Apps[0].HasAccess {
		t.Fatalf("expected the admin system path to grant an entitled app: %#v", access.Apps[0])
	}
	if access.Apps[0].Licensed {
		t.Fatalf("licensing must be reported rather than waived for admins: %#v", access.Apps[0])
	}

	noApps := Resolve(Snapshot{BusinessID: 11, DashRole: "admin", MembershipActive: true})
	if len(noApps.Apps) != 0 {
		t.Fatalf("expected no app access without an entitled app set: %#v", noApps.Apps)
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

func TestResolveRejectsInvalidAssignmentSourceTargets(t *testing.T) {
	flow := AppID("flow")
	roleID := int64(4)
	// Only a malformed *assignment* denies access. The license row is plan
	// metadata, so a malformed license alongside a valid assignment is allowed
	// and covered separately below.
	cases := []struct {
		name      string
		source    AssignmentSource
		licenseID *int64
		roleID    *int64
	}{
		{name: "group role without group", source: SourceGroup, licenseID: int64Pointer(22), roleID: nil},
		{name: "unknown source", source: AssignmentSource("invalid"), licenseID: nil, roleID: nil},
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

func TestResolveAllowsValidAssignmentWithMalformedLicense(t *testing.T) {
	flow := AppID("flow")
	roleID := int64(5)
	// A group-sourced license missing its group is a reporting defect, not an
	// access decision: the role assignment still governs.
	access := Resolve(Snapshot{
		BusinessID:       14,
		DashRole:         "member",
		MembershipActive: true,
		SubscribedApps:   []AppID{flow},
		Licenses:         []License{{AppID: flow, Source: SourceGroup, Active: true}},
		Assignments:      []RoleAssignment{{AppID: flow, GroupID: int64Pointer(22), RoleIDs: []int64{roleID}, Source: SourceGroup, Active: true}},
		Roles:            []Role{{ID: roleID, AppID: &flow, Permissions: []string{"flow:tasks:edit"}}},
	})
	app := access.Apps[0]
	if !app.HasAccess {
		t.Fatalf("a valid group assignment must not be denied by a malformed license: %#v", app)
	}
	if app.Licensed {
		t.Fatalf("a malformed license must not be reported as licensed: %#v", app)
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
