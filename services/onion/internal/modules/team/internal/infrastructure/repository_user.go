package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) replaceUserGroupMemberships(ctx context.Context, tx *sql.Tx, businessID, ownerID, userID int64, groupIDs []int64) error {
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
			groupID, userID, ownerID, businessID, ownerID); err != nil {
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

func isDashRole(role string) bool {
	switch role {
	case "owner", "admin", "manager", "member", "viewer":
		return true
	default:
		return false
	}
}

func legacyDashRole(role domain.Role) string {
	switch role {
	case domain.RoleOwner:
		return "owner"
	case domain.RoleAdmin:
		return "admin"
	case domain.RoleOperator, domain.RoleBilling, domain.RoleMenuManager, domain.RoleStoreManager:
		return "manager"
	case domain.RoleViewer:
		return "viewer"
	default:
		return "member"
	}
}

func appRoleKey(appID string, role domain.Role) string {
	suffix := "viewer"
	switch role {
	case domain.RoleOwner, domain.RoleAdmin:
		suffix = "admin"
	case domain.RoleOperator, domain.RoleStoreManager, domain.RoleMenuManager:
		// Insights has no manager role in the catalog (viewer/analyst/admin
		// only) and the legacy backfill maps these roles to viewer, so follow
		// it instead of looking up a key that does not exist.
		if appID == "insights" {
			suffix = "viewer"
		} else {
			suffix = "manager"
		}
	case domain.RoleStaff:
		switch appID {
		case "pos":
			suffix = "cashier"
		case "flow":
			suffix = "contributor"
		case "hr":
			suffix = "employee"
		}
	}
	if appID == "pos" && (role == domain.RoleOperator || role == domain.RoleStoreManager || role == domain.RoleMenuManager) {
		suffix = "location_manager"
	}
	return appID + "." + suffix
}

func (r *Repository) DeactivateUser(ctx context.Context, ownerID, userID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to deactivate user", logger.Error(err))
		return err
	}
	if err := sqlc.New(tx).DeactivateUser(ctx, sqlc.DeactivateUserParams{OwnerID: ownerID, UserID: userID}); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_business_members
		SET status = 'revoked', authorization_version = authorization_version + 1, updated_at = NOW()
		WHERE business_id IN (SELECT id FROM businesses WHERE owner_id = ?) AND user_id = ? AND status <> 'revoked'`, ownerID, userID); err != nil && !authorizationSchemaUnavailable(err) {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_license_assignments l
		JOIN businesses b ON b.id = l.business_id
		SET l.status = 'revoked', l.revoked_at = NOW()
		WHERE b.owner_id = ? AND l.user_id = ? AND l.status = 'active'`, ownerID, userID); err != nil && !authorizationSchemaUnavailable(err) {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_role_assignments a
		JOIN businesses b ON b.id = a.business_id
		SET a.status = 'revoked', a.revoked_at = NOW()
		WHERE b.owner_id = ? AND a.user_id = ? AND a.status = 'active'`, ownerID, userID); err != nil && !authorizationSchemaUnavailable(err) {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// RevokeBusinessUserPermissions deactivates one user's permission rows in one
// business when hr_core marks that user's employee record terminated. The
// employment decision stays in hr_employees; this module only owns the write.
func (r *Repository) RevokeBusinessUserPermissions(ctx context.Context, businessID, userID int64) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("failed to revoke permissions for terminated employee",
			logger.Any("business_id", businessID), logger.Any("user_id", userID), logger.Error(err))
		return 0, err
	}
	rollback := func(cause error) (int64, error) {
		_ = tx.Rollback()
		if authorizationSchemaUnavailable(cause) {
			rows, legacyErr := r.queries.RevokeBusinessUserPermissions(ctx, sqlc.RevokeBusinessUserPermissionsParams{
				BusinessID:         nullBusinessID(businessID),
				EmployeeBusinessID: businessID,
				UserID:             userID,
			})
			return rows, legacyErr
		}
		return 0, cause
	}
	// The two business-id parameters are the same business. They are separate
	// because the two columns differ in nullability; see the note on the query.
	rows, err := sqlc.New(tx).RevokeBusinessUserPermissions(ctx, sqlc.RevokeBusinessUserPermissionsParams{
		BusinessID:         nullBusinessID(businessID),
		EmployeeBusinessID: businessID,
		UserID:             userID,
	})
	if err != nil {
		return rollback(err)
	}
	if result, normalizedErr := tx.ExecContext(ctx, `
		UPDATE authorization_business_members
		SET status = 'revoked', authorization_version = authorization_version + 1, updated_at = NOW()
		WHERE business_id = ? AND user_id = ? AND status <> 'revoked'`, businessID, userID); normalizedErr != nil {
		return rollback(normalizedErr)
	} else if affected, countErr := result.RowsAffected(); countErr == nil {
		rows += affected
	}
	if result, normalizedErr := tx.ExecContext(ctx, `
		UPDATE authorization_license_assignments
		SET status = 'revoked', revoked_at = NOW()
		WHERE business_id = ? AND user_id = ? AND status = 'active'`, businessID, userID); normalizedErr != nil {
		return rollback(normalizedErr)
	} else if affected, countErr := result.RowsAffected(); countErr == nil {
		rows += affected
	}
	if result, normalizedErr := tx.ExecContext(ctx, `
		UPDATE authorization_role_assignments
		SET status = 'revoked', revoked_at = NOW()
		WHERE business_id = ? AND user_id = ? AND status = 'active'`, businessID, userID); normalizedErr != nil {
		return rollback(normalizedErr)
	} else if affected, countErr := result.RowsAffected(); countErr == nil {
		rows += affected
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO authorization_audit_events
			(business_id, target_user_id, action, reason, source_client)
		VALUES (?, ?, 'member.authorization.revoked.termination', 'HR employee terminated', 'app')`, businessID, userID); err != nil {
		return rollback(err)
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return rows, nil
}
