package invoices

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type invoiceToolPermission struct {
	app        authorization.AppID
	permission string
}

// invoiceToolPermissions is the module's tool allowlist: it maps every tool
// this module dispatches to the permission the same operation requires on its
// HTTP twin, so the Caret surface cannot bypass route authorization. A tool
// absent from this map is never executed, which keeps a newly dispatched tool
// denied until its permission is declared here.
var invoiceToolPermissions = map[string]invoiceToolPermission{
	listInvoicesToolName:  {app: authorization.AppFinance, permission: "finance:invoices:view"},
	getInvoiceToolName:    {app: authorization.AppFinance, permission: "finance:invoices:view"},
	createInvoiceToolName: {app: authorization.AppFinance, permission: "finance:invoices:create"},
	recordPaymentToolName: {app: authorization.AppFinance, permission: "finance:invoices:create"},
	sendInvoiceToolName:   {app: authorization.AppFinance, permission: "finance:invoices:send"},
}

func requireInvoiceToolPermission(ctx context.Context, required invoiceToolPermission) error {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(authUser, required.app, required.permission) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
