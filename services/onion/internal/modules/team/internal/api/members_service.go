package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/modules/team/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubteamgroupsv1 "github.com/voxire/voxie/gen/go/hub/team/groups/v1"
	hubteammembersv1 "github.com/voxire/voxie/gen/go/hub/team/members/v1"
)

type MembersAPI struct {
	Service *application.Service
}

func NewMembersAPI(service *application.Service) *MembersAPI {
	return &MembersAPI{Service: service}
}

func (a *MembersAPI) ListTeam(ctx context.Context, _ *hubteammembersv1.ListTeamRequest) (*hubteammembersv1.ListTeamResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	users, _, err := a.Service.ListStaffUsersByBusiness(ctx, device.User.BusinessID, 1, 50)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list users")
	}

	payloads := make([]*hubteammembersv1.StaffUserPayload, 0, len(users))
	for _, u := range users {
		payloads = append(payloads, newProtoStaffUser(u))
	}

	return &hubteammembersv1.ListTeamResponse{Users: payloads}, nil
}

func (a *MembersAPI) GetMember(ctx context.Context, req *hubteammembersv1.GetMemberRequest) (*hubteammembersv1.GetMemberResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	memberID := req.GetUserId()
	if memberID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid member id")
	}

	member, err := a.Service.GetStaffUserByID(ctx, device.User.UserID, device.User.BusinessID, memberID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Member not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get member")
	}

	groups := make([]*hubteamgroupsv1.GroupSummary, 0, len(member.Groups))
	for _, g := range member.Groups {
		groups = append(groups, &hubteamgroupsv1.GroupSummary{
			Id:    g.ID,
			Name:  g.Name,
			Color: g.Color,
		})
	}

	staff := newProtoStaffUser(*member)
	return &hubteammembersv1.GetMemberResponse{
		User: &hubteammembersv1.StaffUserWithGroups{
			Id:        staff.GetId(),
			Email:     staff.GetEmail(),
			FullName:  staff.GetFullName(),
			CreatedAt: staff.GetCreatedAt(),
			Apps:      staff.GetApps(),
			Role:      staff.GetRole(),
			IsActive:  staff.GetIsActive(),
			Groups:    groups,
		},
	}, nil
}

func (a *MembersAPI) RemoveMember(ctx context.Context, req *hubteammembersv1.RemoveMemberRequest) (*hubteammembersv1.RemoveMemberResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	memberID := req.GetUserId()
	if memberID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid member id")
	}

	actorID := device.User.UserID
	if err := a.Service.DeactivateUser(ctx, actorID, device.User.BusinessID, memberID); err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden):
			return nil, onekiterrors.PermissionDenied()
		case errors.Is(err, domain.ErrUserNotFound):
			return nil, onekiterrors.New(http.StatusNotFound, "Member not found")
		default:
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to deactivate member")
		}
	}

	return &hubteammembersv1.RemoveMemberResponse{Message: "Member deactivated"}, nil
}
