package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/platform/authorization"
)

func (r *Repository) replaceUserGroupMemberships(ctx context.Context, tx *sql.Tx, businessID, ownerID, actorID, userID int64, groupIDs []int64) error {
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM user_group_members
		WHERE user_id = ? AND owner_id = ? AND (business_id = ? OR business_id = 0)`, userID, ownerID, businessID); err != nil {
		return err
	}
	for _, groupID := range uniqueInt64(groupIDs) {
		var found int
		if err := tx.QueryRowContext(ctx, `
			SELECT 1
			FROM user_groups
			WHERE id = ? AND owner_id = ? AND (business_id = ? OR business_id = 0)
			LIMIT 1`, groupID, ownerID, businessID).Scan(&found); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("group %d does not belong to this business", groupID)
			}
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO user_group_members (group_id, user_id, owner_id, business_id, added_by)
			VALUES (?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), added_by = VALUES(added_by)`,
			groupID, userID, ownerID, businessID, actorID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) insertLicenseAssignments(ctx context.Context, tx *sql.Tx, businessID, actorID, userID int64, appID string, assignments []domain.AppRoleAssignmentInput) error {
	matched := false
	for _, assignment := range assignmentsForApp(assignments, appID) {
		if assignment.Status != "" && assignment.Status != "active" {
			continue
		}
		source := authorization.SourceExplicit
		if assignment.GroupID == nil {
			// Per-user assignments are written once below, after every group
			// assignment has been handled.
			continue
		}
		// Assignment source is server-owned metadata: a group license row must
		// say 'group'. The resolver rejects an explicit row that carries a
		// group id, so writing 'explicit' here silently denied the app to every
		// member of the group.
		source = authorization.SourceGroup
		var validGroupID int64
		if err := tx.QueryRowContext(ctx, `
			SELECT id FROM user_groups
			WHERE id = ? AND (business_id = ? OR business_id = 0)
			LIMIT 1`, *assignment.GroupID, businessID).Scan(&validGroupID); err != nil {
			return fmt.Errorf("group license assignment is not valid: %w", err)
		}
		matched = true
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO authorization_license_assignments
				(business_id, user_id, group_id, app_id, source, status, granted_by)
			VALUES (?, NULL, ?, ?, ?, 'active', ?)`, businessID, *assignment.GroupID, appID, source, actorID); err != nil {
			return err
		}
	}
	if matched {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO authorization_license_assignments
			(business_id, user_id, app_id, source, status, granted_by)
		VALUES (?, ?, ?, 'explicit', 'active', ?)`, businessID, userID, appID, actorID)
	return err
}

func uniqueInt64(values []int64) []int64 {
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		found := false
		for _, existing := range result {
			if existing == value {
				found = true
				break
			}
		}
		if !found {
			result = append(result, value)
		}
	}
	return result
}

