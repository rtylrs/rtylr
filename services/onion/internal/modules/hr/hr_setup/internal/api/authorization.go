package api

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

const hrSetupManagePermission = "hr:employees:manage"

// requireHRSetupManagePermission guards the HR setup writes. Completing a setup
// step and persisting the HR config both change tenant-wide configuration, and
// both previously authenticated the caller without checking a permission, so
// any member with HR app access could reconfigure the tenant.
func requireHRSetupManagePermission(ctx context.Context) error {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil ||
		(!authPkg.HasAppPermission(device.User, authorization.AppHR, hrSetupManagePermission) && !authPkg.IsBusinessAdmin(device.User)) ||
		!authPkg.HasAnyAppScope(device.User, authorization.AppHR) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
