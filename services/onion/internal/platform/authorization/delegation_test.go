package authorization

import "testing"

func TestCanDelegatePermissionsRequiresSubsetForNonAdmin(t *testing.T) {
	access := &EffectiveAccess{
		DashPermissions: []string{"hub:roles:edit"},
		Apps: []EffectiveAppAccess{{
			AppID:     AppFlow,
			Licensed:  true,
			HasAccess: true,
			Permissions: []string{
				"flow:projects:view",
			},
			PermissionScopes: map[string][]Scope{"flow:projects:view": {}},
		}},
	}

	if !CanDelegatePermissions(access, AppFlow, []string{"flow:projects:view"}) {
		t.Fatal("expected an actor to delegate a permission they hold")
	}
	if CanDelegatePermissions(access, AppFlow, []string{"flow:projects:edit"}) {
		t.Fatal("did not expect an actor to delegate a permission they do not hold")
	}
}
