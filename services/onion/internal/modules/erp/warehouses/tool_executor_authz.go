package warehouses

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type warehouseToolPermission struct {
	app        authorization.AppID
	permission string
}

// warehouseToolPermissions is the module's tool allowlist: it maps every tool
// this module dispatches to the permission the same operation requires on its
// HTTP twin, so the Caret surface cannot bypass route authorization. A tool
// absent from this map is never executed, which keeps a newly dispatched tool
// denied until its permission is declared here.
var warehouseToolPermissions = map[string]warehouseToolPermission{
	getWarehouseToolName:    {app: authorization.AppERP, permission: "erp:warehouses:manage"},
	listWarehousesToolName:  {app: authorization.AppERP, permission: "erp:warehouses:manage"},
	createWarehouseToolName: {app: authorization.AppERP, permission: "erp:warehouses:manage"},
}

func requireWarehouseToolPermission(ctx context.Context, required warehouseToolPermission) error {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(authUser, required.app, required.permission) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
