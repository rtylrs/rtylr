package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
	sqlnull "github.com/voxire/onion/internal/platform/utils/queries"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

// ─── Permissions ─────────────────────────────────────────────────────────────

func (r *Repository) GetPermission(ctx context.Context, ownerID, userID int64) (*domain.UserAppPermission, error) {
	row, err := r.queries.GetPermission(ctx, sqlc.GetPermissionParams{
		OwnerID: ownerID,
		UserID:  userID,
	})
	if err != nil {
		return nil, err
	}

	return permissionFromRow(row.ID, row.OwnerID, row.UserID, row.Apps, row.Role, row.IsActive, row.CreatedAt, row.UpdatedAt), nil
}

func (r *Repository) GetPermissionByBusiness(ctx context.Context, businessID, userID int64) (*domain.UserAppPermission, error) {
	row, err := r.queries.GetPermissionByBusiness(ctx, sqlc.GetPermissionByBusinessParams{
		BusinessID: sql.NullInt64{Int64: businessID, Valid: true},
		UserID:     userID,
	})
	if err != nil {
		return nil, err
	}

	return permissionFromRow(row.ID, row.OwnerID, row.UserID, row.Apps, row.Role, row.IsActive, row.CreatedAt, row.UpdatedAt), nil
}

func (r *Repository) GetPermissionByUserID(ctx context.Context, userID int64) (*domain.UserAppPermission, error) {
	row, err := r.queries.GetPermissionByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return permissionFromRow(row.ID, row.OwnerID, row.UserID, row.Apps, row.Role, row.IsActive, row.CreatedAt, row.UpdatedAt), nil
}

func (r *Repository) ListPermissions(ctx context.Context, ownerID int64) ([]domain.UserAppPermission, error) {
	rows, err := r.queries.ListPermissions(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list user app permissions", logger.Error(err))

		return nil, err
	}

	perms := make([]domain.UserAppPermission, 0, len(rows))

	for _, row := range rows {
		p := permissionFromRow(row.ID, row.OwnerID, row.UserID, row.Apps, row.Role, row.IsActive, row.CreatedAt, row.UpdatedAt)
		perms = append(perms, *p)
	}

	return perms, nil
}

func (r *Repository) UpsertPermission(ctx context.Context, businessID, ownerID, userID int64, apps []string, role domain.Role) error {
	appsJSON, err := json.Marshal(apps)
	if err != nil {
		return err
	}

	err = r.queries.UpsertPermission(ctx, sqlc.UpsertPermissionParams{
		BusinessID: sql.NullInt64{Int64: businessID, Valid: businessID > 0},
		OwnerID:    ownerID,
		UserID:     userID,
		Apps:       appsJSON,
		Role:       sqlc.NullUserAppPermissionsRole{UserAppPermissionsRole: sqlc.UserAppPermissionsRole(role), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert user app permission", logger.Error(err))

		return err
	}

	return nil
}

// GetAuthorizationRoleGrant returns the canonical app and permissions for a
// role that may be assigned in this business. Keeping this lookup in the
// repository prevents the application layer from making authorization
// decisions from client-supplied role metadata.
func (r *Repository) GetAuthorizationRoleGrant(ctx context.Context, businessID, roleID int64) (domain.AuthorizationRoleGrant, error) {
	var grant domain.AuthorizationRoleGrant
	if err := r.pool.QueryRowContext(ctx, `
		SELECT app_id
		FROM authorization_roles
		WHERE id = ? AND business_id IN (0, ?)
		  AND lifecycle = 'active' AND is_archived = FALSE
		LIMIT 1`, roleID, businessID).Scan(&grant.AppID); err != nil {
		return domain.AuthorizationRoleGrant{}, err
	}

	rows, err := r.pool.QueryContext(ctx, `
		SELECT permission_key
		FROM authorization_role_permissions
		WHERE role_id = ?
		ORDER BY permission_key`, roleID)
	if err != nil {
		return domain.AuthorizationRoleGrant{}, err
	}
	defer rows.Close()
	grant.Permissions = make([]string, 0)
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return domain.AuthorizationRoleGrant{}, err
		}
		grant.Permissions = append(grant.Permissions, permission)
	}
	if err := rows.Err(); err != nil {
		return domain.AuthorizationRoleGrant{}, err
	}
	return grant, nil
}

