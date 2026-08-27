package handlers

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func requireRecruitPermission(ctx context.Context, permission string) (*authPkg.AuthUser, error) {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok || user == nil || !authPkg.HasAppPermission(user, authorization.AppRecruit, permission) {
		return nil, onekiterrors.PermissionDenied()
	}
	return user, nil
}
