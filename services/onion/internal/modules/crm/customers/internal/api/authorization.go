package api

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func requireCRMFeaturePermission(ctx context.Context, permission string) error {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !authPkg.HasAppPermission(device.User, authorization.AppCRM, permission) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}

func requireCRMCustomerListPermission(ctx context.Context) error {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return onekiterrors.PermissionDenied()
	}
	if authPkg.HasAppPermission(device.User, authorization.AppCRM, "crm:customers:view") {
		return nil
	}
	if authPkg.HasAppAccess(device.User, authorization.AppFinance) &&
		(authPkg.HasAppPermission(device.User, authorization.AppFinance, "finance:invoices:view") ||
			authPkg.HasAppPermission(device.User, authorization.AppFinance, "finance:invoices:create")) {
		return nil
	}
	return onekiterrors.PermissionDenied()
}
