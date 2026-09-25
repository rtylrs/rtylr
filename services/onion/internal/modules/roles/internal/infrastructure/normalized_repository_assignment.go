package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"

	"github.com/voxire/onion/internal/modules/roles/internal/domain"
)

func (r *Repository) RemoveAssignment(ctx context.Context, businessID, actorID, assignmentID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return normalizedSchemaError(err)
	}
	var roleID int64
	var appID string
	var userID, groupID sql.NullInt64
	err = tx.QueryRowContext(ctx, `
		SELECT role_id, app_id, user_id, group_id
		FROM authorization_role_assignments
		WHERE id = ? AND business_id = ? AND status = 'active'
		FOR UPDATE`, assignmentID, businessID).Scan(&roleID, &appID, &userID, &groupID)
	if errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return domain.ErrRoleNotFound
	}
	if err != nil {
		_ = tx.Rollback()
		return normalizedSchemaError(err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE authorization_role_assignments SET status = 'revoked', revoked_by = ?, revoked_at = NOW() WHERE id = ? AND business_id = ?`, actorID, assignmentID, businessID); err != nil {
		_ = tx.Rollback()
		return normalizedSchemaError(err)
	}
	if err := insertAuditEvent(ctx, tx, businessID, actorID, "role.assignment.revoked", appID, &roleID, nullableInt64(userID), nullableInt64(groupID), "", "", ""); err != nil {
		_ = tx.Rollback()
		return normalizedSchemaError(err)
	}
	if err := tx.Commit(); err != nil {
		return normalizedSchemaError(err)
	}
	return nil
}

func (r *Repository) PreviewPermissionImpact(ctx context.Context, businessID, roleID int64, permissions []string) (domain.PermissionImpact, error) {
	role, err := r.GetNormalized(ctx, businessID, roleID)
	if err != nil {
		return domain.PermissionImpact{}, err
	}
	oldSet := make(map[string]struct{}, len(role.Permissions))
	for _, permission := range role.Permissions {
		oldSet[permission] = struct{}{}
	}
	newSet := make(map[string]struct{}, len(permissions))
	for _, permission := range permissions {
		newSet[permission] = struct{}{}
	}
	impact := domain.PermissionImpact{AddedPermissions: []string{}, RemovedPermissions: []string{}}
	for permission := range newSet {
		if _, found := oldSet[permission]; !found {
			impact.AddedPermissions = append(impact.AddedPermissions, permission)
		}
	}
	for permission := range oldSet {
		if _, found := newSet[permission]; !found {
			impact.RemovedPermissions = append(impact.RemovedPermissions, permission)
		}
	}
	sort.Strings(impact.AddedPermissions)
	sort.Strings(impact.RemovedPermissions)
	if err := r.pool.QueryRowContext(ctx, `SELECT COUNT(DISTINCT id) FROM authorization_role_assignments WHERE business_id = ? AND role_id = ? AND status = 'active'`, businessID, roleID).Scan(&impact.AffectedMemberCount); err != nil {
		return domain.PermissionImpact{}, normalizedSchemaError(err)
	}
	return impact, nil
}

func (r *Repository) ListAudit(ctx context.Context, businessID int64, filter domain.AuditFilter) (domain.AuditPage, error) {
	page, pageSize := normalizedPage(filter.Page, filter.PageSize)
	offset := (page - 1) * pageSize
	var total int64
	var err error
	if filter.RoleID == nil {
		err = r.pool.QueryRowContext(ctx, `SELECT COUNT(*) FROM authorization_audit_events WHERE business_id = ?`, businessID).Scan(&total)
	} else {
		err = r.pool.QueryRowContext(ctx, `SELECT COUNT(*) FROM authorization_audit_events WHERE business_id = ? AND role_id = ?`, businessID, *filter.RoleID).Scan(&total)
	}
	if err != nil {
		return domain.AuditPage{}, normalizedSchemaError(err)
	}
	query := `
		SELECT id, action, app_id, role_id, actor_user_id, target_user_id,
		       target_group_id, before_json, after_json, reason, created_at
		FROM authorization_audit_events
		WHERE business_id = ?`
	args := []interface{}{businessID}
	if filter.RoleID != nil {
		query += " AND role_id = ?"
		args = append(args, *filter.RoleID)
	}
	query += " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return domain.AuditPage{}, normalizedSchemaError(err)
	}
	defer rows.Close()
	events := make([]domain.AuditEvent, 0)
	for rows.Next() {
		var event domain.AuditEvent
		var appID, beforeJSON, afterJSON, reason sql.NullString
		var roleID, actorID, targetUserID, targetGroupID sql.NullInt64
		if err := rows.Scan(&event.ID, &event.Action, &appID, &roleID, &actorID, &targetUserID, &targetGroupID, &beforeJSON, &afterJSON, &reason, &event.CreatedAt); err != nil {
			return domain.AuditPage{}, err
		}
		event.AppID = appID.String
		event.RoleID = nullableInt64(roleID)
		event.ActorUserID = nullableInt64(actorID)
		event.TargetUserID = nullableInt64(targetUserID)
		event.TargetGroupID = nullableInt64(targetGroupID)
		event.BeforeJSON = beforeJSON.String
		event.AfterJSON = afterJSON.String
		event.Reason = reason.String
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return domain.AuditPage{}, err
	}
	return domain.AuditPage{Events: events, Page: page, PageSize: pageSize, Total: total}, nil
}

func (r *Repository) loadRolePermissions(ctx context.Context, roleID int64, role *domain.Role) error {
	rows, err := r.pool.QueryContext(ctx, `SELECT permission_key FROM authorization_role_permissions WHERE role_id = ? ORDER BY permission_key`, roleID)
	if err != nil {
		return normalizedSchemaError(err)
	}
	defer rows.Close()
	role.Permissions = make([]string, 0)
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return err
		}
		role.Permissions = append(role.Permissions, permission)
	}
	return rows.Err()
}

func (r *Repository) loadAssignmentScopes(ctx context.Context, assignmentID int64) ([]string, error) {
	rows, err := r.pool.QueryContext(ctx, `SELECT scope_type, scope_id FROM authorization_role_assignment_scopes WHERE assignment_id = ? ORDER BY scope_type, scope_id`, assignmentID)
	if err != nil {
		return nil, normalizedSchemaError(err)
	}
	defer rows.Close()
	scopes := make([]string, 0)
	for rows.Next() {
		var scopeType, scopeID string
		if err := rows.Scan(&scopeType, &scopeID); err != nil {
			return nil, err
		}
		scopes = append(scopes, scopeType+":"+scopeID)
	}
	return scopes, rows.Err()
}

func (r *Repository) getAssignment(ctx context.Context, businessID, assignmentID int64) (domain.RoleAssignment, error) {
	var assignment domain.RoleAssignment
	var userID, groupID sql.NullInt64
	err := r.pool.QueryRowContext(ctx, `
		SELECT id, role_id, app_id, user_id, group_id, status, source, is_primary, granted_at
		FROM authorization_role_assignments
		WHERE id = ? AND business_id = ?`, assignmentID, businessID).
		Scan(&assignment.ID, &assignment.RoleID, &assignment.AppID, &userID, &groupID, &assignment.Status, &assignment.Source, &assignment.IsPrimary, &assignment.GrantedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.RoleAssignment{}, domain.ErrRoleNotFound
	}
	if err != nil {
		return domain.RoleAssignment{}, normalizedSchemaError(err)
	}
	assignment.UserID = nullableInt64(userID)
	assignment.GroupID = nullableInt64(groupID)
	assignment.Scopes, err = r.loadAssignmentScopes(ctx, assignment.ID)
	return assignment, err
}

func scanNormalizedRole(scanner interface {
	Scan(dest ...interface{}) error
},
) (domain.Role, error) {
	var role domain.Role
	var scopeJSON sql.NullString
	var parentRoleID sql.NullInt64
	var createdAt, updatedAt sql.NullTime
	if err := scanner.Scan(
		&role.ID, &role.BusinessID, &role.AppID, &role.RoleKey, &role.Name,
		&role.Description, &scopeJSON, &role.Level, &parentRoleID,
		&role.IsSystem, &role.IsArchived, &role.Lifecycle, &role.Version,
		&createdAt, &updatedAt, &role.MemberCount,
	); err != nil {
		return domain.Role{}, err
	}
	role.ParentRoleID = nullableInt64(parentRoleID)
	if scopeJSON.Valid && scopeJSON.String != "" {
		if err := json.Unmarshal([]byte(scopeJSON.String), &role.ScopeRules); err != nil {
			return domain.Role{}, err
		}
	}
	role.ScopeRules = nonNilRoleStrings(role.ScopeRules)
	if createdAt.Valid {
		role.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		role.UpdatedAt = updatedAt.Time
	}
	return role, nil
}
