package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/roles/internal/domain"
)

var roleKeyPartPattern = regexp.MustCompile(`[^a-z0-9]+`)

func (r *Repository) ListNormalized(ctx context.Context, businessID int64, filter domain.RoleFilter) (domain.RolePage, error) {
	if businessID <= 0 {
		return domain.RolePage{}, domain.ErrRoleNotFound
	}
	page, pageSize := normalizedPage(filter.Page, filter.PageSize)
	offset := (page - 1) * pageSize

	var total int64
	var err error
	if filter.AppID == "" {
		err = r.pool.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM authorization_roles
			WHERE business_id IN (0, ?)
			  AND lifecycle = 'active' AND is_archived = FALSE`, businessID).Scan(&total)
	} else {
		err = r.pool.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM authorization_roles
			WHERE business_id IN (0, ?) AND app_id = ?
			  AND lifecycle = 'active' AND is_archived = FALSE`, businessID, filter.AppID).Scan(&total)
	}
	if err != nil {
		return domain.RolePage{}, normalizedSchemaError(err)
	}

	query := `
		SELECT r.id, r.business_id, r.app_id, r.role_key, r.display_name,
		       r.description, r.scope_rules, r.level, r.parent_role_id,
		       r.is_system, r.is_archived, r.lifecycle, r.version,
		       r.created_at, r.updated_at,
		       COUNT(DISTINCT a.id)
		FROM authorization_roles r
		LEFT JOIN authorization_role_assignments a
		  ON a.role_id = r.id AND a.business_id = ? AND a.status = 'active'
		WHERE r.business_id IN (0, ?)
		  AND r.lifecycle = 'active' AND r.is_archived = FALSE`
	args := []interface{}{businessID, businessID}
	if filter.AppID != "" {
		query += " AND r.app_id = ?"
		args = append(args, filter.AppID)
	}
	query += " GROUP BY r.id ORDER BY r.updated_at DESC, r.id DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	rows, queryErr := r.pool.QueryContext(ctx, query, args...)
	if queryErr != nil {
		return domain.RolePage{}, normalizedSchemaError(queryErr)
	}
	defer rows.Close()

	roles := make([]domain.Role, 0)
	for rows.Next() {
		role, scanErr := scanNormalizedRole(rows)
		if scanErr != nil {
			return domain.RolePage{}, scanErr
		}
		if err := r.loadRolePermissions(ctx, role.ID, &role); err != nil {
			return domain.RolePage{}, err
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return domain.RolePage{}, err
	}
	return domain.RolePage{Roles: roles, Page: page, PageSize: pageSize, Total: total}, nil
}

func (r *Repository) GetNormalized(ctx context.Context, businessID, roleID int64) (domain.Role, error) {
	row := r.pool.QueryRowContext(ctx, `
		SELECT r.id, r.business_id, r.app_id, r.role_key, r.display_name,
		       r.description, r.scope_rules, r.level, r.parent_role_id,
		       r.is_system, r.is_archived, r.lifecycle, r.version,
		       r.created_at, r.updated_at,
		       COUNT(DISTINCT a.id)
		FROM authorization_roles r
		LEFT JOIN authorization_role_assignments a
		  ON a.role_id = r.id AND a.business_id = ? AND a.status = 'active'
		WHERE r.id = ? AND r.business_id IN (0, ?)
		GROUP BY r.id`, businessID, roleID, businessID)
	role, err := scanNormalizedRole(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Role{}, domain.ErrRoleNotFound
	}
	if err != nil {
		return domain.Role{}, normalizedSchemaError(err)
	}
	if err := r.loadRolePermissions(ctx, role.ID, &role); err != nil {
		return domain.Role{}, err
	}
	return role, nil
}

