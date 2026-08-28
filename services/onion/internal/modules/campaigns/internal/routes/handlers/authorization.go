package handlers

import (
	"context"
	"strings"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func requireCampaignPermission(ctx context.Context, permission string) error {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(device.User, authorization.AppCRM, permission) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}

func requireCampaignManager(ctx context.Context) error {
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
