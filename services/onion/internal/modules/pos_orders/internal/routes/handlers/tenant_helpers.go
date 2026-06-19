package handlers

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func authenticatedPOSUser(ctx context.Context) (*authPkg.AuthUser, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, errorModels.PermissionDeniedServerError()
	}
	return device.User, nil
}

func resolveBusinessID(ctx context.Context, user *authPkg.AuthUser) (int64, *errorModels.ServerError) {
	businessID, err := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if err != nil {
		return 0, errorModels.PermissionDeniedServerError()
	}
	return businessID, nil
}
