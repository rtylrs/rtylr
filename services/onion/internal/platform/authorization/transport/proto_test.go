package transport

import (
	"testing"

	"github.com/voxire/onion/internal/platform/authorization"
)

func TestPermissionProfileAssignmentsToProtoPreservesSourceAndScopes(t *testing.T) {
	groupID := int64(12)
	got := PermissionProfileAssignmentsToProto([]authorization.RoleAssignment{
		{
			AppID: authorization.AppFlow, GroupID: &groupID, RoleIDs: []int64{73},
			Source: authorization.SourceGroup, Scopes: []authorization.Scope{{Type: "project", ID: "9"}}, Active: true,
		},
		{AppID: authorization.AppPOS, Source: authorization.SourceExplicit},
	})
	if len(got) != 2 {
		t.Fatalf("got %d assignments, want 2", len(got))
	}
	if got[0].GetAppId() != "flow" || got[0].GetAssignmentSource() != "group" || got[0].GetStatus() != "active" || got[0].GetGroupId() != "12" {
		t.Fatalf("group assignment metadata was not preserved: %+v", got[0])
	}
	if len(got[0].GetPermissionProfileIds()) != 1 || got[0].GetPermissionProfileIds()[0] != "73" || len(got[0].GetScopes()) != 1 {
		t.Fatalf("group assignment details were not preserved: %+v", got[0])
	}
	if got[1].GetAppId() != "pos" || got[1].GetAssignmentSource() != "explicit" || got[1].GetStatus() != "revoked" {
		t.Fatalf("direct assignment metadata was not preserved: %+v", got[1])
	}
}

func TestEffectiveAccessToProtoPreservesPermissionsScopesAndLicenses(t *testing.T) {
	groupID := int64(12)
	got := EffectiveAccessToProto(&authorization.EffectiveAccess{
		BusinessID: 5, DashPermissions: []string{"hub:team:view"}, SubscribedApps: []authorization.AppID{authorization.AppFlow},
		Licenses: []authorization.License{{AppID: authorization.AppFlow, Source: authorization.SourceGroup, Active: true, GroupID: &groupID}},
		Apps: []authorization.EffectiveAppAccess{{
			AppID: authorization.AppFlow, Licensed: true, HasAccess: true, Permissions: []string{"flow:tasks:read"},
			PermissionScopes: map[string][]authorization.Scope{"flow:tasks:read": {{Type: "project", ID: "9"}}},
		}},
	})
	if got.GetBusinessId() != "5" || len(got.GetDashPermissions()) != 1 {
		t.Fatalf("effective access identity was not preserved: %+v", got)
	}
	if len(got.GetLicenses()) != 1 || got.GetLicenses()[0].GetGroupId() != "12" {
		t.Fatalf("license source was not preserved: %+v", got.GetLicenses())
	}
	if len(got.GetApps()) != 1 || !got.GetApps()[0].GetLicensed() || !got.GetApps()[0].GetHasAccess() {
		t.Fatalf("effective app access was not preserved: %+v", got.GetApps())
	}
	scopes := got.GetApps()[0].GetPermissionScopes()
	if len(scopes) != 1 || scopes[0].GetPermissionKey() != "flow:tasks:read" || scopes[0].GetScopes()[0].GetScopeId() != "9" {
		t.Fatalf("permission-specific scopes were not preserved: %+v", scopes)
	}
}