func (r *Repository) CreateNormalized(ctx context.Context, businessID, actorID int64, input domain.CreateRoleInput) (domain.Role, error) {
	if businessID <= 0 {
		return domain.Role{}, domain.ErrRoleNotFound
	}
	appID := input.AppID
	if appID == "" {
		appID = "dash"
	}
	if !validAppID(appID) {
		return domain.Role{}, errors.New("invalid app id")
	}
	if err := validateRolePermissions(ctx, r.pool, appID, input.Permissions); err != nil {
		return domain.Role{}, err
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return domain.Role{}, normalizedSchemaError(err)
	}
	rollback := func(cause error) (domain.Role, error) {
		_ = tx.Rollback()
		return domain.Role{}, normalizedSchemaError(cause)
	}
	if err := validateParentRole(ctx, tx, businessID, appID, input.ParentRoleID); err != nil {
		return rollback(err)
	}
	roleKey, err := nextRoleKey(ctx, tx, businessID, appID, input.RoleKey, input.Name)
	if err != nil {
		return rollback(err)
	}
	scopeJSON, err := json.Marshal(nonNilRoleStrings(input.ScopeRules))
	if err != nil {
		return rollback(err)
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO authorization_roles
			(business_id, app_id, role_key, display_name, description, scope_rules,
			 level, parent_role_id, assignment_mode, is_system, is_archived,
			 lifecycle, version, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'many', FALSE, FALSE, 'active', 1, ?)`,
		businessID, appID, roleKey, input.Name, input.Description, scopeJSON,
		input.Level, input.ParentRoleID, actorID,
	)
	if err != nil {
		return rollback(err)
	}
	roleID, err := result.LastInsertId()
	if err != nil {
		return rollback(err)
	}
	if err := replaceRolePermissions(ctx, tx, roleID, input.Permissions); err != nil {
		return rollback(err)
	}
	afterJSON, err := json.Marshal(input)
	if err != nil {
		return rollback(err)
	}
	if err := insertAuditEvent(ctx, tx, businessID, actorID, "role.created", appID, &roleID, nil, nil, "", string(afterJSON), ""); err != nil {
		return rollback(err)
	}
	if err := tx.Commit(); err != nil {
		return domain.Role{}, normalizedSchemaError(err)
	}
	return r.GetNormalized(ctx, businessID, roleID)
}

func (r *Repository) UpdateNormalized(ctx context.Context, businessID, actorID int64, input domain.UpdateRoleInput) (domain.Role, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return domain.Role{}, normalizedSchemaError(err)
	}
	rollback := func(cause error) (domain.Role, error) {
		_ = tx.Rollback()
		return domain.Role{}, normalizedSchemaError(cause)
	}
	role, err := scanNormalizedRole(tx.QueryRowContext(ctx, `
		SELECT r.id, r.business_id, r.app_id, r.role_key, r.display_name,
		       r.description, r.scope_rules, r.level, r.parent_role_id,
		       r.is_system, r.is_archived, r.lifecycle, r.version,
		       r.created_at, r.updated_at,
		       COUNT(DISTINCT a.id)
		FROM authorization_roles r
		LEFT JOIN authorization_role_assignments a
		  ON a.role_id = r.id AND a.business_id = ? AND a.status = 'active'
		WHERE r.id = ? AND r.business_id = ?
		GROUP BY r.id
		FOR UPDATE`, businessID, input.ID, businessID))
	if errors.Is(err, sql.ErrNoRows) {
		return rollback(domain.ErrRoleNotFound)
	}
	if err != nil {
		return rollback(err)
	}
	if role.IsSystem {
		return rollback(domain.ErrSystemRoleMutation)
	}
	if role.IsArchived || role.Lifecycle != "active" {
		return rollback(domain.ErrRoleArchived)
	}
	if input.AppID != "" && input.AppID != role.AppID {
		return rollback(errors.New("role app cannot be changed"))
	}
	if err := validateRolePermissions(ctx, tx, role.AppID, input.Permissions); err != nil {
		return rollback(err)
	}
	if err := validateParentRole(ctx, tx, businessID, role.AppID, input.ParentRoleID); err != nil {
		return rollback(err)
	}
	scopeJSON, err := json.Marshal(nonNilRoleStrings(input.ScopeRules))
	if err != nil {
		return rollback(err)
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE authorization_roles
		SET display_name = ?, description = ?, scope_rules = ?, level = ?,
		    parent_role_id = ?, version = version + 1, updated_at = NOW()
		WHERE id = ? AND business_id = ? AND is_system = FALSE
		  AND lifecycle = 'active'`,
		input.Name, input.Description, scopeJSON, input.Level, input.ParentRoleID,
		input.ID, businessID,
	)
	if err != nil {
		return rollback(err)
	}
	if err := replaceRolePermissions(ctx, tx, input.ID, input.Permissions); err != nil {
		return rollback(err)
	}
	beforeJSON, err := json.Marshal(role)
	if err != nil {
		return rollback(err)
	}
	afterJSON, err := json.Marshal(input)
	if err != nil {
		return rollback(err)
	}
	if err := insertAuditEvent(ctx, tx, businessID, actorID, "role.updated", role.AppID, &input.ID, nil, nil, string(beforeJSON), string(afterJSON), ""); err != nil {
		return rollback(err)
	}
	if err := tx.Commit(); err != nil {
		return domain.Role{}, normalizedSchemaError(err)
	}
	return r.GetNormalized(ctx, businessID, input.ID)
}

