package api

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func requireCRMFeaturePermission(ctx context.Context, permission string) error {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !authPkg.HasAppPermission(device.User, authorization.AppCRM, permission) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
