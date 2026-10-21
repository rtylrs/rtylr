package expenses

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type expenseToolPermission struct {
	app        authorization.AppID
	permission string
}

// expenseToolPermissions is the module's tool allowlist: it maps every tool
// this module dispatches to the permission the same operation requires on its
// HTTP twin, so the Caret surface cannot bypass route authorization. A tool
// absent from this map is never executed, which keeps a newly dispatched tool
// denied until its permission is declared here.
var expenseToolPermissions = map[string]expenseToolPermission{
	listExpensesToolName:          {app: authorization.AppFinance, permission: "finance:expenses:view"},
	getExpenseToolName:            {app: authorization.AppFinance, permission: "finance:expenses:view"},
	listExpenseCategoriesToolName: {app: authorization.AppFinance, permission: "finance:expenses:view"},
	recordExpenseToolName:         {app: authorization.AppFinance, permission: "finance:expenses:create"},
}

func requireExpenseToolPermission(ctx context.Context, required expenseToolPermission) error {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(authUser, required.app, required.permission) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
