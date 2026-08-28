package handlers

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func requireShiftPermission(ctx context.Context) error {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !authPkg.HasAppPermission(device.User, authorization.AppHR, "hr:schedules:manage") || !authPkg.HasAnyAppScope(device.User, authorization.AppHR) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