func (r *Repository) ArchiveNormalized(ctx context.Context, businessID, actorID, roleID int64, reason string) (domain.Role, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return domain.Role{}, normalizedSchemaError(err)
	}
	rollback := func(cause error) (domain.Role, error) {
		_ = tx.Rollback()
		return domain.Role{}, normalizedSchemaError(cause)
	}
	role, err := scanNormalizedRole(tx.QueryRowContext(ctx, `
		SELECT r.id, r.business_id, r.app_id, r.role_key, r.display_name,
		       r.description, r.scope_rules, r.level, r.parent_role_id,
		       r.is_system, r.is_archived, r.lifecycle, r.version,
		       r.created_at, r.updated_at,
		       COUNT(DISTINCT a.id)
		FROM authorization_roles r
		LEFT JOIN authorization_role_assignments a
		  ON a.role_id = r.id AND a.business_id = ? AND a.status = 'active'
		WHERE r.id = ? AND r.business_id = ?
		GROUP BY r.id
		FOR UPDATE`, businessID, roleID, businessID))
	if errors.Is(err, sql.ErrNoRows) {
		return rollback(domain.ErrRoleNotFound)
	}
	if err != nil {
		return rollback(err)
	}
	if role.IsSystem {
		return rollback(domain.ErrSystemRoleMutation)
	}
	if role.IsArchived || role.Lifecycle != "active" {
		return rollback(domain.ErrRoleArchived)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_roles
		SET is_archived = TRUE, lifecycle = 'archived', version = version + 1, updated_at = NOW()
		WHERE id = ? AND business_id = ?`, roleID, businessID); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_role_assignments
		SET status = 'revoked', revoked_by = ?, revoked_at = NOW()
		WHERE business_id = ? AND role_id = ? AND status = 'active'`, actorID, businessID, roleID); err != nil {
		return rollback(err)
	}
	beforeJSON, err := json.Marshal(role)
	if err != nil {
		return rollback(err)
	}
	if err := insertAuditEvent(ctx, tx, businessID, actorID, "role.archived", role.AppID, &roleID, nil, nil, string(beforeJSON), "", reason); err != nil {
		return rollback(err)
	}
	if err := tx.Commit(); err != nil {
		return domain.Role{}, normalizedSchemaError(err)
	}
	return r.GetNormalized(ctx, businessID, roleID)
}

func (r *Repository) CloneNormalized(ctx context.Context, businessID, actorID, roleID int64, name, description string) (domain.Role, error) {
	source, err := r.GetNormalized(ctx, businessID, roleID)
	if err != nil {
		return domain.Role{}, err
	}
	return r.CreateNormalized(ctx, businessID, actorID, domain.CreateRoleInput{
		Name:         name,
		Description:  description,
		AppID:        source.AppID,
		Level:        source.Level,
		ParentRoleID: source.ParentRoleID,
		ScopeRules:   append([]string{}, source.ScopeRules...),
		Permissions:  append([]string{}, source.Permissions...),
	})
}

