package customers

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type customerToolPermission struct {
	app        authorization.AppID
	permission string
}

// customerToolPermissions is the module's tool allowlist: it maps every tool
// this module dispatches to the permission the same operation requires on its
// HTTP twin, so the Caret surface cannot bypass route authorization. A tool
// absent from this map is never executed, which keeps a newly dispatched tool
// denied until its permission is declared here.
var customerToolPermissions = map[string]customerToolPermission{
	listCustomersToolName:  {app: authorization.AppCRM, permission: "crm:customers:view"},
	getCustomerToolName:    {app: authorization.AppCRM, permission: "crm:customers:view"},
	createCustomerToolName: {app: authorization.AppCRM, permission: "crm:customers:create"},
}

func requireCustomerToolPermission(ctx context.Context, required customerToolPermission) error {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(authUser, required.app, required.permission) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
