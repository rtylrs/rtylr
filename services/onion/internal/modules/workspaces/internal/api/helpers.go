package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func resolveWorkspaceQuery(ctx context.Context, svc *application.Service, authUser *authPkg.AuthUser, workspaceID string) (int64, error) {
	wsID := authUser.BusinessID
	if workspaceID == "" {
		return wsID, nil
	}
	if id, err := strconv.ParseInt(workspaceID, 10, 64); err == nil {
		return id, nil
	}
	resolved, err := svc.ResolveWorkspaceID(ctx, authUser.UserID, workspaceID)
	if err != nil || resolved == 0 {
		return 0, onekiterrors.New(http.StatusBadRequest, "Invalid workspace_id")
	}
	return resolved, nil
}
