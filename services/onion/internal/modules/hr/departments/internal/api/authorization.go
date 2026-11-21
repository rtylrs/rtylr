package api

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

const departmentManagePermission = "hr:employees:manage"

var hrDepartmentPermissionKeys = []string{"manage_departments", "manage_employees"}

// PermissionProvider exposes the hr_core employee/role lookups so members who
// hold the HR-internal HR Admin role (manage_departments / manage_employees)
// can manage departments alongside app-level hr:employees:manage holders.
type PermissionProvider interface {
	GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error)
	HasPermission(ctx context.Context, businessID, employeeID int64, permissionKey string) (bool, error)
}

func requireDepartmentManagePermission(ctx context.Context, permissions PermissionProvider, businessID int64) error {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil || !authPkg.HasAnyAppScope(device.User, authorization.AppHR) {
		return onekiterrors.PermissionDenied()
	}
	if authPkg.IsBusinessAdmin(device.User) ||
		authPkg.HasAppPermission(device.User, authorization.AppHR, departmentManagePermission) {
		return nil
	}
	if permissions != nil && businessID > 0 {
		employeeID, err := permissions.GetEmployeeIDByUserID(ctx, businessID, device.User.UserID)
		if err == nil && employeeID > 0 {
			for _, key := range hrDepartmentPermissionKeys {
				if allowed, permErr := permissions.HasPermission(ctx, businessID, employeeID, key); permErr == nil && allowed {
					return nil
				}
			}
		}
	}
	return onekiterrors.PermissionDenied()
}
