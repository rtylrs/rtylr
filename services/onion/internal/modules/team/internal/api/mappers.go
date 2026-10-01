package api

import (
	"github.com/voxire/onion/internal/modules/team/internal/domain"
	authorizationtransport "github.com/voxire/onion/internal/platform/authorization/transport"
	hubteammembersv1 "github.com/voxire/voxie/gen/go/hub/team/members/v1"
)

func newProtoStaffUser(u domain.StaffUser) *hubteammembersv1.StaffUserPayload {
	return &hubteammembersv1.StaffUserPayload{
		Id:                           u.ID,
		Email:                        u.Email,
		FullName:                     u.FullName,
		CreatedAt:                    u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		IsActive:                     u.IsActive,
		MemberType:                   u.MemberType,
		LicensedAppIds:               append([]string{}, u.LicensedAppIDs...),
		PermissionProfileAssignments: authorizationtransport.PermissionProfileAssignmentsToProto(u.PermissionProfileAssignments),
		EffectiveAccess:              authorizationtransport.EffectiveAccessToProto(u.EffectiveAccess),
	}
}
