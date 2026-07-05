package api

import (
	"context"
	"errors"
	"net/http"

	hubteampermissionsv1 "github.com/voxire/voxie/gen/go/hub/team/permissions/v1"

	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/modules/team/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type PermissionsAPI struct {
	Service *application.Service
}

func NewPermissionsAPI(service *application.Service) *PermissionsAPI {
	return &PermissionsAPI{Service: service}
}

func (a *PermissionsAPI) UpdatePermissions(ctx context.Context, req *hubteampermissionsv1.UpdatePermissionsRequest) (*hubteampermissionsv1.UpdatePermissionsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	appsSlice := req.GetApps()

	if err := a.Service.UpdatePermissions(ctx, device.User.UserID, device.User.BusinessID, domain.UpdatePermissionsInput{
		UserID: req.GetUserId(),
		Apps:   &appsSlice,
		Role:   domain.Role(req.GetRole()),
	}); err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden):
			return nil, onekiterrors.PermissionDenied()
		case errors.Is(err, domain.ErrUserNotFound), errors.Is(err, domain.ErrPermissionNotFound):
			return nil, onekiterrors.New(http.StatusNotFound, "Permission not found")
		default:
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update permissions")
		}
	}

	return &hubteampermissionsv1.UpdatePermissionsResponse{}, nil
}