func (r *Repository) ListAssignments(ctx context.Context, businessID int64, filter domain.AssignmentFilter) (domain.RoleAssignmentPage, error) {
	page, pageSize := normalizedPage(filter.Page, filter.PageSize)
	offset := (page - 1) * pageSize
	var total int64
	var err error
	if filter.RoleID == nil && filter.AppID == "" {
		err = r.pool.QueryRowContext(ctx, `SELECT COUNT(*) FROM authorization_role_assignments WHERE business_id = ? AND status = 'active'`, businessID).Scan(&total)
	} else if filter.RoleID != nil && filter.AppID == "" {
		err = r.pool.QueryRowContext(ctx, `SELECT COUNT(*) FROM authorization_role_assignments WHERE business_id = ? AND role_id = ? AND status = 'active'`, businessID, *filter.RoleID).Scan(&total)
	} else if filter.RoleID == nil {
		err = r.pool.QueryRowContext(ctx, `SELECT COUNT(*) FROM authorization_role_assignments WHERE business_id = ? AND app_id = ? AND status = 'active'`, businessID, filter.AppID).Scan(&total)
	} else {
		err = r.pool.QueryRowContext(ctx, `SELECT COUNT(*) FROM authorization_role_assignments WHERE business_id = ? AND role_id = ? AND app_id = ? AND status = 'active'`, businessID, *filter.RoleID, filter.AppID).Scan(&total)
	}
	if err != nil {
		return domain.RoleAssignmentPage{}, normalizedSchemaError(err)
	}
	query := `
		SELECT id, role_id, app_id, user_id, group_id, status, source, is_primary, granted_at
		FROM authorization_role_assignments
		WHERE business_id = ? AND status = 'active'`
	args := []interface{}{businessID}
	if filter.RoleID != nil {
		query += " AND role_id = ?"
		args = append(args, *filter.RoleID)
	}
	if filter.AppID != "" {
		query += " AND app_id = ?"
		args = append(args, filter.AppID)
	}
	query += " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return domain.RoleAssignmentPage{}, normalizedSchemaError(err)
	}
	defer rows.Close()
	assignments := make([]domain.RoleAssignment, 0)
	for rows.Next() {
		var assignment domain.RoleAssignment
		var userID, groupID sql.NullInt64
		if err := rows.Scan(&assignment.ID, &assignment.RoleID, &assignment.AppID, &userID, &groupID, &assignment.Status, &assignment.Source, &assignment.IsPrimary, &assignment.GrantedAt); err != nil {
			return domain.RoleAssignmentPage{}, err
		}
		assignment.UserID = nullableInt64(userID)
		assignment.GroupID = nullableInt64(groupID)
		assignment.Scopes, err = r.loadAssignmentScopes(ctx, assignment.ID)
		if err != nil {
			return domain.RoleAssignmentPage{}, err
		}
		assignments = append(assignments, assignment)
	}
	if err := rows.Err(); err != nil {
		return domain.RoleAssignmentPage{}, err
	}
	return domain.RoleAssignmentPage{Assignments: assignments, Page: page, PageSize: pageSize, Total: total}, nil
}

