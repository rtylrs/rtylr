package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/domain"
)

// compensationPermissionKeys are the role permission keys that grant visibility
// of another employee's compensation. Either one is sufficient: HR administers
// employee records, payroll administers what they are paid.
var compensationPermissionKeys = []string{permissionManageEmployees, permissionManagePayroll}

// canViewAllCompensation reports whether the caller may see compensation on
// employee records other than their own. It is resolved once per request: the
// caller's employee row is looked up once and each permission key is checked at
// most once, never per listed employee.
//
// It fails closed, so a degraded permission store hides compensation rather
// than exposing it.
func canViewAllCompensation(ctx context.Context, viewer permissionResolver, ownerID, viewerUserID int64) bool {
	return hasAnyPermission(ctx, viewer, ownerID, viewerUserID, compensationPermissionKeys...)
}

// maskCompensation returns a copy of the record with the compensation fields
// zeroed unless the caller is allowed to see them. The value receiver means the
// caller's record is never mutated, so cached records stay intact.
//
// Zeroing is the whole mechanism: SalaryCents, HourlyRateCents and
// OvertimeRateCents are converted to float64 dollars with `omitempty` on the
// contract type, so a zero cents value converts to a zero dollar value and is
// dropped from the JSON response entirely.
func maskCompensation(e domain.EmployeeRecord, viewerUserID int64, canViewAll bool) domain.EmployeeRecord {
	if canViewAll || isOwnRecord(e, viewerUserID) {
		return e
	}
	e.SalaryCents = 0
	e.HourlyRateCents = 0
	e.OvertimeRateCents = 0
	return e
}

// isOwnRecord reports whether the record belongs to the calling user. An
// employee row that is not linked to a platform user carries user_id 0 and must
// never match a zero viewer id.
func isOwnRecord(e domain.EmployeeRecord, viewerUserID int64) bool {
	return viewerUserID > 0 && e.UserID == viewerUserID
}

// ListEmployeesForViewer lists employees with compensation visible only on the
// records the caller is entitled to see. ownerID is the tenant scope, while
// viewerUserID identifies the calling platform user whose entitlement is being
// resolved.
//
// The gating deliberately runs after the cached read rather than inside it:
// ListEmployees caches per tenant, so masking inside the cache-fill would let
// whichever caller warmed the entry decide what every later caller sees.
func (s *Service) ListEmployeesForViewer(
	ctx context.Context,
	ownerID, viewerUserID int64,
	limit, offset int32,
) ([]domain.EmployeeRecord, int64, error) {
	return s.ListEmployeesForViewerWithFilters(ctx, ownerID, viewerUserID, domain.ListEmployeesParams{Limit: limit, Offset: offset})
}

// ListEmployeesForViewerWithFilters applies the same compensation gate as the
// unfiltered list while preserving the contract's optional tenant filters.
func (s *Service) ListEmployeesForViewerWithFilters(
	ctx context.Context,
	ownerID, viewerUserID int64,
	params domain.ListEmployeesParams,
) ([]domain.EmployeeRecord, int64, error) {
	employees, total, err := s.ListEmployeesWithFilters(ctx, ownerID, params)
	if err != nil {
		return nil, 0, err
	}
	return gateCompensation(ctx, s, ownerID, viewerUserID, employees), total, nil
}

// GetEmployeeForViewer fetches one employee with compensation visible only if
// the caller is entitled to it or the record is their own.
func (s *Service) GetEmployeeForViewer(
	ctx context.Context,
	ownerID, viewerUserID, employeeID int64,
) (*domain.EmployeeRecord, error) {
	emp, err := s.GetEmployee(ctx, ownerID, employeeID)
	if err != nil {
		return nil, err
	}
	if emp == nil {
		return nil, domain.ErrEmployeeNotFound
	}

	gated := maskCompensation(*emp, viewerUserID, canViewAllCompensation(ctx, s, ownerID, viewerUserID))
	return &gated, nil
}

// gateCompensation applies compensation visibility to a set of employee records.
// The permission decision is resolved once for the caller and then applied to
// every row with a single field comparison.
func gateCompensation(
	ctx context.Context,
	viewer permissionResolver,
	ownerID, viewerUserID int64,
	records []domain.EmployeeRecord,
) []domain.EmployeeRecord {
	canViewAll := canViewAllCompensation(ctx, viewer, ownerID, viewerUserID)

	gated := make([]domain.EmployeeRecord, 0, len(records))
	for _, e := range records {
		gated = append(gated, maskCompensation(e, viewerUserID, canViewAll))
	}
	return gated
}
