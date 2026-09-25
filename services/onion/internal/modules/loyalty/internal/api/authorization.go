package api

import (
	"context"
	"strings"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func requireLoyaltyManager(ctx context.Context) error {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return onekiterrors.PermissionDenied()
	}
	switch strings.ToLower(strings.TrimSpace(device.User.Role)) {
	case "owner", "admin", "console_admin":
		return nil
	default:
		return onekiterrors.PermissionDenied()
	}
}