func (r *Repository) AssignRole(ctx context.Context, businessID, actorID int64, input domain.AssignRoleInput) (domain.RoleAssignment, error) {
	if (input.UserID == nil) == (input.GroupID == nil) {
		return domain.RoleAssignment{}, domain.ErrInvalidAssignmentTarget
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return domain.RoleAssignment{}, normalizedSchemaError(err)
	}
	rollback := func(cause error) (domain.RoleAssignment, error) {
		_ = tx.Rollback()
		return domain.RoleAssignment{}, normalizedSchemaError(cause)
	}
	var roleID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM authorization_roles
		WHERE id = ? AND business_id IN (0, ?) AND app_id = ?
		  AND lifecycle = 'active' AND is_archived = FALSE`, input.RoleID, businessID, input.AppID).Scan(&roleID)
	if errors.Is(err, sql.ErrNoRows) {
		return rollback(domain.ErrRoleNotFound)
	}
	if err != nil {
		return rollback(err)
	}
	if input.UserID != nil {
		var memberID int64
		err = tx.QueryRowContext(ctx, `
			SELECT id FROM authorization_business_members
			WHERE business_id = ? AND user_id = ? AND status = 'active'`, businessID, *input.UserID).Scan(&memberID)
		if errors.Is(err, sql.ErrNoRows) {
			return rollback(domain.ErrInvalidAssignmentTarget)
		}
		if err != nil {
			return rollback(err)
		}
	}
	if input.GroupID != nil {
		var groupID int64
		err = tx.QueryRowContext(ctx, `SELECT id FROM user_groups WHERE id = ? AND business_id = ?`, *input.GroupID, businessID).Scan(&groupID)
		if errors.Is(err, sql.ErrNoRows) {
			return rollback(domain.ErrInvalidAssignmentTarget)
		}
		if err != nil {
			return rollback(err)
		}
	}
	if input.IsPrimary {
		if input.UserID != nil {
			if _, err := tx.ExecContext(ctx, `UPDATE authorization_role_assignments SET is_primary = FALSE WHERE business_id = ? AND user_id = ? AND app_id = ? AND status = 'active'`, businessID, *input.UserID, input.AppID); err != nil {
				return rollback(err)
			}
		}
		if input.GroupID != nil {
			if _, err := tx.ExecContext(ctx, `UPDATE authorization_role_assignments SET is_primary = FALSE WHERE business_id = ? AND group_id = ? AND app_id = ? AND status = 'active'`, businessID, *input.GroupID, input.AppID); err != nil {
				return rollback(err)
			}
		}
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO authorization_role_assignments
			(business_id, user_id, group_id, app_id, role_id, source, status, is_primary, granted_by)
		VALUES (?, ?, ?, ?, ?, 'explicit', 'active', ?, ?)`,
		businessID, input.UserID, input.GroupID, input.AppID, roleID, input.IsPrimary, actorID)
	if err != nil {
		return rollback(err)
	}
	assignmentID, err := result.LastInsertId()
	if err != nil {
		return rollback(err)
	}
	if err := insertAssignmentScopes(ctx, tx, businessID, assignmentID, input.Scopes); err != nil {
		return rollback(err)
	}
	afterJSON, err := json.Marshal(input)
	if err != nil {
		return rollback(err)
	}
	if err := insertAuditEvent(ctx, tx, businessID, actorID, "role.assignment.created", input.AppID, &roleID, input.UserID, input.GroupID, "", string(afterJSON), ""); err != nil {
		return rollback(err)
	}
	if err := tx.Commit(); err != nil {
		return domain.RoleAssignment{}, normalizedSchemaError(err)
	}
	return r.getAssignment(ctx, businessID, assignmentID)
}

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
}) (domain.Role, error) {
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

func validateRolePermissions(ctx context.Context, executor interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}, appID string, permissions []string) error {
	for _, permission := range permissions {
		var catalogApp string
		err := executor.QueryRowContext(ctx, `SELECT app_id FROM authorization_permission_catalog WHERE permission_key = ? LIMIT 1`, permission).Scan(&catalogApp)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrInvalidPermission
			}
			return normalizedSchemaError(err)
		}
		if catalogApp != appID {
			return domain.ErrInvalidPermission
		}
	}
	return nil
}

func validateParentRole(ctx context.Context, tx *sql.Tx, businessID int64, appID string, parentRoleID *int64) error {
	if parentRoleID == nil {
		return nil
	}
	var id int64
	err := tx.QueryRowContext(ctx, `
		SELECT id FROM authorization_roles
		WHERE id = ? AND business_id IN (0, ?) AND app_id = ?
		  AND lifecycle = 'active' AND is_archived = FALSE`, *parentRoleID, businessID, appID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrRoleNotFound
	}
	return err
}

