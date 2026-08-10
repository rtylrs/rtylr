package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// compensationPermissionKeys are the role permission keys that grant visibility
// of another employee's compensation. Either one is sufficient: HR administers
// employee records, payroll administers what they are paid.
var compensationPermissionKeys = []string{"manage_employees", "manage_payroll"}

// compensationViewer resolves which hr_employees row belongs to the calling
// platform user and checks that employee's role permissions. *Service satisfies
// it; tests supply a fake so the gating rules can be exercised without a
// database.
type compensationViewer interface {
	GetEmployeeIDByUserID(ctx context.Context, ownerID, userID int64) (int64, error)
	HasPermission(ctx context.Context, ownerID, employeeID int64, permissionKey string) (bool, error)
}

// canViewAllCompensation reports whether the caller may see compensation on
// employee records other than their own. It is resolved once per request: the
// caller's employee row is looked up once and each permission key is checked at
// most once, never per listed employee.
//
// It fails closed. A caller with no employee row (so no role) and any lookup
// error both yield false, so a degraded permission store hides compensation
// rather than exposing it.
func canViewAllCompensation(ctx context.Context, viewer compensationViewer, ownerID, viewerUserID int64) bool {
	viewerEmployeeID, err := viewer.GetEmployeeIDByUserID(ctx, ownerID, viewerUserID)
	if err != nil {
		if !errors.Is(err, domain.ErrEmployeeNotFound) {
			logger.Ctx(ctx).Warn("failed to resolve caller employee row for compensation gating",
				logger.Any("owner_id", ownerID), logger.Any("user_id", viewerUserID), logger.Error(err))
		}
		return false
	}

	for _, key := range compensationPermissionKeys {
		allowed, permErr := viewer.HasPermission(ctx, ownerID, viewerEmployeeID, key)
		if permErr != nil {
			logger.Ctx(ctx).Warn("failed to check compensation permission",
				logger.Any("owner_id", ownerID), logger.Any("employee_id", viewerEmployeeID),
				logger.Any("permission_key", key), logger.Error(permErr))
			return false
		}
		if allowed {
			return true
		}
	}
	return false
}

// maskCompensation returns a copy of the record with the compensation fields
// zeroed unless the caller is allowed to see them. The value receiver means the
// caller's record is never mutated, so cached records stay intact.
//
// Zeroing is the whole mechanism: Salary, HourlyRate and OvertimeRate are
// float64 with `omitempty` on the contract type, so a zero value is dropped from
// the JSON response entirely.
func maskCompensation(e domain.EmployeeRecord, viewerUserID int64, canViewAll bool) domain.EmployeeRecord {
	if canViewAll || isOwnRecord(e, viewerUserID) {
		return e
	}
	e.Salary = 0
	e.HourlyRate = 0
	e.OvertimeRate = 0
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
	employees, total, err := s.ListEmployees(ctx, ownerID, limit, offset)
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
	viewer compensationViewer,
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
