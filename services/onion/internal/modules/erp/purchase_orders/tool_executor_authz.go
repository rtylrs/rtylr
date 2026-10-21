package purchase_orders

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type purchaseOrderToolPermission struct {
	app        authorization.AppID
	permission string
}

// purchaseOrderToolPermissions is the module's tool allowlist: it maps every
// tool this module dispatches to the permission the same operation requires on
// its HTTP twin, so the Caret surface cannot bypass route authorization. A tool
// absent from this map is never executed, which keeps a newly dispatched tool
// denied until its permission is declared here.
var purchaseOrderToolPermissions = map[string]purchaseOrderToolPermission{
	listPurchaseOrdersToolName:   {app: authorization.AppERP, permission: "erp:purchasing:view"},
	getPurchaseOrderToolName:     {app: authorization.AppERP, permission: "erp:purchasing:view"},
	createPurchaseOrderToolName:  {app: authorization.AppERP, permission: "erp:purchasing:create"},
	approvePurchaseOrderToolName: {app: authorization.AppERP, permission: "erp:purchasing:approve"},
	receivePurchaseOrderToolName: {app: authorization.AppERP, permission: "erp:purchasing:approve"},
}

func requirePurchaseOrderToolPermission(ctx context.Context, required purchaseOrderToolPermission) error {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(authUser, required.app, required.permission) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