// UpsertNormalizedAuthorization dual-writes a member's normalized role,
// licenses, scopes, and audit event. Legacy rows remain updated by
// UpsertPermission during the migration window.
func (r *Repository) UpsertNormalizedAuthorization(ctx context.Context, businessID, actorID, userID int64, apps []string, legacyRole domain.Role, dashRole string, assignments []domain.AppRoleAssignmentInput, groupIDs []int64) error {
	if businessID <= 0 || userID <= 0 {
		return nil
	}
	if dashRole == "" {
		dashRole = legacyDashRole(legacyRole)
	}
	dashRole = strings.TrimPrefix(dashRole, "dash.")
	if !isDashRole(dashRole) {
		return errors.New("invalid dash role")
	}

	appIDs := uniqueAuthorizationApps(apps)
	for _, assignment := range assignments {
		if assignment.AppID != "" {
			appIDs = appendUniqueAuthorizationApp(appIDs, assignment.AppID)
		}
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(cause error) error {
		_ = tx.Rollback()
		if authorizationSchemaUnavailable(cause) {
			return nil
		}
		return cause
	}

	var dashRoleID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM authorization_roles
		WHERE business_id = 0 AND app_id = 'dash' AND role_key = ?
		  AND lifecycle = 'active' AND is_archived = FALSE
		LIMIT 1`, "dash."+dashRole).Scan(&dashRoleID)
	if err != nil {
		return rollback(err)
	}
	if groupIDs != nil {
		if err = r.replaceUserGroupMemberships(ctx, tx, businessID, actorID, userID, groupIDs); err != nil {
			return rollback(err)
		}
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO authorization_business_members
			(business_id, user_id, dash_role_id, dash_role_key, status, authorization_version)
		VALUES (?, ?, ?, ?, 'active', 1)
		ON DUPLICATE KEY UPDATE
			dash_role_id = VALUES(dash_role_id),
			dash_role_key = VALUES(dash_role_key),
			status = 'active',
			authorization_version = authorization_version + 1,
			updated_at = NOW()`,
		businessID, userID, dashRoleID, dashRole,
	)
	if err != nil {
		return rollback(err)
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE authorization_license_assignments
		SET status = 'revoked', revoked_by = ?, revoked_at = NOW()
		WHERE business_id = ? AND user_id = ? AND group_id IS NULL AND status = 'active'`, actorID, businessID, userID); err != nil {
		return rollback(err)
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE authorization_role_assignments
		SET status = 'revoked', revoked_by = ?, revoked_at = NOW()
		WHERE business_id = ? AND user_id = ? AND group_id IS NULL AND status = 'active'`, actorID, businessID, userID); err != nil {
		return rollback(err)
	}

	for _, appID := range appIDs {
		if appID == "" || appID == "dash" {
			continue
		}
		if !isTenantAppID(appID) {
			return rollback(fmt.Errorf("invalid app id: %s", appID))
		}
		if err = r.insertLicenseAssignments(ctx, tx, businessID, actorID, userID, appID, assignments); err != nil {
			return rollback(err)
		}

		appAssignments := assignmentsForApp(assignments, appID)
		if len(appAssignments) == 0 {
			roleKey := appRoleKey(appID, legacyRole)
			roleID, roleErr := r.lookupAuthorizationRoleID(ctx, tx, businessID, appID, roleKey)
			if roleErr != nil {
				return rollback(roleErr)
			}
			appAssignments = []domain.AppRoleAssignmentInput{{
				AppID:         appID,
				RoleIDs:       []int64{roleID},
				PrimaryRoleID: &roleID,
				Source:        "explicit",
				Status:        "active",
			}}
		}

		for index, assignment := range appAssignments {
			if assignment.Status != "" && assignment.Status != "active" {
				continue
			}
			roleIDs := assignment.RoleIDs
			if len(roleIDs) == 0 {
				roleKey := appRoleKey(appID, legacyRole)
				roleID, roleErr := r.lookupAuthorizationRoleID(ctx, tx, businessID, appID, roleKey)
				if roleErr != nil {
					return rollback(roleErr)
				}
				roleIDs = []int64{roleID}
			}
			for roleIndex, roleID := range roleIDs {
				if err = r.validateAuthorizationRoleID(ctx, tx, businessID, appID, roleID); err != nil {
					return rollback(err)
				}
				isPrimary := roleIndex == 0 && index == 0
				if assignment.PrimaryRoleID != nil {
					isPrimary = *assignment.PrimaryRoleID == roleID
				}
				var assignmentID int64
				// Assignment source is server-owned metadata. Browser input may
				// request an explicit or group assignment, but cannot manufacture a
				// system or Dash-default grant.
				source := authorization.SourceExplicit
				assignmentUserID := sql.NullInt64{Int64: userID, Valid: true}
				groupID := sql.NullInt64{}
				if assignment.GroupID != nil || source == authorization.SourceGroup {
					if assignment.GroupID == nil {
						return rollback(errors.New("group role assignment requires a group id"))
					}
					var validGroupID int64
					if err := tx.QueryRowContext(ctx, `
						SELECT id FROM user_groups
						WHERE id = ? AND (business_id = ? OR business_id = 0)
						LIMIT 1`, *assignment.GroupID, businessID).Scan(&validGroupID); err != nil {
						return rollback(fmt.Errorf("group role assignment is not valid: %w", err))
					}
					source = authorization.SourceGroup
					assignmentUserID = sql.NullInt64{}
					groupID = sql.NullInt64{Int64: *assignment.GroupID, Valid: true}
				}
				result, insertErr := tx.ExecContext(ctx, `
					INSERT INTO authorization_role_assignments
						(business_id, user_id, group_id, app_id, role_id, source, status, is_primary, granted_by)
					VALUES (?, ?, ?, ?, ?, ?, 'active', ?, ?)`,
					businessID, assignmentUserID, groupID, appID, roleID, source, isPrimary, actorID,
				)
				if insertErr != nil {
					return rollback(insertErr)
				}
				assignmentID, err = result.LastInsertId()
				if err != nil {
					return rollback(err)
				}
				for _, scope := range assignment.Scopes {
					if _, err = tx.ExecContext(ctx, `
						INSERT INTO authorization_role_assignment_scopes
							(assignment_id, scope_type, scope_id)
						VALUES (?, ?, ?)`, assignmentID, scope.Type, scope.ID); err != nil {
						return rollback(err)
					}
				}
			}
		}
	}

	auditAfter, err := json.Marshal(struct {
		DashRole  string   `json:"dash_role"`
		Apps      []string `json:"apps"`
		RoleCount int      `json:"role_count"`
	}{DashRole: dashRole, Apps: appIDs, RoleCount: len(assignments)})
	if err != nil {
		return rollback(err)
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO authorization_audit_events
			(business_id, actor_user_id, target_user_id, action, before_json, after_json, source_client)
		VALUES (?, ?, ?, 'member.authorization.updated', NULL, ?, 'dash')`,
		businessID, actorID, userID, auditAfter); err != nil {
		return rollback(err)
	}

	return tx.Commit()
}

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

// ReconcileTerminatedEmployeeAccess sweeps every business and deactivates any
// still-active permission row whose user is a terminated employee of that
// same business. Idempotent; the daily job runs it so a missed synchronous
// revoke is corrected within a day.
func (r *Repository) ReconcileTerminatedEmployeeAccess(ctx context.Context) (int64, error) {
	rows, err := r.queries.ReconcileTerminatedEmployeeAccess(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("terminated-employee access reconciliation failed", logger.Error(err))
		return 0, err
	}
	if _, err := r.pool.ExecContext(ctx, `
		UPDATE authorization_business_members m
		JOIN hr_employees e ON e.user_id = m.user_id AND e.business_id = m.business_id
		SET m.status = 'revoked', m.authorization_version = m.authorization_version + 1, m.updated_at = NOW()
		WHERE e.status = 'terminated' AND m.status <> 'revoked'`); err != nil && !authorizationSchemaUnavailable(err) {
		return 0, err
	}
	return rows, nil
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
			name: "legacy permission",
			query: `DELETE FROM user_app_permissions
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
			if authorizationSchemaUnavailable(err) {
				continue
			}
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

// ─── Staff listing (join users + permissions) ─────────────────────────────────

func (r *Repository) ListStaffUsers(ctx context.Context, ownerID int64, limit, offset int32) ([]domain.StaffUser, error) {
	rows, err := r.queries.ListStaffUsers(ctx, sqlc.ListStaffUsersParams{
		OwnerID: ownerID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list staff users", logger.Error(err))

		return nil, err
	}

	users := make([]domain.StaffUser, 0, len(rows))

	for _, row := range rows {
		var su domain.StaffUser
		su.ID = row.ID
		su.Email = row.Email
		su.FullName = row.FullName
		if row.CreatedAt.Valid {
			su.CreatedAt = row.CreatedAt.Time
		}
		su.Role = domain.Role(row.Role.UserAppPermissionsRole)
		if row.IsActive.Valid {
			su.IsActive = row.IsActive.Bool
		}

		if err := json.Unmarshal(row.Apps, &su.Apps); err != nil {
			su.Apps = []string{}
		}

		users = append(users, su)
	}

	userIDs := make([]int64, len(users))
	for i := range users {
		userIDs[i] = users[i].ID
		users[i].Groups = []domain.GroupSummary{}
	}

	groupsByUser, err := r.groupsForUsers(ctx, ownerID, userIDs)
	if err == nil {
		for i := range users {
			if groups := groupsByUser[users[i].ID]; groups != nil {
				users[i].Groups = groups
			}
		}
	}

	return users, nil
}

// ListStaffUsersByBusinessID lists team members scoped to a single business
// (via user_app_permissions.business_id), so each business sees only its own
// team. Group enrichment stays owner-scoped (groups are owner-level), resolved
// from the business's owner.
func (r *Repository) ListStaffUsersByBusinessID(ctx context.Context, businessID int64, limit, offset int32) ([]domain.StaffUser, error) {
	rows, err := r.queries.ListStaffUsersByBusinessID(ctx, sqlc.ListStaffUsersByBusinessIDParams{
		BusinessID: sqlnull.Int64(businessID),
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list staff users by business", logger.Error(err))

		return nil, err
	}

	users := make([]domain.StaffUser, 0, len(rows))

	for _, row := range rows {
		var su domain.StaffUser
		su.ID = row.ID
		su.Email = row.Email
		su.FullName = row.FullName
		if row.CreatedAt.Valid {
			su.CreatedAt = row.CreatedAt.Time
		}
		su.Role = domain.Role(row.Role.UserAppPermissionsRole)
		if row.IsActive.Valid {
			su.IsActive = row.IsActive.Bool
		}

		if err := json.Unmarshal(row.Apps, &su.Apps); err != nil {
			su.Apps = []string{}
		}

		users = append(users, su)
	}

	userIDs := make([]int64, len(users))
	for i := range users {
		userIDs[i] = users[i].ID
		users[i].Groups = []domain.GroupSummary{}
	}

	if _, ownerID, ownerErr := r.GetBusinessByID(ctx, businessID); ownerErr == nil {
		if groupsByUser, gErr := r.groupsForUsers(ctx, ownerID, userIDs); gErr == nil {
			for i := range users {
				if groups := groupsByUser[users[i].ID]; groups != nil {
					users[i].Groups = groups
				}
			}
		}
	}

	return users, nil
}

// groupsForUsers batches group lookups for many users into a single query,
// mapping the result back per user ID in memory (avoids an N+1 over users).
func (r *Repository) groupsForUsers(ctx context.Context, ownerID int64, userIDs []int64) (map[int64][]domain.GroupSummary, error) {
	result := make(map[int64][]domain.GroupSummary, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	rows, err := r.queries.ListUserGroupsForUsers(ctx, sqlc.ListUserGroupsForUsersParams{
		UserIds: userIDs,
		OwnerID: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to batch list user groups", logger.Error(err))
		return nil, err
	}

	for _, row := range rows {
		summary := domain.GroupSummary{ID: row.ID, Name: row.Name, Color: row.Color.String}
		result[row.UserID] = append(result[row.UserID], summary)
	}

	return result, nil
}

func (r *Repository) GetStaffUserByID(ctx context.Context, ownerID, businessID, userID int64) (*domain.StaffUser, error) {
	row, err := r.queries.GetStaffUserByID(ctx, sqlc.GetStaffUserByIDParams{
		OwnerID:    ownerID,
		BusinessID: sql.NullInt64{Int64: businessID, Valid: true},
		UserID:     userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		logger.Ctx(ctx).Error("Failed to get staff user by id", logger.Error(err))

		return nil, err
	}

	su := &domain.StaffUser{
		ID:       row.ID,
		Email:    row.Email,
		FullName: row.FullName,
		Role:     domain.Role(row.Role.UserAppPermissionsRole),
	}

	if row.CreatedAt.Valid {
		su.CreatedAt = row.CreatedAt.Time
	}

	if row.IsActive.Valid {
		su.IsActive = row.IsActive.Bool
	}

	if err := json.Unmarshal(row.Apps, &su.Apps); err != nil {
		su.Apps = []string{}
	}

	groups, err := r.GetUserGroups(ctx, ownerID, su.ID)
	if err != nil || groups == nil {
		groups = []domain.GroupSummary{}
	}
	su.Groups = groups

	return su, nil
}

// ─── Invite Tokens ──────────────────────────────────────────────────────────

func (r *Repository) CreateInviteToken(ctx context.Context, userID, ownerID int64, token string, expiresAt time.Time) error {
	err := r.queries.CreateInviteToken(ctx, sqlc.CreateInviteTokenParams{
		UserID:    userID,
		OwnerID:   ownerID,
		Token:     token,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create invite token", logger.Error(err))
	}
	return err
}

func (r *Repository) GetInviteToken(ctx context.Context, token string) (int64, int64, time.Time, error) {
	row, err := r.queries.GetInviteToken(ctx, token)
	if err != nil {
		return 0, 0, time.Time{}, err
	}

	if row.UsedAt.Valid {
		return 0, 0, time.Time{}, domain.ErrTokenAlreadyUsed
	}

	if time.Now().After(row.ExpiresAt) {
		return 0, 0, time.Time{}, domain.ErrTokenExpired
	}

	return row.UserID, row.OwnerID, row.ExpiresAt, nil
}

func (r *Repository) MarkInviteTokenUsed(ctx context.Context, token string) error {
	// Atomic single-use claim: only an unused, unexpired token is marked, so
	// concurrent accepts cannot both succeed.
	res, err := r.queries.ConsumeInviteToken(ctx, token)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrTokenAlreadyUsed
	}
	return nil
}

// ReleaseInviteToken reverts the single-use claim taken by MarkInviteTokenUsed.
// AcceptInvite calls it when a later step fails, so the invitee can retry the
// accept instead of being left with a consumed token and no activation.
func (r *Repository) ReleaseInviteToken(ctx context.Context, token string) error {
	return r.queries.ReleaseInviteToken(ctx, token)
}

func (r *Repository) SetPassword(ctx context.Context, userID int64, passwordHash string) error {
	return r.queries.SetPassword(ctx, sqlc.SetPasswordParams{
		PasswordHash: passwordHash,
		ID:           userID,
	})
}

// GetBusinessByID returns the business name and owner ID.
func (r *Repository) GetBusinessByID(ctx context.Context, businessID int64) (string, int64, error) {
	row, err := r.queries.GetBusinessByID(ctx, businessID)
	if err != nil {
		return "", 0, err
	}
	return row.Name, row.OwnerID, nil
}

func (r *Repository) GetBusinessIDByOwnerID(ctx context.Context, ownerID int64) (int64, error) {
	return r.queries.GetBusinessIDByOwnerID(ctx, ownerID)
}

// ─── Access Requests ────────────────────────────────────────────────────────

func (r *Repository) CreateAccessRequest(ctx context.Context, businessID, userID int64, app string) error {
	return r.queries.CreateAccessRequest(ctx, sqlc.CreateAccessRequestParams{
		BusinessID: businessID,
		UserID:     userID,
		App:        app,
	})
}

// ─── Onboarding ─────────────────────────────────────────────────────────────
//
// The `users` table (schemas/shared/users.sql) is a GLOBAL identity table keyed
// by `id` — it has no `owner_id` column (tenant association lives in the nullable
// `business_id`). Onboarding state belongs to the user identity, not to a tenant,
// so these queries intentionally key on `id` only and do NOT filter by owner_id.
// If `users` is ever re-scoped to a tenant during a schema refactor, revisit these
// two queries (and add an owner_id filter) before that migration ships.

func (r *Repository) IsOnboarded(ctx context.Context, userID int64) bool {
	onboarded, err := r.queries.IsOnboarded(ctx, userID)
	if err != nil {
		// A missing row (sql.ErrNoRows) legitimately means "not onboarded"; any
		// other error indicates a DB problem and must not be swallowed silently.
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Warn("Failed to read onboarding state, defaulting to not onboarded",
				logger.Any("user_id", userID), logger.Error(err))
		}
		return false
	}
	return onboarded
}

func (r *Repository) MarkOnboarded(ctx context.Context, userID int64) error {
	return r.queries.MarkOnboarded(ctx, userID)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func permissionFromRow(id, ownerID, userID int64, apps json.RawMessage, role sqlc.NullUserAppPermissionsRole, isActive sql.NullBool, createdAt, updatedAt sql.NullTime) *domain.UserAppPermission {
	p := &domain.UserAppPermission{
		ID:      id,
		OwnerID: ownerID,
		UserID:  userID,
	}

	if role.Valid {
		p.Role = domain.Role(role.UserAppPermissionsRole)
	}

	if isActive.Valid {
		p.IsActive = isActive.Bool
	}

	if createdAt.Valid {
		p.CreatedAt = createdAt.Time
	}

	if updatedAt.Valid {
		p.UpdatedAt = updatedAt.Time
	}

	if err := json.Unmarshal(apps, &p.Apps); err != nil {
		p.Apps = []string{}
	}

	return p
}

// CountActiveMembers returns the billable active seat count for an owner.
func (r *Repository) CountActiveMembers(ctx context.Context, ownerID int64) (int, error) {
	count, err := r.queries.CountActiveMembers(ctx, sqlc.CountActiveMembersParams{
		OwnerID:   ownerID,
		OwnerID_2: ownerID,
		OwnerID_3: ownerID,
	})
	if err != nil {
		return 0, err
	}
	if count < 1 {
		return 1, nil
	}
	return int(count), nil
}

// ─── Groups ──────────────────────────────────────────────────────────────────

// CreateGroup writes the group row and its resource permissions in one
// transaction. Without it a failure mid-rewrite leaves the group created with
// no permissions and blocks a retry under the same name (uk_group_owner_name).
func (r *Repository) CreateGroup(ctx context.Context, ownerID int64, input domain.CreateGroupInput) (int64, error) {
	if err := validateResourcePermissions(input.ResourcePermissions); err != nil {
		return 0, err
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	qtx := r.queries.WithTx(tx)

	res, err := qtx.CreateGroup(ctx, sqlc.CreateGroupParams{
		OwnerID:     ownerID,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Color:       sql.NullString{String: input.Color, Valid: input.Color != ""},
	})
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if err := r.setGroupResourcePermissions(ctx, qtx, id, input.ResourcePermissions); err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

// validateResourcePermissions rejects levels outside the
// user_group_resource_permissions.permission_level ENUM('viewer','editor')
// before any row is written.
func validateResourcePermissions(perms []domain.ResourcePermission) error {
	for _, p := range perms {
		switch p.PermissionLevel {
		case domain.PermLevelViewer, domain.PermLevelEditor:
		default:
			return domain.ErrInvalidPermissionLevel
		}
	}
	return nil
}

// setGroupResourcePermissions replaces a group's resource-permission rows. It
// must run inside the caller's transaction: the delete-then-upsert rewrite
// would otherwise drop the stored permissions when an upsert fails.
func (r *Repository) setGroupResourcePermissions(ctx context.Context, qtx *sqlc.Queries, groupID int64, perms []domain.ResourcePermission) error {
	if err := qtx.DeleteGroupResourcePermissions(ctx, groupID); err != nil {
		return err
	}
	for _, p := range perms {
		if err := qtx.UpsertGroupResourcePermission(ctx, sqlc.UpsertGroupResourcePermissionParams{
			GroupID:         groupID,
			ResourceKey:     p.ResourceKey,
			PermissionLevel: sqlc.UserGroupResourcePermissionsPermissionLevel(p.PermissionLevel),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) GetGroup(ctx context.Context, ownerID, groupID int64) (*domain.Group, error) {
	row, err := r.queries.GetGroup(ctx, sqlc.GetGroupParams{ID: groupID, OwnerID: ownerID})
	if err != nil {
		return nil, err
	}
	perms, err := r.queries.ListGroupResourcePermissions(ctx, groupID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list group resource permissions", logger.Error(err))
		return nil, err
	}
	g := &domain.Group{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		Name:        row.Name,
		Description: row.Description,
		Color:       row.Color.String,
	}
	if row.CreatedAt.Valid {
		g.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		g.UpdatedAt = row.UpdatedAt.Time
	}
	g.ResourcePermissions = make([]domain.ResourcePermission, 0, len(perms))
	for _, p := range perms {
		g.ResourcePermissions = append(g.ResourcePermissions, domain.ResourcePermission{
			ResourceKey:     p.ResourceKey,
			PermissionLevel: domain.PermissionLevel(p.PermissionLevel),
		})
	}
	return g, nil
}

func (r *Repository) ListGroups(ctx context.Context, ownerID int64) ([]domain.Group, error) {
	rows, err := r.queries.ListGroups(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	groupIDs := make([]int64, len(rows))
	for i, row := range rows {
		groupIDs[i] = row.ID
	}
	permsByGroup := r.resourcePermissionsForGroups(ctx, groupIDs)

	groups := make([]domain.Group, 0, len(rows))
	for _, row := range rows {
		g := domain.Group{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
			Name:        row.Name,
			Description: row.Description,
			Color:       row.Color.String,
			MemberCount: int(row.MemberCount),
		}
		if row.CreatedAt.Valid {
			g.CreatedAt = row.CreatedAt.Time
		}
		if row.UpdatedAt.Valid {
			g.UpdatedAt = row.UpdatedAt.Time
		}
		g.ResourcePermissions = permsByGroup[row.ID]
		groups = append(groups, g)
	}
	return groups, nil
}

// resourcePermissionsForGroups batches resource-permission lookups for many
// groups into one query, mapping the result per group ID in memory (avoids an
// N+1 over the group list). On error it returns an empty map so callers degrade
// to groups without resource permissions, matching the previous best-effort
// per-group behaviour.
func (r *Repository) resourcePermissionsForGroups(ctx context.Context, groupIDs []int64) map[int64][]domain.ResourcePermission {
	result := make(map[int64][]domain.ResourcePermission, len(groupIDs))
	if len(groupIDs) == 0 {
		return result
	}

	rows, err := r.queries.ListGroupResourcePermissionsForGroups(ctx, groupIDs)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to batch list group resource permissions", logger.Error(err))
		return result
	}

	for _, row := range rows {
		p := domain.ResourcePermission{
			ResourceKey:     row.ResourceKey,
			PermissionLevel: domain.PermissionLevel(row.PermissionLevel),
		}
		result[row.GroupID] = append(result[row.GroupID], p)
	}
	return result
}

// UpdateGroup writes the group row and its rewritten resource permissions in
// one transaction so a failed permission write cannot leave the row updated
// with the previous permissions silently deleted.
func (r *Repository) UpdateGroup(ctx context.Context, ownerID, groupID int64, input domain.UpdateGroupInput) error {
	if err := validateResourcePermissions(input.ResourcePermissions); err != nil {
		return err
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	qtx := r.queries.WithTx(tx)

	res, err := qtx.UpdateGroup(ctx, sqlc.UpdateGroupParams{
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Color:       sql.NullString{String: input.Color, Valid: input.Color != ""},
		ID:          groupID,
		OwnerID:     ownerID,
	})
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	// A foreign or stale group id must never fall through to the unscoped
	// resource-permission rewrite below.
	if affected == 0 {
		_ = tx.Rollback()
		return domain.ErrGroupNotFound
	}
	if err := r.setGroupResourcePermissions(ctx, qtx, groupID, input.ResourcePermissions); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *Repository) DeleteGroup(ctx context.Context, ownerID, groupID int64) error {
	// Scoped delete first: only when the caller actually owns the group may
	// the (unscoped) resource-permission rows be removed.
	res, err := r.queries.DeleteGroup(ctx, sqlc.DeleteGroupParams{ID: groupID, OwnerID: ownerID})
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrGroupNotFound
	}
	return r.queries.DeleteGroupResourcePermissions(ctx, groupID)
}

func (r *Repository) AddGroupMembers(ctx context.Context, ownerID, groupID, businessID, addedBy int64, userIDs []int64) error {
	for _, uid := range userIDs {
		if err := r.queries.AddGroupMember(ctx, sqlc.AddGroupMemberParams{
			GroupID:    groupID,
			UserID:     uid,
			OwnerID:    ownerID,
			BusinessID: businessID,
			AddedBy:    sql.NullInt64{Int64: addedBy, Valid: addedBy > 0},
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) RemoveGroupMember(ctx context.Context, groupID, userID int64) error {
	return r.queries.RemoveGroupMember(ctx, sqlc.RemoveGroupMemberParams{
		GroupID: groupID,
		UserID:  userID,
	})
}

func (r *Repository) GetGroupDetail(ctx context.Context, ownerID, groupID int64) (*domain.GroupDetail, error) {
	g, err := r.GetGroup(ctx, ownerID, groupID)
	if err != nil {
		return nil, err
	}
	memberRows, err := r.queries.ListGroupMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	detail := &domain.GroupDetail{Group: *g}
	detail.Members = make([]domain.GroupMember, 0, len(memberRows))
	for _, m := range memberRows {
		gm := domain.GroupMember{
			UserID:   m.UserID,
			Email:    m.Email,
			FullName: m.FullName,
		}
		if m.AddedAt.Valid {
			gm.AddedAt = m.AddedAt.Time
		}
		detail.Members = append(detail.Members, gm)
	}
	return detail, nil
}

func (r *Repository) GetUserGroups(ctx context.Context, ownerID, userID int64) ([]domain.GroupSummary, error) {
	rows, err := r.queries.ListUserGroups(ctx, sqlc.ListUserGroupsParams{UserID: userID, OwnerID: ownerID})
	if err != nil {
		return nil, err
	}
	summaries := make([]domain.GroupSummary, 0, len(rows))
	for _, row := range rows {
		summaries = append(summaries, domain.GroupSummary{ID: row.ID, Name: row.Name, Color: row.Color.String})
	}
	return summaries, nil
}

// GetUserEffectiveResources returns the union of all group resource permissions for a user.
// Returns map[resourceKey]highest_permission_level — viewer=1, editor=2.
func (r *Repository) GetUserEffectiveResources(ctx context.Context, ownerID, userID int64) (map[string]domain.PermissionLevel, error) {
	effective := map[string]domain.PermissionLevel{}
	levelRank := map[domain.PermissionLevel]int{
		domain.PermLevelViewer: 1,
		domain.PermLevelEditor: 2,
	}
	merge := func(resourceKey string, level domain.PermissionLevel) {
		if existing, ok := effective[resourceKey]; !ok || levelRank[level] > levelRank[existing] {
			effective[resourceKey] = level
		}
	}
	groupPerms, err := r.queries.ListUserGroupResourcePermissions(ctx, sqlc.ListUserGroupResourcePermissionsParams{
		UserID:  userID,
		OwnerID: ownerID,
	})
	if err == nil {
		for _, gp := range groupPerms {
			merge(gp.ResourceKey, domain.PermissionLevel(gp.PermissionLevel))
		}
	}
	return effective, nil
}

// GetOwnerApps returns the list of apps the owner is subscribed to.
func (r *Repository) GetOwnerApps(ctx context.Context, ownerID int64) ([]string, error) {
	// cross-schema read from subscriptions + plans — approved for owner app resolution
	appsJSON, err := r.queries.GetOwnerApps(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	var apps []string
	if jsonErr := json.Unmarshal(appsJSON, &apps); jsonErr != nil {
		return []string{}, nil
	}

	return apps, nil
}

// GetAuthorizationSnapshot loads only normalized authorization rows for a
// member. The bool is false during the dual-read migration window when no
// normalized member row exists, allowing the caller to use the legacy profile.
// SQL stays in this repository so the application layer receives a persistence
// independent snapshot for the pure resolver.
func (r *Repository) GetAuthorizationSnapshot(ctx context.Context, businessID, userID int64, subscribedApps []string) (authorization.Snapshot, bool, error) {
	if businessID <= 0 || userID <= 0 {
		return authorization.Snapshot{}, false, nil
	}

	var (
		dashRoleID           sql.NullInt64
		dashRoleKey          string
		status               string
		authorizationVersion int64
	)
	err := r.pool.QueryRowContext(ctx, `
		SELECT dash_role_id, dash_role_key, status, authorization_version
		FROM authorization_business_members
		WHERE business_id = ? AND user_id = ?
		LIMIT 1`, businessID, userID).Scan(&dashRoleID, &dashRoleKey, &status, &authorizationVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return authorization.Snapshot{}, false, nil
	}
	if err != nil {
		if authorizationSchemaUnavailable(err) {
			return authorization.Snapshot{}, false, nil
		}
		return authorization.Snapshot{}, false, err
	}

	roles, err := r.loadAuthorizationRoles(ctx, businessID)
	if err != nil {
		return authorization.Snapshot{}, false, err
	}
	if err := r.loadAuthorizationRolePermissions(ctx, businessID, roles); err != nil {
		return authorization.Snapshot{}, false, err
	}

	roleID := nullInt64Pointer(dashRoleID)
	if dashRoleKey == "" {
		dashRoleKey = "member"
	}
	if !strings.HasPrefix(dashRoleKey, "dash.") {
		dashRoleKey = "dash." + dashRoleKey
	}
	if !dashRoleID.Valid {
		for _, role := range roles {
			if role.AppID == nil && role.Key == dashRoleKey {
				roleID = &role.ID
				break
			}
		}
	}

	dashPermissions := []string{}
	if roleID != nil {
		dashPermissions = authorization.PermissionsForRole(*roleID, roles)
	}
	allPermissions, err := r.loadAuthorizationCatalog(ctx)
	if err != nil {
		return authorization.Snapshot{}, false, err
	}

	licenses, err := r.loadAuthorizationLicenses(ctx, businessID, userID)
	if err != nil {
		return authorization.Snapshot{}, false, err
	}
	assignments, err := r.loadAuthorizationRoleAssignments(ctx, businessID, userID)
	if err != nil {
		return authorization.Snapshot{}, false, err
	}

	apps := make([]authorization.AppID, 0, len(subscribedApps))
	for _, app := range subscribedApps {
		if app == "" || app == "dash" {
			continue
		}
		apps = append(apps, authorization.AppID(app))
	}

	return authorization.Snapshot{
		BusinessID:           businessID,
		DashRole:             strings.TrimPrefix(dashRoleKey, "dash."),
		DashRoleID:           roleID,
		MembershipActive:     status == "active",
		SubscribedApps:       apps,
		DashPermissions:      dashPermissions,
		Licenses:             licenses,
		Assignments:          assignments,
		Roles:                roles,
		AllPermissions:       allPermissions,
		AuthorizationVersion: authorizationVersion,
	}, true, nil
}

func (r *Repository) loadAuthorizationRoles(ctx context.Context, businessID int64) ([]authorization.Role, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id, app_id, role_key, display_name, description, level,
		       parent_role_id, is_archived, is_system, assignment_mode
		FROM authorization_roles
		WHERE business_id IN (0, ?)
		  AND lifecycle = 'active'
		ORDER BY business_id, app_id, level, id`, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]authorization.Role, 0)
	for rows.Next() {
		var (
			role                 authorization.Role
			appID                string
			description          sql.NullString
			parentRoleID         sql.NullInt64
			isArchived, isSystem bool
		)
		if err := rows.Scan(
			&role.ID,
			&appID,
			&role.Key,
			&role.DisplayName,
			&description,
			&role.Level,
			&parentRoleID,
			&isArchived,
			&isSystem,
			&role.AssignmentMode,
		); err != nil {
			return nil, err
		}
		role.AppID = authorizationAppPointer(appID)
		role.ParentRoleID = nullInt64Pointer(parentRoleID)
		role.IsArchived = isArchived
		role.IsSystem = isSystem
		if description.Valid {
			role.Description = description.String
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *Repository) loadAuthorizationRolePermissions(ctx context.Context, businessID int64, roles []authorization.Role) error {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT rp.role_id, rp.permission_key
		FROM authorization_role_permissions rp
		JOIN authorization_roles r ON r.id = rp.role_id
		WHERE r.business_id IN (0, ?) AND r.lifecycle = 'active'`, businessID)
	if err != nil {
		return err
	}
	defer rows.Close()

	permissionsByRole := make(map[int64][]string, len(roles))
	for rows.Next() {
		var roleID int64
		var permission string
		if err := rows.Scan(&roleID, &permission); err != nil {
			return err
		}
		permissionsByRole[roleID] = append(permissionsByRole[roleID], permission)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for index := range roles {
		roles[index].Permissions = append([]string{}, permissionsByRole[roles[index].ID]...)
	}
	return nil
}

func (r *Repository) loadAuthorizationCatalog(ctx context.Context) ([]string, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT permission_key
		FROM authorization_permission_catalog
		ORDER BY permission_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := make([]string, 0)
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, rows.Err()
}

func (r *Repository) loadAuthorizationLicenses(ctx context.Context, businessID, userID int64) ([]authorization.License, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT app_id, source, status, group_id, granted_at, revoked_at
		FROM authorization_license_assignments
		WHERE business_id = ?
		  AND status = 'active'
		  AND (user_id = ? OR group_id IN (
			SELECT group_id
			FROM user_group_members
			WHERE business_id = ? AND user_id = ?
		  ))`, businessID, userID, businessID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	licenses := make([]authorization.License, 0)
	for rows.Next() {
		var (
			appID, source, status string
			groupID               sql.NullInt64
			grantedAt, revokedAt  sql.NullTime
		)
		if err := rows.Scan(&appID, &source, &status, &groupID, &grantedAt, &revokedAt); err != nil {
			return nil, err
		}
		license := authorization.License{
			AppID:     authorization.AppID(appID),
			Source:    authorization.AssignmentSource(source),
			Active:    status == "active",
			GroupID:   nullInt64Pointer(groupID),
			GrantedAt: time.Time{}.Format(time.RFC3339),
		}
		if grantedAt.Valid {
			license.GrantedAt = grantedAt.Time.UTC().Format(time.RFC3339)
		}
		if revokedAt.Valid {
			value := revokedAt.Time.UTC().Format(time.RFC3339)
			license.RevokedAt = &value
		}
		licenses = append(licenses, license)
	}
	return licenses, rows.Err()
}

func (r *Repository) loadAuthorizationRoleAssignments(ctx context.Context, businessID, userID int64) ([]authorization.RoleAssignment, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT a.id, a.app_id, a.group_id, a.role_id, a.source, a.is_primary,
		       s.scope_type, s.scope_id
		FROM authorization_role_assignments a
		LEFT JOIN authorization_role_assignment_scopes s ON s.assignment_id = a.id
		WHERE a.business_id = ?
		  AND a.status = 'active'
		  AND (a.user_id = ? OR a.group_id IN (
			SELECT group_id
			FROM user_group_members
			WHERE business_id = ? AND user_id = ?
		  ))
		ORDER BY a.id`, businessID, userID, businessID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type assignmentRow struct {
		assignment authorization.RoleAssignment
	}
	byID := make(map[int64]*assignmentRow)
	order := make([]int64, 0)
	for rows.Next() {
		var (
			assignmentID int64
			appID        string
			groupID      sql.NullInt64
			roleID       int64
			source       string
			isPrimary    bool
			scopeType    sql.NullString
			scopeID      sql.NullString
		)
		if err := rows.Scan(&assignmentID, &appID, &groupID, &roleID, &source, &isPrimary, &scopeType, &scopeID); err != nil {
			return nil, err
		}
		entry, ok := byID[assignmentID]
		if !ok {
			entry = &assignmentRow{assignment: authorization.RoleAssignment{
				AppID:   authorization.AppID(appID),
				GroupID: nullInt64Pointer(groupID),
				Source:  authorization.AssignmentSource(source),
				RoleIDs: []int64{},
				Scopes:  []authorization.Scope{},
				Active:  true,
			}}
			byID[assignmentID] = entry
			order = append(order, assignmentID)
		}
		entry.assignment.RoleIDs = appendUniqueInt64(entry.assignment.RoleIDs, roleID)
		if isPrimary {
			primaryRoleID := roleID
			entry.assignment.PrimaryRoleID = &primaryRoleID
		}
		if scopeType.Valid && scopeID.Valid {
			entry.assignment.Scopes = appendUniqueAuthorizationScope(entry.assignment.Scopes, authorization.Scope{
				Type: scopeType.String,
				ID:   scopeID.String,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	assignments := make([]authorization.RoleAssignment, 0, len(order))
	for _, assignmentID := range order {
		assignments = append(assignments, byID[assignmentID].assignment)
	}
	return assignments, nil
}

func authorizationSchemaUnavailable(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "authorization_") &&
		(strings.Contains(message, "doesn't exist") || strings.Contains(message, "no such table"))
}

func authorizationAppPointer(appID string) *authorization.AppID {
	if appID == "" || appID == "dash" {
		return nil
	}
	value := authorization.AppID(appID)
	return &value
}

func nullInt64Pointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}

func appendUniqueInt64(values []int64, value int64) []int64 {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func appendUniqueAuthorizationScope(values []authorization.Scope, value authorization.Scope) []authorization.Scope {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func (r *Repository) UpsertProvisioningState(ctx context.Context, userID, businessID int64, mode, state string, emailVerified, passwordChangeRequired bool) error {
	_, err := r.pool.ExecContext(ctx, `
		INSERT INTO user_provisioning_states
			(user_id, business_id, mode, state, email_verified, password_change_required)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			business_id = VALUES(business_id),
			mode = VALUES(mode),
			state = VALUES(state),
			email_verified = VALUES(email_verified),
			password_change_required = VALUES(password_change_required),
			updated_at = NOW()`,
		userID, businessID, mode, state, emailVerified, passwordChangeRequired,
	)
	if err != nil && authorizationSchemaUnavailable(err) {
		return nil
	}
	return err
}

func (r *Repository) GetProvisioningStatus(ctx context.Context, userID int64) (*authorization.ProvisioningStatus, bool, error) {
	var (
		mode, state                                 string
		emailVerified, passwordChangeRequired       bool
		mfaRequired, mfaEnrolled, recoveryAvailable bool
		lastMFAVerified                             sql.NullTime
	)
	err := r.pool.QueryRowContext(ctx, `
		SELECT mode, state, email_verified, password_change_required,
		       mfa_required, mfa_enrolled, recovery_available, last_mfa_verified_at
		FROM user_provisioning_states
		WHERE user_id = ?`, userID).Scan(
		&mode, &state, &emailVerified, &passwordChangeRequired,
		&mfaRequired, &mfaEnrolled, &recoveryAvailable, &lastMFAVerified,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		if authorizationSchemaUnavailable(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	mfaVerified := !mfaRequired || (mfaEnrolled && lastMFAVerified.Valid && time.Since(lastMFAVerified.Time) <= 5*time.Minute)
	return &authorization.ProvisioningStatus{
		Mode:                   authorization.ProvisioningMode(mode),
		State:                  authorization.ProvisioningState(state),
		EmailVerified:          emailVerified,
		PasswordChangeRequired: passwordChangeRequired,
		MFARequired:            mfaRequired,
		MFAEnrolled:            mfaEnrolled,
		MFAVerified:            mfaVerified,
		RecoveryAvailable:      recoveryAvailable,
	}, true, nil
}

func (r *Repository) IsMFAVerified(ctx context.Context, userID int64) (bool, error) {
	var required, enrolled bool
	var lastVerified sql.NullTime
	err := r.pool.QueryRowContext(ctx, `
		SELECT mfa_required, mfa_enrolled, last_mfa_verified_at
		FROM user_provisioning_states
		WHERE user_id = ?`, userID).Scan(&required, &enrolled, &lastVerified)
	if errors.Is(err, sql.ErrNoRows) || authorizationSchemaUnavailable(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !required {
		return true, nil
	}
	return enrolled && lastVerified.Valid && time.Since(lastVerified.Time) <= 5*time.Minute, nil
}

func (r *Repository) MarkPasswordChanged(ctx context.Context, userID int64) error {
	var businessID int64
	if err := r.pool.QueryRowContext(ctx, `
		SELECT business_id FROM user_provisioning_states WHERE user_id = ?`, userID).Scan(&businessID); err != nil {
		if errors.Is(err, sql.ErrNoRows) || authorizationSchemaUnavailable(err) {
			return nil
		}
		return err
	}
	_, err := r.pool.ExecContext(ctx, `
		UPDATE user_provisioning_states
		SET password_change_required = FALSE,
		    state = CASE WHEN email_verified THEN 'active' ELSE 'pending_activation' END,
		    password_changed_at = NOW(),
		    updated_at = NOW()
		WHERE user_id = ?`, userID)
	if err != nil && authorizationSchemaUnavailable(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return r.ActivateNormalizedAuthorization(ctx, businessID, userID)
}

func (r *Repository) SetNormalizedAuthorizationPending(ctx context.Context, businessID, userID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(cause error) error {
		_ = tx.Rollback()
		if authorizationSchemaUnavailable(cause) {
			return nil
		}
		return cause
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_business_members
		SET status = 'pending', authorization_version = authorization_version + 1, updated_at = NOW()
		WHERE business_id = ? AND user_id = ?`, businessID, userID); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_license_assignments
		SET status = 'pending'
		WHERE business_id = ? AND user_id = ? AND status = 'active'`, businessID, userID); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_role_assignments
		SET status = 'pending'
		WHERE business_id = ? AND user_id = ? AND status = 'active'`, businessID, userID); err != nil {
		return rollback(err)
	}
	return tx.Commit()
}

func (r *Repository) ActivateNormalizedAuthorization(ctx context.Context, businessID, userID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(cause error) error {
		_ = tx.Rollback()
		if authorizationSchemaUnavailable(cause) {
			return nil
		}
		return cause
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_business_members
		SET status = 'active', authorization_version = authorization_version + 1, updated_at = NOW()
		WHERE business_id = ? AND user_id = ? AND status = 'pending'`, businessID, userID); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_license_assignments
		SET status = 'active'
		WHERE business_id = ? AND user_id = ? AND status = 'pending'`, businessID, userID); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_role_assignments
		SET status = 'active'
		WHERE business_id = ? AND user_id = ? AND status = 'pending'`, businessID, userID); err != nil {
		return rollback(err)
	}
	return tx.Commit()
}

// GetOwnerSubscriptionInfo returns the plan slug and status for the owner's subscription.
func (r *Repository) GetOwnerSubscriptionInfo(ctx context.Context, ownerID int64) (string, string, error) {
	// cross-schema read from subscriptions + plans — approved for owner subscription info
	row, err := r.queries.GetOwnerSubscriptionInfo(ctx, ownerID)
	if err != nil {
		return "", "", err
	}
	status := ""
	if row.Status.Valid {
		status = string(row.Status.SubscriptionsStatus)
	}
	return row.Slug, status, nil
}

// RecordSeatChange updates the billing seat_periods table after a team member is added or removed.
// All three writes (close the active period, open a new one, update the subscription seat count)
// run inside a single transaction so billing state can never end up partially applied.
func (r *Repository) RecordSeatChange(ctx context.Context, ownerID int64, newCount int) error {
	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1)
	billingMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	// cross-schema read from subscriptions — approved for seat period management
	sub, err := qtx.LockSubscriptionForSeatChange(ctx, ownerID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// cross-schema write to seat_periods — approved for seat billing accuracy
	if err := qtx.CloseActiveSeatPeriod(ctx, sqlc.CloseActiveSeatPeriodParams{
		PeriodEnd:      sql.NullTime{Time: yesterday, Valid: true},
		SubscriptionID: sub.ID,
	}); err != nil {
		_ = tx.Rollback()
		return err
	}

	// cross-schema write to seat_periods — approved for seat billing accuracy
	if err := qtx.OpenSeatPeriod(ctx, sqlc.OpenSeatPeriodParams{
		SubscriptionID:    sub.ID,
		OwnerID:           ownerID,
		SeatCount:         int32(newCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PricePerSeatCents: sub.PricePerSeatCents,
		PeriodStart:       now,
		BillingMonth:      billingMonth,
	}); err != nil {
		_ = tx.Rollback()
		return err
	}

	// cross-schema write to subscriptions — approved for seat billing accuracy
	if err := qtx.UpdateSubscriptionSeatCount(ctx, sqlc.UpdateSubscriptionSeatCountParams{
		SeatCount: int32(newCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		OwnerID:   ownerID,
	}); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// ─── Access Requests ──────────────────────────────────────────────────────────

func (r *Repository) ListAccessRequests(ctx context.Context, businessID int64) ([]domain.AccessRequest, error) {
	rows, err := r.queries.ListAccessRequestsByBusiness(ctx, businessID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.AccessRequest, 0, len(rows))
	for _, row := range rows {
		status := "pending"
		if row.Status.Valid {
			status = string(row.Status.AccessRequestsStatus)
		}
		out = append(out, domain.AccessRequest{
			ID:             row.ID,
			UserID:         row.UserID,
			App:            row.App,
			Status:         status,
			RequesterName:  row.RequesterName,
			RequesterEmail: row.RequesterEmail,
			CreatedAt:      row.CreatedAt.Time,
		})
	}
	return out, nil
}

// GetAdminsAndOwnerUserIDs returns all user IDs that should receive management
// notifications for a business: the primary owner plus any staff with
// role in ('admin','owner') in user_app_permissions.
func (r *Repository) GetAdminsAndOwnerUserIDs(ctx context.Context, businessID int64) ([]int64, error) {
	// cross-schema read from hub — approved for business-wide admin notification fan-out
	return r.queries.ListAdminsAndOwnerUserIDs(ctx, sqlc.ListAdminsAndOwnerUserIDsParams{
		ID:         businessID,
		BusinessID: sql.NullInt64{Int64: businessID, Valid: true},
	})
}

// BulkCreateHubNotification writes an in-app notification to hub_notifications for
// each recipient in userIDs. Best-effort: errors are logged, not propagated.
func (r *Repository) ResolveAccessRequest(ctx context.Context, id, businessID int64, status string) error {
	return r.queries.ResolveAccessRequest(ctx, sqlc.ResolveAccessRequestParams{
		ID:         id,
		BusinessID: businessID,
		Status: sqlc.NullAccessRequestsStatus{
			AccessRequestsStatus: sqlc.AccessRequestsStatus(status),
			Valid:                true,
		},
	})
}

func (r *Repository) BulkCreateHubNotification(ctx context.Context, ownerID int64, userIDs []int64, title, body, notifType, link string) error {
	for _, uid := range userIDs {
		// cross-schema write to hub_notifications — approved for cross-module access request alerts
		if err := r.queries.CreateHubNotification(ctx, sqlc.CreateHubNotificationParams{
			OwnerID: ownerID,
			UserID:  uid,
			Title:   title,
			Body:    sql.NullString{String: body, Valid: body != ""},
			Type:    notifType,
			Link:    sql.NullString{String: link, Valid: link != ""},
		}); err != nil {
			logger.Ctx(ctx).Warn("failed to create hub notification",
				logger.Int("user_id", int(uid)), logger.Error(err))
		}
	}
	return nil
}

// nullBusinessID wraps the join-side business id so sqlc's named parameter
// (shared by both placeholders in RevokeBusinessUserPermissions) scans as
// NOT NULL int64.
func nullBusinessID(v int64) sql.NullInt64 { return sql.NullInt64{Int64: v, Valid: true} }
