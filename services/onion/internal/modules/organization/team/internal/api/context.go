package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/organization/team/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func requireHubPermission(ctx context.Context, permission string) (*authPkg.AuthUser, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if !authPkg.HasDashboardPermission(device.User, permission) && !authPkg.IsBusinessAdmin(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}
	return device.User, nil
}

func teamOwnerIDFromCtx(ctx context.Context, service *application.Service) (int64, bool) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, false
	}
	if _, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID); err != nil {
		return 0, false
	}
	ownerID, err := service.BusinessOwnerID(ctx, device.User.UserID, device.User.BusinessID)
	return ownerID, err == nil && ownerID != 0
}

func teamManagerFromCtx(ctx context.Context, service *application.Service) (actorID, ownerID int64, ok bool) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, 0, false
	}
	if _, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID); err != nil {
		return 0, 0, false
	}
	ownerID, err := service.TeamManagerOwnerID(ctx, device.User.UserID, device.User.BusinessID)
	return device.User.UserID, ownerID, err == nil && ownerID != 0
}

func teamBusinessIDFromCtx(ctx context.Context) (int64, bool) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, false
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	return businessID, err == nil && businessID > 0
}
