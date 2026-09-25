package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/modules/team/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
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
	if req.GetApps() != nil && !authPkg.HasDashboardPermission(device, "hub:licenses:manage") {
		return nil, onekiterrors.PermissionDenied()
	}

	var apps *[]string
	if req.Apps != nil {
		apps = &req.Apps
	}
	appAssignments, assignmentErr := appAssignmentsFromProto(req.GetAppAssignments())
	if assignmentErr != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, assignmentErr.Error())
	}

	if err := a.Service.UpdatePermissions(ctx, device.UserID, device.BusinessID, domain.UpdatePermissionsInput{
		UserID:         req.GetUserId(),
		FullName:       req.FullName,
		Apps:           apps,
		Role:           domain.Role(req.GetRole()),
		DashRole:       req.GetDashRole(),
		AppAssignments: appAssignments,
		GroupIDs:       req.GetGroupIds(),
	}); err != nil {
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

	return &hubteampermissionsv1.UpdatePermissionsResponse{}, nil
}
