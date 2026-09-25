package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/voxire/onion/internal/modules/roles/internal/domain"
)

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
