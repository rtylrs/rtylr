package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// ReconcileTerminatedEmployeeAccess sweeps every business and deactivates any
// still-active permission row whose user is a terminated employee of that
// same business. Idempotent; the daily job runs it so a missed synchronous
// revoke is corrected within a day.
func (r *Repository) ReconcileTerminatedEmployeeAccess(ctx context.Context) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	rollback := func(cause error) (int64, error) { _ = tx.Rollback(); return 0, cause }
	var changed int64
	for _, query := range []string{
		`UPDATE authorization_business_members member JOIN hr_employees employee ON employee.user_id = member.user_id AND employee.business_id = member.business_id SET member.status = 'revoked', member.authorization_version = member.authorization_version + 1, member.updated_at = NOW() WHERE employee.status = 'terminated' AND member.status <> 'revoked'`,
		`UPDATE authorization_license_assignments license JOIN hr_employees employee ON employee.user_id = license.user_id AND employee.business_id = license.business_id SET license.status = 'revoked', license.revoked_at = NOW() WHERE employee.status = 'terminated' AND license.status = 'active'`,
		`UPDATE authorization_role_assignments assignment JOIN hr_employees employee ON employee.user_id = assignment.user_id AND employee.business_id = assignment.business_id SET assignment.status = 'revoked', assignment.revoked_at = NOW() WHERE employee.status = 'terminated' AND assignment.status = 'active'`,
	} {
		result, execErr := tx.ExecContext(ctx, query)
		if execErr != nil {
			return rollback(execErr)
		}
		count, countErr := result.RowsAffected()
		if countErr != nil {
			return rollback(countErr)
		}
		changed += count
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return changed, nil
}

// ─── Users ────────────────────────────────────────────────────────────────────

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (int64, string, string, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return 0, "", "", err
	}

	return row.ID, row.Email, row.FullName, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (string, string, bool, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", false, domain.ErrUserNotFound
		}

		return "", "", false, err
	}

	return row.Email, row.FullName, row.IsConsoleAdmin, nil
}

func (r *Repository) GetUserBusinessID(ctx context.Context, id int64) (int64, error) {
	businessID, err := r.queries.GetUserBusinessID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrUserNotFound
		}
		return 0, err
	}
	if !businessID.Valid {
		return 0, nil
	}
	return businessID.Int64, nil
}

func (r *Repository) CreateUser(ctx context.Context, email, username, passwordHash, fullName string) (int64, error) {
	res, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        email,
		Username:     sql.NullString{String: username, Valid: username != ""},
		PasswordHash: passwordHash,
		FullName:     sql.NullString{String: fullName, Valid: fullName != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create user via hub", logger.Error(err))

		return 0, err
	}

	return res.LastInsertId()
}

