package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/organization/team/internal/application"
	"github.com/voxire/onion/internal/modules/organization/team/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubteamgroupsv1 "github.com/voxire/voxie/gen/go/hub/team/groups/v1"
	hubteammembersv1 "github.com/voxire/voxie/gen/go/hub/team/members/v1"
)

type MembersAPI struct {
	Service *application.Service
}

func NewMembersAPI(service *application.Service) *MembersAPI {
	return &MembersAPI{Service: service}
}

func (a *MembersAPI) ListTeam(ctx context.Context, req *hubteammembersv1.ListTeamRequest) (*hubteammembersv1.ListTeamResponse, error) {
	device, permissionErr := requireHubPermission(ctx, "hub:team:view")
	if permissionErr != nil {
		return nil, permissionErr
	}
	businessID, businessErr := tenant.ResolveBusinessID(ctx, device.BusinessID, device.UserID)
	if businessErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	page := req.GetPage()
	if page < 1 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}
	users, _, err := a.Service.ListStaffUsersByBusiness(ctx, businessID, int64(page), int64(pageSize))
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list users")
	}
	total, err := a.Service.CountStaffUsersByBusiness(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to count users")
	}
	const maxInt32 = int64(1<<31 - 1)
	if total > maxInt32 {
		total = maxInt32
	}

	payloads := make([]*hubteammembersv1.StaffUserPayload, 0, len(users))
	for _, u := range users {
		payloads = append(payloads, newProtoStaffUser(u))
	}

	return &hubteammembersv1.ListTeamResponse{Users: payloads, Total: int32(total)}, nil
}

func (a *MembersAPI) GetMember(ctx context.Context, req *hubteammembersv1.GetMemberRequest) (*hubteammembersv1.GetMemberResponse, error) {
	device, permissionErr := requireHubPermission(ctx, "hub:team:view")
	if permissionErr != nil {
		return nil, permissionErr
	}

	memberID := req.GetUserId()
	if memberID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid member id")
	}

	member, err := a.Service.GetStaffUserByID(ctx, device.BusinessID, memberID)
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
			Id:                           staff.GetId(),
			Email:                        staff.GetEmail(),
			FullName:                     staff.GetFullName(),
			CreatedAt:                    staff.GetCreatedAt(),
			IsActive:                     staff.GetIsActive(),
			MemberType:                   staff.GetMemberType(),
			LicensedAppIds:               staff.GetLicensedAppIds(),
			PermissionProfileAssignments: staff.GetPermissionProfileAssignments(),
			EffectiveAccess:              staff.GetEffectiveAccess(),
			Groups:                       groups,
		},
	}, nil
}

func (a *MembersAPI) RemoveMember(ctx context.Context, req *hubteammembersv1.RemoveMemberRequest) (*hubteammembersv1.RemoveMemberResponse, error) {
	device, permissionErr := requireHubPermission(ctx, "hub:team:deactivate")
	if permissionErr != nil {
		return nil, permissionErr
	}

	memberID := req.GetUserId()
	if memberID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid member id")
	}

	actorID := device.UserID
	if err := a.Service.DeactivateUser(ctx, actorID, device.BusinessID, memberID); err != nil {
		logger.Ctx(ctx).Error("Failed to deactivate team member", logger.Error(err))
		switch {
		case errors.Is(err, domain.ErrForbidden):
			return nil, onekiterrors.PermissionDenied()
		case errors.Is(err, domain.ErrUserNotFound):
			return nil, onekiterrors.New(http.StatusNotFound, "Member not found")
		default:
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to deactivate member")
		}
	}

	recordSeatChangeAsync(ctx, a.Service, device.BusinessID)

	return &hubteammembersv1.RemoveMemberResponse{Message: "Member deactivated"}, nil
}
