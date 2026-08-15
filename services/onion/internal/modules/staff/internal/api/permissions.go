package api

import "context"

// PermissionProvider is the small portion of hr_core needed to decide whether
// a staff response may include compensation fields.
type PermissionProvider interface {
	GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error)
	HasPermission(ctx context.Context, businessID, employeeID int64, permissionKey string) (bool, error)
}

// sensitiveDetailPermissionKeys gates every staff field a casual directory
// listing should not hand out by default: compensation (salary/hourly/
// overtime rate) and personal contact detail (home address, emergency
// contact). Both were previously ungated for compensation and unconditional
// for the contact fields - see newStaffPayload.
var sensitiveDetailPermissionKeys = []string{"manage_employees", "manage_payroll"}

// canViewSensitiveDetails fails closed when the caller cannot be resolved or
// the permission store is unavailable. The response builder then emits zero
// values, which the generated contract omits from JSON.
func canViewSensitiveDetails(
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

	for _, permissionKey := range sensitiveDetailPermissionKeys {
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
