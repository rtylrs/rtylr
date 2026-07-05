package handlers

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func authenticatedPOSUser(ctx context.Context) (*authPkg.AuthUser, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	return device.User, nil
}

func resolveBusinessID(ctx context.Context, user *authPkg.AuthUser) (int64, error) {
	businessID, err := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if err != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}
