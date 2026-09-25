package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// Role permission keys seeded by SeedDefaultRoles and stored on hr_roles.
const (
	permissionManageEmployees = "manage_employees"
	permissionManagePayroll   = "manage_payroll"
	permissionManageRoles     = "manage_roles"
	permissionViewDirectory   = "view_directory"
)

// permissionResolver resolves which hr_employees row belongs to the calling
// platform user and checks that employee's role permissions. *Service satisfies
// it; tests supply a fake so the gating rules can be exercised without a
// database.
type permissionResolver interface {
	GetEmployeeIDByUserID(ctx context.Context, ownerID, userID int64) (int64, error)
	HasPermission(ctx context.Context, ownerID, employeeID int64, permissionKey string) (bool, error)
}

// hasAnyPermission reports whether the calling user's employee row holds at
// least one of the given keys. The caller's row is resolved once and each key
// is checked at most once, stopping at the first grant.
//
// It fails closed. A caller with no employee row (so no role) and any lookup
// error both yield false, so a degraded permission store denies rather than
// allows.
func hasAnyPermission(
	ctx context.Context,
	resolver permissionResolver,
	ownerID, callerUserID int64,
	keys ...string,
) bool {
	if callerUserID <= 0 {
		return false
	}

	employeeID, err := resolver.GetEmployeeIDByUserID(ctx, ownerID, callerUserID)
	if err != nil {
		if !errors.Is(err, domain.ErrEmployeeNotFound) {
			logger.Ctx(ctx).Warn("failed to resolve caller employee row for permission check",
				logger.Any("owner_id", ownerID), logger.Any("user_id", callerUserID), logger.Error(err))
		}
		return false
	}

	for _, key := range keys {
		allowed, permErr := resolver.HasPermission(ctx, ownerID, employeeID, key)
		if permErr != nil {
			logger.Ctx(ctx).Warn("failed to check role permission",
				logger.Any("owner_id", ownerID), logger.Any("employee_id", employeeID),
				logger.Any("permission_key", key), logger.Error(permErr))
			return false
		}
		if allowed {
			return true
		}
	}
	return false
}