func nextRoleKey(ctx context.Context, tx *sql.Tx, businessID int64, appID, requested, name string) (string, error) {
	base := requested
	if base == "" {
		base = "custom." + roleKeyPartPattern.ReplaceAllString(strings.ToLower(strings.TrimSpace(name)), "-")
	}
	if !strings.HasPrefix(base, "custom.") {
		base = "custom." + base
	}
	base = roleKeyPartPattern.ReplaceAllString(strings.ToLower(base), "-")
	base = strings.Trim(base, "-")
	if base == "custom" || len(base) > 96 {
		return "", errors.New("invalid role key")
	}
	for suffix := 1; suffix <= 1000; suffix++ {
		candidate := base
		if suffix > 1 {
			candidate = fmt.Sprintf("%s-%d", base, suffix)
		}
		var count int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM authorization_roles WHERE business_id = ? AND app_id = ? AND role_key = ?`, businessID, appID, candidate).Scan(&count); err != nil {
			return "", err
		}
		if count == 0 {
			return candidate, nil
		}
	}
	return "", errors.New("could not allocate role key")
}

func replaceRolePermissions(ctx context.Context, tx *sql.Tx, roleID int64, permissions []string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM authorization_role_permissions WHERE role_id = ?`, roleID); err != nil {
		return err
	}
	for _, permission := range permissions {
		if _, err := tx.ExecContext(ctx, `INSERT INTO authorization_role_permissions (role_id, permission_key) VALUES (?, ?)`, roleID, permission); err != nil {
			return err
		}
	}
	return nil
}

func insertAssignmentScopes(ctx context.Context, tx *sql.Tx, businessID, assignmentID int64, scopes []string) error {
	for _, scope := range scopes {
		parts := strings.SplitN(scope, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return errors.New("scope must be formatted as type:id")
		}
		if err := validateAssignmentScope(ctx, tx, businessID, parts[0], parts[1]); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO authorization_role_assignment_scopes (assignment_id, scope_type, scope_id) VALUES (?, ?, ?)`, assignmentID, parts[0], parts[1]); err != nil {
			return err
		}
	}
	return nil
}

func validateAssignmentScope(ctx context.Context, tx *sql.Tx, businessID int64, scopeType, scopeID string) error {
	id, err := strconv.ParseInt(scopeID, 10, 64)
	if err != nil || id <= 0 {
		return domain.ErrInvalidScope
	}
	switch scopeType {
	case "business":
		if id != businessID {
			return domain.ErrInvalidScope
		}
		return nil
	case "pos_location":
		var found int
		err = tx.QueryRowContext(ctx, `SELECT 1 FROM pos_locations WHERE business_id = ? AND id = ? LIMIT 1`, businessID, id).Scan(&found)
	case "flow_project":
		var found int
		err = tx.QueryRowContext(ctx, `SELECT 1 FROM flow_projects WHERE business_id = ? AND id = ? LIMIT 1`, businessID, id).Scan(&found)
	default:
		return domain.ErrInvalidScope
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrInvalidScope
	}
	return err
}

func insertAuditEvent(ctx context.Context, tx *sql.Tx, businessID, actorID int64, action, appID string, roleID, targetUserID, targetGroupID *int64, beforeJSON, afterJSON, reason string) error {
	var actor interface{}
	if actorID > 0 {
		actor = actorID
	}
	returnValue := actor
	_, err := tx.ExecContext(ctx, `
		INSERT INTO authorization_audit_events
			(business_id, actor_user_id, target_user_id, target_group_id, action,
			 app_id, role_id, before_json, after_json, reason, source_client)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'dash')`,
		businessID, returnValue, targetUserID, targetGroupID, action, appID,
		roleID, nullableString(beforeJSON), nullableString(afterJSON), nullableString(reason),
	)
	return err
}

func validAppID(appID string) bool {
	switch appID {
	case "dash", "pos", "erp", "hr", "crm", "finance", "flow", "insights", "menu", "recruit":
		return true
	default:
		return false
	}
}

func normalizedPage(page, pageSize int64) (int64, int64) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func nullableInt64(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}

func nullableString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func nonNilRoleStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func normalizedSchemaError(err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "authorization_") && (strings.Contains(message, "doesn't exist") || strings.Contains(message, "no such table")) {
		return domain.ErrAuthorizationSchemaUnavailable
	}
	return err
}
