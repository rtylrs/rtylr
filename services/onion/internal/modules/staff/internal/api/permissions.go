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

// permissionViewDirectory gates whether a caller may list the business-wide
// staff roster at all - a different question from which fields come back
// once they can. Every SeedDefaultRoles entry grants it by default, since
// most businesses want every employee to see the company directory; a
// business that wants to restrict it needs a real Roles/Permissions admin
// UI this codebase does not have working yet (see the setup wizard's
// discarded roles step).
const permissionViewDirectory = "view_directory"

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

// canViewDirectory fails closed the same way canViewSensitiveDetails does:
// an unresolvable caller or a permission-store error means no access, not a
// default allow.
func canViewDirectory(
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

	allowed, err := permissions.HasPermission(ctx, businessID, employeeID, permissionViewDirectory)
	if err != nil {
		return false
	}
	return allowed
}
