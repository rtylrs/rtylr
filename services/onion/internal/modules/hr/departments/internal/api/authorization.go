package api

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

const departmentManagePermission = "hr:employees:manage"

// requireDepartmentManagePermission guards the department writes. Department
// structure is tenant org configuration, and these handlers previously
// authenticated the caller without checking a permission, so any member with HR
// app access could create, rename or delete departments across the tenant.
func requireDepartmentManagePermission(ctx context.Context) error {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil ||
		(!authPkg.HasAppPermission(device.User, authorization.AppHR, departmentManagePermission) && !authPkg.IsBusinessAdmin(device.User)) ||
		!authPkg.HasAnyAppScope(device.User, authorization.AppHR) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
