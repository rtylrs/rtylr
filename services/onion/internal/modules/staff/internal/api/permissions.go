package api

import "context"

// PermissionProvider is the small portion of hr_core needed to decide whether
// a staff response may include compensation fields, and separately whether
// the business-wide directory may be served at all.
type PermissionProvider interface {
	GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error)
	HasPermission(ctx context.Context, businessID, employeeID int64, permissionKey string) (bool, error)
	IsDirectoryEnabled(ctx context.Context, businessID int64) (bool, error)
}

// permissionViewDirectory mirrors hr_core's own unexported permission key of
// the same name (internal/modules/hr_core/internal/application/permissions.go)
// - the same duplication already used by sensitiveDetailPermissionKeys below,
// since hr_core's permission keys are unexported and only reachable here as
// plain strings through the adapter.
const permissionViewDirectory = "view_directory"

// canViewDirectory reports whether the caller may see the company-wide staff
// directory: the business must have turned the feature on, and the caller's
// role must hold view_directory. Both gates fail closed - a caller with no
// resolvable employee row, or a permission-store error, is denied rather than
// let through, matching canViewSensitiveDetails below.
func canViewDirectory(ctx context.Context, permissions PermissionProvider, businessID, userID int64) bool {
	if permissions == nil || userID <= 0 {
		return false
	}

	enabled, err := permissions.IsDirectoryEnabled(ctx, businessID)
	if err != nil || !enabled {
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
