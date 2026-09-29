package workspaces

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type flowToolPermission struct {
	app        authorization.AppID
	permission string
}

// flowToolPermissions is the module's tool allowlist: it maps every tool this
// module dispatches to the permission the same operation requires on its HTTP
// twin, so the Caret surface cannot bypass route authorization. A tool absent
// from this map is never executed, which keeps a newly dispatched tool denied
// until its permission is declared here.
var flowToolPermissions = map[string]flowToolPermission{
	listTasksToolName:  {app: authorization.AppFlow, permission: "flow:tasks:view"},
	getTaskToolName:    {app: authorization.AppFlow, permission: "flow:tasks:view"},
	createTaskToolName: {app: authorization.AppFlow, permission: "flow:tasks:create"},
}

// requireFlowToolPermission mirrors requireFlowPermission in internal/api,
// which also requires a normalized Flow assignment scope.
func requireFlowToolPermission(ctx context.Context, required flowToolPermission) error {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(authUser, required.app, required.permission) ||
		!authPkg.HasAnyAppScope(authUser, required.app) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
