package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/authorization"
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

// UpsertNormalizedAuthorization atomically replaces a member's direct licenses,
// permission-profile assignments, group memberships, and authorization record.
func (r *Repository) UpsertNormalizedAuthorization(ctx context.Context, businessID, ownerID, actorID, userID int64, apps []string, memberType, memberStatus string, assignments []domain.AppRoleAssignmentInput, groupIDs []int64) error {
	if businessID <= 0 || actorID <= 0 || userID <= 0 {
		return errors.New("business, actor, and user are required")
	}
	if memberType == "" {
		memberType = "member"
	}
	if memberType != "member" && memberType != "admin" {
		return errors.New("invalid member type")
	}
	if memberStatus == "" {
		memberStatus = "active"
	}
	if memberStatus != "pending" && memberStatus != "active" {
		return errors.New("invalid membership status")
	}

	apps = uniqueAuthorizationApps(apps)
	for _, appID := range apps {
		if !isTenantAppID(appID) {
			return fmt.Errorf("invalid app id: %s", appID)
		}
	}
	for _, assignment := range assignments {
		if assignment.AppID == "" || (assignment.AppID != "dash" && !isTenantAppID(assignment.AppID)) || len(assignment.RoleIDs) == 0 {
			return errors.New("permission profile assignment requires an app and at least one profile")
		}
		if assignment.GroupID != nil || assignment.Source == string(authorization.SourceGroup) || assignment.Source == string(authorization.SourceSystem) {
			return errors.New("member assignments must be direct permission profiles")
		}
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(cause error) error { _ = tx.Rollback(); return cause }
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO authorization_business_members
			(business_id, user_id, member_type, status, authorization_version)
		VALUES (?, ?, ?, ?, 1)
		ON DUPLICATE KEY UPDATE member_type = VALUES(member_type), status = VALUES(status),
			authorization_version = authorization_version + 1, updated_at = NOW()`,
		businessID, userID, memberType, memberStatus); err != nil {
		return rollback(err)
	}
	if err = r.replaceUserGroupMemberships(ctx, tx, businessID, ownerID, actorID, userID, groupIDs); err != nil {
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

	for _, appID := range apps {
		if appID == "dash" {
			continue
		}
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO authorization_license_assignments
				(business_id, user_id, app_id, source, status, granted_by)
			VALUES (?, ?, ?, 'explicit', ?, ?)`, businessID, userID, appID, memberStatus, actorID); err != nil {
			return rollback(err)
		}
	}

	for _, assignment := range assignments {
		for _, profileID := range uniqueInt64(assignment.RoleIDs) {
			if err = r.validateAuthorizationRoleID(ctx, tx, businessID, assignment.AppID, profileID); err != nil {
				return rollback(fmt.Errorf("permission profile %d is invalid for %s: %w", profileID, assignment.AppID, err))
			}
			result, insertErr := tx.ExecContext(ctx, `
				INSERT INTO authorization_role_assignments
					(business_id, user_id, app_id, role_id, source, status, is_primary, granted_by)
				VALUES (?, ?, ?, ?, 'explicit', ?, FALSE, ?)`,
				businessID, userID, assignment.AppID, profileID, memberStatus, actorID)
			if insertErr != nil {
				return rollback(insertErr)
			}
			assignmentID, insertErr := result.LastInsertId()
			if insertErr != nil {
				return rollback(insertErr)
			}
			for _, scope := range assignment.Scopes {
				if err = authorization.ValidateAssignmentScope(ctx, tx, businessID, scope.Type, scope.ID); err != nil {
					return rollback(fmt.Errorf("invalid scope %q:%q: %w", scope.Type, scope.ID, err))
				}
				if _, err = tx.ExecContext(ctx, `
					INSERT INTO authorization_role_assignment_scopes (assignment_id, scope_type, scope_id)
					VALUES (?, ?, ?)`, assignmentID, scope.Type, scope.ID); err != nil {
					return rollback(err)
				}
			}
		}
	}

	auditAfter, err := json.Marshal(struct {
		MemberType string   `json:"member_type"`
		Apps       []string `json:"licensed_app_ids"`
		Profiles   int      `json:"permission_profile_assignments"`
	}{MemberType: memberType, Apps: apps, Profiles: len(assignments)})
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
