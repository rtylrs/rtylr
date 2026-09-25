package application

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

// employeeWriteGate is the lookup surface an HR administration write needs: the
// caller's role permissions, plus whether they own the business outright.
// *Service satisfies it; tests supply a fake so the gating rules can be
// exercised without a database.
type employeeWriteGate interface {
	permissionResolver
	IsBusinessOwner(ctx context.Context, userID int64) (bool, error)
}

// canManage reports whether the caller may perform an HR administration write
// guarded by permissionKey.
//
// The business owner passes without holding the permission, and that bypass is
// load-bearing rather than a convenience. Default roles are seeded as a side
// effect of the first CreateEmployee call, so a fresh tenant has no employee
// rows and therefore nobody holding manage_employees or manage_roles. Requiring
// the permission alone would leave the first employee permanently uncreatable.
// The owner is also the account that holds no hr_employees row of their own.
//
// It fails closed everywhere else: a caller who owns no business and holds no
// matching permission is denied, as is any caller whose lookups error.
func canManage(
	ctx context.Context,
	gate employeeWriteGate,
	ownerID, callerUserID int64,
	permissionKey string,
) bool {
	if callerUserID <= 0 {
		return false
	}
	if isBusinessOwner(ctx, gate, callerUserID) {
		return true
	}
	return hasAnyPermission(ctx, gate, ownerID, callerUserID, permissionKey)
}

// isBusinessOwner resolves business ownership, treating a failed lookup as "not
// the owner" so the caller falls through to the permission check. A degraded
// business store must not grant the write, but it also must not lock out an HR
// admin who would pass on their role alone.
func isBusinessOwner(ctx context.Context, gate employeeWriteGate, callerUserID int64) bool {
	owner, err := gate.IsBusinessOwner(ctx, callerUserID)
	if err != nil {
		logger.Ctx(ctx).Warn("failed to resolve business ownership for HR write gating",
			logger.Any("user_id", callerUserID), logger.Error(err))
		return false
	}
	return owner
}

// IsBusinessOwner reports whether the platform user owns an active business. A
// user who owns none is not an error, only a non-owner; anything else is
// reported so the gate can log it and fall through to the permission check.
func (s *Service) IsBusinessOwner(ctx context.Context, userID int64) (bool, error) {
	if userID <= 0 {
		return false, nil
	}

	if _, err := tenant.ResolveOwnedConfigBusinessID(ctx, userID, 0); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// --- Gated employee writes ---

// CreateEmployeeForWriter creates an employee once the caller is confirmed to
// administer employee records. ownerID is the tenant scope, callerUserID the
// platform user whose authority is being resolved.
func (s *Service) CreateEmployeeForWriter(
	ctx context.Context,
	ownerID, callerUserID int64,
	emp domain.EmployeeRecord,
) (*domain.EmployeeRecord, error) {
	if !canManage(ctx, s, ownerID, callerUserID, permissionManageEmployees) {
		return nil, domain.ErrPermissionDenied
	}
	return s.CreateEmployee(ctx, ownerID, emp)
}

// UpdateEmployeeForWriter updates an employee once the caller is confirmed to
// administer employee records. It guards compensation and status alike: the
// request body carries salary, hourly_rate, overtime_rate and status, so there
// is no lesser write to allow through.
func (s *Service) UpdateEmployeeForWriter(
	ctx context.Context,
	ownerID, callerUserID int64,
	emp domain.EmployeeRecord,
) (*domain.EmployeeRecord, error) {
	if !canManage(ctx, s, ownerID, callerUserID, permissionManageEmployees) {
		return nil, domain.ErrPermissionDenied
	}
	return s.UpdateEmployee(ctx, ownerID, emp)
}

// DeactivateEmployeeForWriter terminates an employee without exposing the
// full-row update path to a delete-shaped request.
func (s *Service) DeactivateEmployeeForWriter(ctx context.Context, ownerID, callerUserID, employeeID int64) error {
	if !canManage(ctx, s, ownerID, callerUserID, permissionManageEmployees) {
		return domain.ErrPermissionDenied
	}
	return s.DeactivateEmployee(ctx, ownerID, employeeID)
}

// InviteEmployeeForWriter invites an employee once the caller is confirmed to
// administer employee records. Inviting creates an hr_employees row, so it is
// gated as a create rather than as a message send.
func (s *Service) InviteEmployeeForWriter(
	ctx context.Context,
	ownerID, callerUserID int64,
	emp domain.EmployeeRecord,
) (string, error) {
	if !canManage(ctx, s, ownerID, callerUserID, permissionManageEmployees) {
		return "", domain.ErrPermissionDenied
	}
	return s.InviteEmployee(ctx, ownerID, emp)
}

// --- Gated role writes ---

// CreateRoleForWriter creates a role once the caller is confirmed to administer
// roles.
func (s *Service) CreateRoleForWriter(
	ctx context.Context,
	ownerID, callerUserID int64,
	role domain.RoleRecord,
) (*domain.RoleRecord, error) {
	if !canManage(ctx, s, ownerID, callerUserID, permissionManageRoles) {
		return nil, domain.ErrPermissionDenied
	}
	return s.CreateRole(ctx, ownerID, role)
}

// UpdateRoleForWriter updates a role once the caller is confirmed to administer
// roles. Role permissions live on the role row, so an ungated update is a way
// to grant yourself any permission by editing the role you already hold.
func (s *Service) UpdateRoleForWriter(
	ctx context.Context,
	ownerID, callerUserID int64,
	role domain.RoleRecord,
) (*domain.RoleRecord, error) {
	if !canManage(ctx, s, ownerID, callerUserID, permissionManageRoles) {
		return nil, domain.ErrPermissionDenied
	}
	return s.UpdateRole(ctx, ownerID, role)
}

// AssignRoleForWriter assigns a role once the caller is confirmed to administer
// roles. Without this gate the employee write gate is decorative: a denied
// caller could assign themselves the seeded HR Admin role and retry.
func (s *Service) AssignRoleForWriter(ctx context.Context, ownerID, callerUserID, employeeID, roleID int64) error {
	if !canManage(ctx, s, ownerID, callerUserID, permissionManageRoles) {
		return domain.ErrPermissionDenied
	}
	return s.AssignRole(ctx, ownerID, employeeID, roleID)
}
