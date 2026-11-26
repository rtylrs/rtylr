package api

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
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

func requirePOSPermission(ctx context.Context, permission string) (*authPkg.AuthUser, error) {
	user, err := authenticatedPOSUser(ctx)
	if err != nil {
		return nil, err
	}
	if !authPkg.HasAppPermission(user, authorization.AppPOS, permission) || !authPkg.HasAnyAppScope(user, authorization.AppPOS) {
		return nil, onekiterrors.PermissionDenied()
	}
	return user, nil
}

// requirePOSOrMenuPermission grants access to POS staff with the POS permission
// OR standalone Digital Menu users with the matching Menu permission so menu-only
// workspaces can view and manage incoming menu orders in Portal without POS.
func requirePOSOrMenuPermission(ctx context.Context, posPermission, menuPermission string) (*authPkg.AuthUser, error) {
	user, err := authenticatedPOSUser(ctx)
	if err != nil {
		return nil, err
	}
	hasPOS := authPkg.HasAppPermission(user, authorization.AppPOS, posPermission) &&
		authPkg.HasAnyAppScope(user, authorization.AppPOS)
	hasMenu := menuPermission != "" && authPkg.HasAppPermission(user, authorization.AppMenu, menuPermission)
	if !hasPOS && !hasMenu {
		return nil, onekiterrors.PermissionDenied()
	}
	return user, nil
}

func resolveBusinessID(ctx context.Context, user *authPkg.AuthUser) (int64, error) {
	businessID, err := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if err != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}
