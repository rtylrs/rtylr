package authorization

import "testing"

func TestCanDelegatePermissionsRequiresSubsetForNonAdmin(t *testing.T) {
	access := &EffectiveAccess{
		DashRole:        "manager",
		DashPermissions: []string{"hub:roles:edit"},
		Apps: []EffectiveAppAccess{{
			AppID:     AppFlow,
			Licensed:  true,
			HasAccess: true,
			Permissions: []string{
				"flow:projects:view",
			},
		}},
	}

	if !CanDelegatePermissions(access, AppFlow, []string{"flow:projects:view"}) {
		t.Fatal("expected an actor to delegate a permission they hold")
	}
	if CanDelegatePermissions(access, AppFlow, []string{"flow:projects:edit"}) {
		t.Fatal("did not expect an actor to delegate a permission they do not hold")
	}
}

func TestCanAssignDashboardRoleEnforcesStrictCeiling(t *testing.T) {
	access := &EffectiveAccess{DashRole: "manager"}
	if !CanAssignDashboardRole(access, "member") {
		t.Fatal("expected a manager to assign a lower Dash role")
	}
	if CanAssignDashboardRole(access, "manager") {
		t.Fatal("did not expect a manager to assign a peer role")
	}
	if CanAssignDashboardRole(access, "admin") {
		t.Fatal("did not expect a manager to assign an admin role")
	}
}
