package stocktransfers

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type stockTransferToolPermission struct {
	app        authorization.AppID
	permission string
}

// stockTransferToolPermissions is the module's tool allowlist: it maps every
// tool this module dispatches to the permission the same operation requires on
// its HTTP twin, so the Caret surface cannot bypass route authorization. A tool
// absent from this map is never executed, which keeps a newly dispatched tool
// denied until its permission is declared here.
var stockTransferToolPermissions = map[string]stockTransferToolPermission{
	listStockTransfersToolName:   {app: authorization.AppERP, permission: "erp:inventory:view"},
	getStockTransferToolName:     {app: authorization.AppERP, permission: "erp:inventory:view"},
	createStockTransferToolName:  {app: authorization.AppERP, permission: "erp:inventory:transfer"},
	receiveStockTransferToolName: {app: authorization.AppERP, permission: "erp:inventory:transfer"},
}

func requireStockTransferToolPermission(ctx context.Context, required stockTransferToolPermission) error {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(authUser, required.app, required.permission) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
