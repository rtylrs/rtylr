package handlers

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

func authenticatedUser(ctx context.Context) (*authPkg.AuthUser, bool) {
	if user, ok := authPkg.GetAuthUser(ctx); ok && user != nil && user.UserID > 0 {
		return user, true
	}

	device := authPkg.DeviceCtx(ctx)
	if device.User != nil && device.User.UserID > 0 {
		return device.User, true
	}

	return nil, false
}
