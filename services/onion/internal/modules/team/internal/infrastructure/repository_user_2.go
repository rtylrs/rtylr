package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/authorization"
)

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
