package api

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

const onboardingManagePermission = "hr:employees:manage"

func requireOnboardingManagePermission(ctx context.Context) error {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil ||
		(!authPkg.HasAppPermission(device.User, authorization.AppHR, onboardingManagePermission) && !authPkg.IsBusinessAdmin(device.User)) ||
		!authPkg.HasAnyAppScope(device.User, authorization.AppHR) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