// DeleteInvitedUser removes the rows created while provisioning a new invitee.
// It is used as compensation when a later invite step fails because CreateUser
// is not part of the normalized authorization transaction.
func (r *Repository) DeleteInvitedUser(ctx context.Context, businessID, ownerID, userID int64) error {
	if userID <= 0 {
		return nil
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(cause error) error {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			return fmt.Errorf("%w; rollback failed: %w", cause, rollbackErr)
		}
		return cause
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
				logger.Ctx(ctx).Error("failed to rollback invited user cleanup transaction", logger.Error(rollbackErr))
			}
			panic(recovered)
		}
	}()

	statements := []struct {
		name  string
		query string
		args  []any
	}{
		{
			name: "group memberships",
			query: `DELETE FROM user_group_members
				WHERE user_id = ? AND owner_id = ? AND (business_id = ? OR business_id = 0)`,
			args: []any{userID, ownerID, businessID},
		},
		{
			name: "role assignment scopes",
			query: `DELETE scopes
				FROM authorization_role_assignment_scopes scopes
				JOIN authorization_role_assignments assignments
				  ON assignments.id = scopes.assignment_id
				WHERE assignments.business_id = ? AND assignments.user_id = ?`,
			args: []any{businessID, userID},
		},
		{
			name: "role assignments",
			query: `DELETE FROM authorization_role_assignments
				WHERE business_id = ? AND user_id = ?`,
			args: []any{businessID, userID},
		},
		{
			name: "license assignments",
			query: `DELETE FROM authorization_license_assignments
				WHERE business_id = ? AND user_id = ?`,
			args: []any{businessID, userID},
		},
		{
			name: "authorization audit events",
			query: `DELETE FROM authorization_audit_events
				WHERE business_id = ? AND actor_user_id = ? AND target_user_id = ?
				  AND action = 'member.authorization.updated'`,
			args: []any{businessID, ownerID, userID},
		},
		{
			name: "authorization membership",
			query: `DELETE FROM authorization_business_members
				WHERE business_id = ? AND user_id = ?`,
			args: []any{businessID, userID},
		},
		{
			name: "provisioning state",
			query: `DELETE FROM user_provisioning_states
				WHERE business_id = ? AND user_id = ?`,
			args: []any{businessID, userID},
		},
		{
			name: "invite tokens",
			query: `DELETE FROM invite_tokens
				WHERE owner_id = ? AND user_id = ?`,
			args: []any{ownerID, userID},
		},
		{
			name:  "MFA enrollments",
			query: `DELETE FROM user_mfa_enrollments WHERE user_id = ?`,
			args:  []any{userID},
		},
		{
			name: "employee record",
			query: `DELETE FROM hr_employees
				WHERE owner_id = ? AND business_id = ? AND user_id = ?`,
			args: []any{ownerID, businessID, userID},
		},
		{
			name: "staff directory record",
			query: `DELETE FROM hr_staff
				WHERE owner_id = ? AND business_id = ? AND user_id = ?`,
			args: []any{ownerID, businessID, userID},
		},
		{
			name:  "user",
			query: `DELETE FROM users WHERE id = ?`,
			args:  []any{userID},
		},
	}

	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement.query, statement.args...); err != nil {
			return rollback(fmt.Errorf("delete invited user %s: %w", statement.name, err))
		}
	}

	return tx.Commit()
}

// AttachUserToBusiness fills users.business_id for a member who does not have
// one yet, and reports whether it applied. It deliberately leaves an existing
// value alone: a person can belong to more than one business, and business_id is
// the one their session resolves to, so inviting somebody who already has one
// must not move them out of it.
func (r *Repository) AttachUserToBusiness(ctx context.Context, userID, businessID int64) (bool, error) {
	rows, err := r.queries.SetUserBusinessIDIfUnset(ctx, sqlc.SetUserBusinessIDIfUnsetParams{
		BusinessID: sql.NullInt64{Int64: businessID, Valid: businessID != 0},
		ID:         userID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to attach user to business",
			logger.Int("user_id", int(userID)), logger.Int("business_id", int(businessID)), logger.Error(err))
		return false, err
	}
	return rows > 0, nil
}

// EnsureEmployeeForUser creates the HR employee record for a user. The business
// is passed in rather than read from the user's own row: at invite time
// users.business_id is still NULL, so reading it there scoped the record to no
// business at all.
func (r *Repository) EnsureEmployeeForUser(ctx context.Context, ownerID, businessID, userID int64) error {
	return r.queries.EnsureEmployeeForUser(ctx, sqlc.EnsureEmployeeForUserParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		ID:         userID,
	})
}

// EnsureStaffForUser mirrors an invited employee in hr_staff, which is the
// source used by the Directory. Existing staff rows are retained and linked
// to their employee so this also repairs rows created before the mirror ran.
func (r *Repository) EnsureStaffForUser(ctx context.Context, ownerID, businessID, userID int64, role string) error {
	if err := r.queries.LinkStaffEmployeeForUser(ctx, sqlc.LinkStaffEmployeeForUserParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		UserID:     sql.NullInt64{Int64: userID, Valid: true},
	}); err != nil {
		return err
	}

	return r.queries.EnsureStaffForUser(ctx, sqlc.EnsureStaffForUserParams{
		Role:       role,
		OwnerID:    ownerID,
		BusinessID: businessID,
		UserID:     sql.NullInt64{Int64: userID, Valid: true},
	})
}

func (r *Repository) UpdateUserProfile(ctx context.Context, userID int64, fullName string) error {
	err := r.queries.UpdateUserProfile(ctx, sqlc.UpdateUserProfileParams{
		FullName: sql.NullString{String: strings.TrimSpace(fullName), Valid: strings.TrimSpace(fullName) != ""},
		ID:       userID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update user profile via team", logger.Error(err))
		return err
	}
	return nil
}
