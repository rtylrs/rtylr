package api

import "context"

// PermissionProvider is the small portion of hr_core needed to decide whether
// a payroll response may include compensation fields.
type PermissionProvider interface {
	GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error)
	HasPermission(ctx context.Context, businessID, employeeID int64, permissionKey string) (bool, error)
}

var compensationPermissionKeys = []string{"manage_employees", "manage_payroll"}

// canViewCompensation fails closed when the caller cannot be resolved or the
// permission store is unavailable. The generated response omits zero-valued
// monetary fields from JSON.
func canViewCompensation(
	ctx context.Context,
	permissions PermissionProvider,
	businessID, userID int64,
) bool {
	if permissions == nil || userID <= 0 {
		return false
	}

	employeeID, err := permissions.GetEmployeeIDByUserID(ctx, businessID, userID)
	if err != nil || employeeID <= 0 {
		return false
	}

	for _, permissionKey := range compensationPermissionKeys {
		allowed, err := permissions.HasPermission(ctx, businessID, employeeID, permissionKey)
		if err != nil {
			return false
		}
		if allowed {
			return true
		}
	}
	return false
}
