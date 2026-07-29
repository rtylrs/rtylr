package api

import (
	"github.com/voxire/onion/internal/modules/team/internal/domain"
	hubteammembersv1 "github.com/voxire/voxie/gen/go/hub/team/members/v1"
)

func newProtoStaffUser(u domain.StaffUser) *hubteammembersv1.StaffUserPayload {
	apps := u.Apps
	if apps == nil {
		apps = []string{}
	}

	return &hubteammembersv1.StaffUserPayload{
		Id:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Apps:      apps,
		Role:      string(u.Role),
		IsActive:  u.IsActive,
	}
}
