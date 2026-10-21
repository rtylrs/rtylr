package api

import (
	"context"
	"github.com/voxire/onion/internal/platform/authorization"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func requireLoyaltyManager(ctx context.Context) error {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return onekiterrors.PermissionDenied()
	}
	if authPkg.IsBusinessAdmin(device.User) || authPkg.HasAppPermission(device.User, authorization.AppCRM, "crm:loyalty:manage") {
		return nil
	}
	return onekiterrors.PermissionDenied()
}
