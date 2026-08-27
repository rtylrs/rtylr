package handlers

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func requireHRPermission(ctx context.Context, permission string) error {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !authPkg.HasAppPermission(device.User, authorization.AppHR, permission) || !authPkg.HasAnyAppScope(device.User, authorization.AppHR) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
