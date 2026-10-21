package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/organization/team/internal/application"
	"github.com/voxire/onion/internal/modules/organization/team/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	authorizationtransport "github.com/voxire/onion/internal/platform/authorization/transport"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubteampermissionsv1 "github.com/voxire/voxie/gen/go/hub/team/permissions/v1"
)

type PermissionsAPI struct {
	Service *application.Service
}

func NewPermissionsAPI(service *application.Service) *PermissionsAPI {
	return &PermissionsAPI{Service: service}
}

func (a *PermissionsAPI) UpdatePermissions(ctx context.Context, req *hubteampermissionsv1.UpdatePermissionsRequest) (*hubteampermissionsv1.UpdatePermissionsResponse, error) {
	device, permissionErr := requireHubPermission(ctx, "hub:team:edit")
	if permissionErr != nil {
		return nil, permissionErr
	}
	if req.GetReplaceLicensedAppIds() && !authPkg.HasDashboardPermission(device, "hub:licenses:manage") {
		return nil, onekiterrors.PermissionDenied()
	}

	var assignments *[]domain.AppRoleAssignmentInput
	if req.GetReplacePermissionProfileAssignments() {
		parsed, assignmentErr := permissionProfileAssignmentsFromProto(req.GetPermissionProfileAssignments())
		if assignmentErr != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, assignmentErr.Error())
		}
		assignments = &parsed
	}
	var licensedApps *[]string
	if req.GetReplaceLicensedAppIds() {
		licensedApps = &req.LicensedAppIds
	}
	var groupIDs *[]int64
	if req.GetReplaceGroupIds() {
		groupIDs = &req.GroupIds
	}

	if err := a.Service.UpdatePermissions(ctx, device.UserID, device.BusinessID, domain.UpdatePermissionsInput{
		UserID:                       req.GetUserId(),
		MemberType:                   req.GetMemberType(),
		FullName:                     req.FullName,
		LicensedAppIDs:               licensedApps,
		PermissionProfileAssignments: assignments,
		GroupIDs:                     groupIDs,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to update team authorization", logger.Error(err))
		switch {
		case errors.Is(err, domain.ErrForbidden):
			return nil, onekiterrors.PermissionDenied()
		case errors.Is(err, domain.ErrDelegationCeiling):
			return nil, onekiterrors.PermissionDenied()
		case errors.Is(err, domain.ErrUserNotFound), errors.Is(err, domain.ErrPermissionNotFound):
			return nil, onekiterrors.New(http.StatusNotFound, "Permission not found")
		default:
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update permissions")
		}
	}

	access, err := a.Service.GetEffectiveAccess(ctx, device.BusinessID, req.GetUserId())
	if err != nil {
		logger.Ctx(ctx).Error("Failed to resolve updated team authorization", logger.Error(err))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to resolve updated access")
	}
	return &hubteampermissionsv1.UpdatePermissionsResponse{EffectiveAccess: authorizationtransport.EffectiveAccessToProto(access)}, nil
}
