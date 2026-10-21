package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"

	"github.com/voxire/onion/internal/modules/organization/roles/internal/domain"
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
	afterJSON, err := marshalCreateRoleInputAudit(input)
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
	beforeJSON, err := marshalRoleAudit(role)
	if err != nil {
		return rollback(err)
	}
	afterJSON, err := marshalUpdateRoleInputAudit(input)
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