func (r *Repository) lookupAuthorizationRoleID(ctx context.Context, tx *sql.Tx, businessID int64, appID, roleKey string) (int64, error) {
	var roleID int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM authorization_roles
		WHERE business_id IN (0, ?) AND app_id = ? AND role_key = ?
		  AND lifecycle = 'active' AND is_archived = FALSE
		ORDER BY business_id DESC
		LIMIT 1`, businessID, appID, roleKey).Scan(&roleID)
	return roleID, err
}

func (r *Repository) validateAuthorizationRoleID(ctx context.Context, tx *sql.Tx, businessID int64, appID string, roleID int64) error {
	var found int
	return tx.QueryRowContext(ctx, `
		SELECT 1
		FROM authorization_roles
		WHERE id = ? AND business_id IN (0, ?) AND app_id = ?
		  AND lifecycle = 'active' AND is_archived = FALSE
		LIMIT 1`, roleID, businessID, appID).Scan(&found)
}

func assignmentsForApp(values []domain.AppRoleAssignmentInput, appID string) []domain.AppRoleAssignmentInput {
	result := make([]domain.AppRoleAssignmentInput, 0)
	for _, value := range values {
		if value.AppID == appID {
			result = append(result, value)
		}
	}
	return result
}

func uniqueAuthorizationApps(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = appendUniqueAuthorizationApp(result, value)
	}
	return result
}

func appendUniqueAuthorizationApp(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func isTenantAppID(appID string) bool {
	switch appID {
	case "pos", "erp", "hr", "crm", "finance", "flow", "insights", "menu", "recruit":
		return true
	default:
		return false
	}
}

func (r *Repository) DeactivateUser(ctx context.Context, businessID, ownerID, actorID, userID int64) error {
	if businessID <= 0 || ownerID <= 0 || actorID <= 0 || userID <= 0 {
		return domain.ErrUserNotFound
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(cause error) error { _ = tx.Rollback(); return cause }
	result, err := tx.ExecContext(ctx, `
		UPDATE authorization_business_members member
		JOIN businesses business ON business.id = member.business_id
		SET member.status = 'revoked', member.authorization_version = member.authorization_version + 1, member.updated_at = NOW()
		WHERE business.id = ? AND business.owner_id = ? AND member.user_id = ? AND member.status <> 'revoked'`, businessID, ownerID, userID)
	if err != nil {
		return rollback(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return rollback(err)
	}
	if rows == 0 {
		return rollback(domain.ErrUserNotFound)
	}
	for _, query := range []string{
		`UPDATE authorization_license_assignments SET status = 'revoked', revoked_at = NOW(), revoked_by = ? WHERE business_id = ? AND user_id = ? AND status = 'active'`,
		`UPDATE authorization_role_assignments SET status = 'revoked', revoked_at = NOW(), revoked_by = ? WHERE business_id = ? AND user_id = ? AND status = 'active'`,
		`DELETE FROM user_group_members WHERE business_id = ? AND user_id = ?`,
	} {
		var execErr error
		if query == `DELETE FROM user_group_members WHERE business_id = ? AND user_id = ?` {
			_, execErr = tx.ExecContext(ctx, query, businessID, userID)
		} else {
			_, execErr = tx.ExecContext(ctx, query, actorID, businessID, userID)
		}
		if execErr != nil {
			return rollback(execErr)
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO authorization_audit_events (business_id, actor_user_id, target_user_id, action, source_client) VALUES (?, ?, ?, 'member.deactivated', 'dash')`, businessID, actorID, userID); err != nil {
		return rollback(err)
	}
	return tx.Commit()
}

// RevokeBusinessUserPermissions deactivates one user's permission rows in one
// business when hr_core marks that user's employee record terminated. The
// employment decision stays in hr_employees; this module only owns the write.
func (r *Repository) RevokeBusinessUserPermissions(ctx context.Context, businessID, userID int64) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	rollback := func(cause error) (int64, error) { _ = tx.Rollback(); return 0, cause }
	var changed int64
	for _, query := range []string{
		`UPDATE authorization_business_members SET status = 'revoked', authorization_version = authorization_version + 1, updated_at = NOW() WHERE business_id = ? AND user_id = ? AND status <> 'revoked'`,
		`UPDATE authorization_license_assignments SET status = 'revoked', revoked_at = NOW() WHERE business_id = ? AND user_id = ? AND status = 'active'`,
		`UPDATE authorization_role_assignments SET status = 'revoked', revoked_at = NOW() WHERE business_id = ? AND user_id = ? AND status = 'active'`,
	} {
		result, execErr := tx.ExecContext(ctx, query, businessID, userID)
		if execErr != nil {
			return rollback(execErr)
		}
		count, countErr := result.RowsAffected()
		if countErr != nil {
			return rollback(countErr)
		}
		changed += count
	}
	if changed > 0 {
		if _, err := tx.ExecContext(ctx, `INSERT INTO authorization_audit_events (business_id, target_user_id, action, reason, source_client) VALUES (?, ?, 'member.authorization.revoked.termination', 'HR employee terminated', 'app')`, businessID, userID); err != nil {
			return rollback(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return changed, nil
}
