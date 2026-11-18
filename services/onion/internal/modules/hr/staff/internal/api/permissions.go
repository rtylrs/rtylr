package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/platform/onekiterrors"
)

// PermissionProvider is the small portion of hr_core needed to decide whether
// a staff response may include compensation fields, and separately whether
// the business-wide directory may be served at all.
type PermissionProvider interface {
	GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error)
	HasPermission(ctx context.Context, businessID, employeeID int64, permissionKey string) (bool, error)
	IsDirectoryEnabled(ctx context.Context, businessID int64) (bool, error)
}

// permissionViewDirectory mirrors hr_core's own unexported permission key of
// the same name (internal/modules/hr/hr_core/internal/application/permissions.go)
// - the same duplication already used by sensitiveDetailPermissionKeys below,
// since hr_core's permission keys are unexported and only reachable here as
// plain strings through the adapter.
const permissionViewDirectory = "view_directory"

// Stable codes the client uses to tell "an admin turned the directory off"
// apart from "your role can't see it". Both are 403s.
const (
	directoryDisabledCode          = "directory_disabled"
	directoryPermissionMissingCode = "directory_permission_missing"
)

type directoryAccessError struct {
	code    string
	message string
}

func (e *directoryAccessError) Error() string         { return e.message }
func (e *directoryAccessError) HTTPStatusCode() int   { return http.StatusForbidden }
func (e *directoryAccessError) PublicMessage() string { return e.message }
func (e *directoryAccessError) PublicCode() string    { return e.code }

var (
	errDirectoryDisabled = &directoryAccessError{
		code:    directoryDisabledCode,
		message: "The company directory is turned off for this business",
	}
	errDirectoryPermissionMissing = &directoryAccessError{
		code:    directoryPermissionMissingCode,
		message: onekiterrors.PermissionDeniedMessage,
	}
)

// checkDirectoryAccess returns nil when the caller may see the company-wide
// staff directory: the business must have it on, and the caller's role must
// hold view_directory. It fails closed - a caller with no resolvable employee
// row is denied, and a store error is a 500, never a pass.
func checkDirectoryAccess(ctx context.Context, permissions PermissionProvider, businessID, userID int64) error {
	if permissions == nil || userID <= 0 {
		return errDirectoryPermissionMissing
	}

	enabled, err := permissions.IsDirectoryEnabled(ctx, businessID)
	if err != nil {
		return onekiterrors.New(http.StatusInternalServerError, "Failed to check directory access")
	}
	if !enabled {
		return errDirectoryDisabled
	}

	employeeID, err := permissions.GetEmployeeIDByUserID(ctx, businessID, userID)
	if err != nil || employeeID <= 0 {
		return errDirectoryPermissionMissing
	}

	allowed, err := permissions.HasPermission(ctx, businessID, employeeID, permissionViewDirectory)
	if err != nil {
		return onekiterrors.New(http.StatusInternalServerError, "Failed to check directory access")
	}
	if !allowed {
		return errDirectoryPermissionMissing
	}
	return nil
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
