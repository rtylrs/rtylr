package api

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func requireDashboardPermission(ctx context.Context) error {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil || !authPkg.HasDashboardPermission(device.User, "hub:dashboard:view") {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
