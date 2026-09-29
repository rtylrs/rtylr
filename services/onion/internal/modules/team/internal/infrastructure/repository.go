package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
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
					// Validate before inserting. The scopes table's CHECK constraint
					// would reject the row too late: by this point the member's
					// licenses and role assignments have already been revoked in this
					// same transaction, so a bad scope would roll back their whole
					// authorization and surface as an opaque 500.
					if err = authorization.ValidateAssignmentScope(ctx, tx, businessID, scope.Type, scope.ID); err != nil {
						return rollback(fmt.Errorf("invalid scope %q:%q: %w", scope.Type, scope.ID, err))
					}
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
