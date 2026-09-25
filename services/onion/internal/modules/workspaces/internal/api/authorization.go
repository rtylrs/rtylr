package api

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func requireFlowPermission(ctx context.Context, permission string) error {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok || user == nil || !authPkg.HasAppPermission(user, authorization.AppFlow, permission) || !authPkg.HasAnyAppScope(user, authorization.AppFlow) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
