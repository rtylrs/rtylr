package api

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/platform/authorization"
)

func TestNewProtoStaffUserIncludesNormalizedAccess(t *testing.T) {
	groupID := int64(12)
	got := newProtoStaffUser(domain.StaffUser{
		ID: 42, MemberType: "member", IsActive: true,
		LicensedAppIDs: []string{"flow"},
		CreatedAt: time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC),
		PermissionProfileAssignments: []authorization.RoleAssignment{{
			AppID: authorization.AppFlow, RoleIDs: []int64{18}, Source: authorization.SourceGroup, GroupID: &groupID, Active: true,
		}},
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID: 5,
			Apps: []authorization.EffectiveAppAccess{{AppID: authorization.AppFlow, Licensed: true, HasAccess: true}},
		},
	})
	if got.GetMemberType() != "member" || !got.GetIsActive() || len(got.GetLicensedAppIds()) != 1 {
		t.Fatalf("member fields missing from payload: %+v", got)
	}
	if len(got.GetPermissionProfileAssignments()) != 1 {
		t.Fatalf("permission profile assignments missing from payload: %+v", got)
	}
	assignment := got.GetPermissionProfileAssignments()[0]
	if assignment.GetAssignmentSource() != "group" || assignment.GetGroupId() != "12" || len(assignment.GetPermissionProfileIds()) != 1 {
		t.Fatalf("group profile assignment missing from payload: %+v", assignment)
	}
	if access := got.GetEffectiveAccess(); access == nil || len(access.GetApps()) != 1 || !access.GetApps()[0].GetLicensed() {
		t.Fatalf("effective access missing from payload: %+v", access)
	}
}
