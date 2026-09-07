package api

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func requirePayrollPermission(ctx context.Context, permission string) error {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(device.User, authorization.AppHR, permission) || !authPkg.HasAnyAppScope(device.User, authorization.AppHR) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
