package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/roles/internal/domain"
)

func validateRolePermissions(ctx context.Context, executor interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}, appID string, permissions []string,
) error {
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
