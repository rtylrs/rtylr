package products

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type productToolPermission struct {
	app        authorization.AppID
	permission string
}

// productToolPermissions is the module's tool allowlist: it maps every tool
// this module dispatches to the permission the same operation requires on its
// HTTP twin, so the Caret surface cannot bypass route authorization. A tool
// absent from this map is never executed, which keeps a newly dispatched tool
// denied until its permission is declared here.
var productToolPermissions = map[string]productToolPermission{
	getProductToolName:             {app: authorization.AppERP, permission: "erp:products:view"},
	listReorderSuggestionsToolName: {app: authorization.AppERP, permission: "erp:products:view"},
	createProductToolName:          {app: authorization.AppERP, permission: "erp:products:create"},
	updateProductToolName:          {app: authorization.AppERP, permission: "erp:products:edit"},
	deleteProductToolName:          {app: authorization.AppERP, permission: "erp:products:delete"},
}

func requireProductToolPermission(ctx context.Context, required productToolPermission) error {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(authUser, required.app, required.permission) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
