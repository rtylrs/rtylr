package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/modules/organization/team/internal/domain"
	"github.com/voxire/onion/internal/modules/organization/team/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/authorization"
)

func (r *Repository) CreateGroup(ctx context.Context, actorID int64, input domain.CreateGroupInput) (int64, error) {
	if actorID <= 0 || input.BusinessID <= 0 {
		return 0, errors.New("actor and business are required")
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	rollback := func(cause error) (int64, error) { _ = tx.Rollback(); return 0, cause }
	qtx := r.queries.WithTx(tx)
	result, err := qtx.CreateGroup(ctx, sqlc.CreateGroupParams{
		BusinessID: input.BusinessID, Name: input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Color:       sql.NullString{String: input.Color, Valid: input.Color != ""},
	})
	if err != nil {
		return rollback(err)
	}
	groupID, err := result.LastInsertId()
	if err != nil {
		return rollback(err)
	}
	if err := r.replaceGroupAuthorization(ctx, tx, input.BusinessID, actorID, groupID, input.LicensedAppIDs, input.PermissionProfileAssignments); err != nil {
		return rollback(err)
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return groupID, nil
}

func (r *Repository) replaceGroupAuthorization(ctx context.Context, tx *sql.Tx, businessID, actorID, groupID int64, apps []string, assignments []domain.AppRoleAssignmentInput) error {
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_license_assignments
		SET status = 'revoked', revoked_by = ?, revoked_at = NOW()
		WHERE business_id = ? AND group_id = ? AND status = 'active'`, actorID, businessID, groupID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_role_assignments
		SET status = 'revoked', revoked_by = ?, revoked_at = NOW()
		WHERE business_id = ? AND group_id = ? AND status = 'active'`, actorID, businessID, groupID); err != nil {
		return err
	}
	for _, appID := range uniqueAuthorizationApps(apps) {
		if !isTenantAppID(appID) {
			return fmt.Errorf("invalid app id: %s", appID)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO authorization_license_assignments
				(business_id, group_id, app_id, source, status, granted_by)
			VALUES (?, ?, ?, 'group', 'active', ?)`, businessID, groupID, appID, actorID); err != nil {
			return err
		}
	}
	for _, assignment := range assignments {
		if assignment.AppID == "" || (assignment.AppID != "dash" && !isTenantAppID(assignment.AppID)) || len(assignment.RoleIDs) == 0 {
			return errors.New("permission profile assignment requires an app and at least one profile")
		}
		if assignment.GroupID != nil || assignment.Source == string(authorization.SourceSystem) {
			return errors.New("group assignments cannot set their own source or target")
		}
		for _, profileID := range uniqueInt64(assignment.RoleIDs) {
			if err := r.validateAuthorizationRoleID(ctx, tx, businessID, assignment.AppID, profileID); err != nil {
				return fmt.Errorf("permission profile %d is invalid for %s: %w", profileID, assignment.AppID, err)
			}
			result, err := tx.ExecContext(ctx, `
				INSERT INTO authorization_role_assignments
					(business_id, group_id, app_id, role_id, source, status, is_primary, granted_by)
				VALUES (?, ?, ?, ?, 'group', 'active', FALSE, ?)`,
				businessID, groupID, assignment.AppID, profileID, actorID)
			if err != nil {
				return err
			}
			assignmentID, err := result.LastInsertId()
			if err != nil {
				return err
			}
			for _, scope := range assignment.Scopes {
				if err := authorization.ValidateAssignmentScope(ctx, businessID, scope.Type, scope.ID); err != nil {
					return fmt.Errorf("invalid scope %q:%q: %w", scope.Type, scope.ID, err)
				}
				if _, err := tx.ExecContext(ctx, `
					INSERT INTO authorization_role_assignment_scopes (assignment_id, scope_type, scope_id)
					VALUES (?, ?, ?)`, assignmentID, scope.Type, scope.ID); err != nil {
					return err
				}
			}
		}
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO authorization_audit_events
			(business_id, actor_user_id, target_group_id, action, source_client)
		VALUES (?, ?, ?, 'group.authorization.updated', 'dash')`, businessID, actorID, groupID)
	return err
}

func (r *Repository) GetGroup(ctx context.Context, businessID, groupID int64) (*domain.Group, error) {
	row, err := r.queries.GetGroup(ctx, sqlc.GetGroupParams{ID: groupID, BusinessID: businessID})
	if err != nil {
		return nil, err
	}
	group := &domain.Group{
		ID: row.ID, BusinessID: row.BusinessID,
		Name: row.Name, Description: row.Description, Color: row.Color.String,
	}
	if row.CreatedAt.Valid {
		group.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		group.UpdatedAt = row.UpdatedAt.Time
	}
	group.LicensedAppIDs, group.PermissionProfileAssignments, err = r.groupAuthorization(ctx, businessID, groupID)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (r *Repository) groupAuthorization(ctx context.Context, businessID, groupID int64) ([]string, []authorization.RoleAssignment, error) {
	licenseRows, err := r.pool.QueryContext(ctx, `
		SELECT app_id FROM authorization_license_assignments
		WHERE business_id = ? AND group_id = ? AND status = 'active' ORDER BY app_id`, businessID, groupID)
	if err != nil {
		return nil, nil, err
	}
	licenses := make([]string, 0)
	for licenseRows.Next() {
		var appID string
		if err := licenseRows.Scan(&appID); err != nil {
			licenseRows.Close()
			return nil, nil, err
		}
		licenses = append(licenses, appID)
	}
	if err := licenseRows.Err(); err != nil {
		licenseRows.Close()
		return nil, nil, err
	}
	if err := licenseRows.Close(); err != nil {
		return nil, nil, err
	}

	rows, err := r.pool.QueryContext(ctx, `
		SELECT a.app_id, a.role_id, a.is_primary, s.scope_type, s.scope_id
		FROM authorization_role_assignments a
		LEFT JOIN authorization_role_assignment_scopes s ON s.assignment_id = a.id
		WHERE a.business_id = ? AND a.group_id = ? AND a.status = 'active'
		ORDER BY a.id, s.scope_type, s.scope_id`, businessID, groupID)
	if err != nil {
		return nil, nil, err
	}
	assignments := make([]authorization.RoleAssignment, 0)
	for rows.Next() {
		var appID string
		var profileID int64
		var primary bool
		var scopeType, scopeID sql.NullString
		if err := rows.Scan(&appID, &profileID, &primary, &scopeType, &scopeID); err != nil {
			rows.Close()
			return nil, nil, err
		}
		assignment := authorization.RoleAssignment{
			AppID: authorization.AppID(appID), RoleIDs: []int64{profileID},
			Source: authorization.SourceGroup, Active: true,
		}
		if primary {
			id := profileID
			assignment.PrimaryRoleID = &id
		}
		if scopeType.Valid && scopeID.Valid {
			assignment.Scopes = []authorization.Scope{{Type: scopeType.String, ID: scopeID.String}}
		}
		assignments = append(assignments, assignment)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, nil, err
	}
	return licenses, assignments, nil
}

func (r *Repository) ListGroups(ctx context.Context, businessID int64) ([]domain.Group, error) {
	rows, err := r.queries.ListGroups(ctx, businessID)
	if err != nil {
		return nil, err
	}
	groups := make([]domain.Group, 0, len(rows))
	for _, row := range rows {
		group := domain.Group{
			ID: row.ID, BusinessID: row.BusinessID,
			Name: row.Name, Description: row.Description, Color: row.Color.String, MemberCount: int(row.MemberCount),
		}
		if row.CreatedAt.Valid {
			group.CreatedAt = row.CreatedAt.Time
		}
		if row.UpdatedAt.Valid {
			group.UpdatedAt = row.UpdatedAt.Time
		}
		group.LicensedAppIDs, group.PermissionProfileAssignments, err = r.groupAuthorization(ctx, businessID, row.ID)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (r *Repository) UpdateGroup(ctx context.Context, actorID, businessID, groupID int64, input domain.UpdateGroupInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(cause error) error { _ = tx.Rollback(); return cause }
	qtx := r.queries.WithTx(tx)
	result, err := qtx.UpdateGroup(ctx, sqlc.UpdateGroupParams{
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Color:       sql.NullString{String: input.Color, Valid: input.Color != ""},
		ID:          groupID, BusinessID: businessID,
	})
	if err != nil {
		return rollback(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return rollback(err)
	}
	if affected == 0 {
		var found int
		if err := tx.QueryRowContext(ctx, `SELECT 1 FROM user_groups WHERE id = ? AND business_id = ?`, groupID, businessID).Scan(&found); errors.Is(err, sql.ErrNoRows) {
			return rollback(domain.ErrGroupNotFound)
		} else if err != nil {
			return rollback(err)
		}
	}
	if err := r.replaceGroupAuthorization(ctx, tx, businessID, actorID, groupID, input.LicensedAppIDs, input.PermissionProfileAssignments); err != nil {
		return rollback(err)
	}
	return tx.Commit()
}

func (r *Repository) DeleteGroup(ctx context.Context, businessID, groupID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(cause error) error { _ = tx.Rollback(); return cause }
	if _, err := tx.ExecContext(ctx, `
		DELETE scope FROM authorization_role_assignment_scopes scope
		JOIN authorization_role_assignments assignment ON assignment.id = scope.assignment_id
		WHERE assignment.business_id = ? AND assignment.group_id = ?`, businessID, groupID); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM authorization_role_assignments WHERE business_id = ? AND group_id = ?`, businessID, groupID); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM authorization_license_assignments WHERE business_id = ? AND group_id = ?`, businessID, groupID); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_group_members WHERE business_id = ? AND group_id = ?`, businessID, groupID); err != nil {
		return rollback(err)
	}
	result, err := r.queries.WithTx(tx).DeleteGroup(ctx, sqlc.DeleteGroupParams{ID: groupID, BusinessID: businessID})
	if err != nil {
		return rollback(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return rollback(err)
	}
	if affected == 0 {
		return rollback(domain.ErrGroupNotFound)
	}
	return tx.Commit()
}

func (r *Repository) AddGroupMembers(ctx context.Context, groupID, businessID, addedBy int64, userIDs []int64) error {
	for _, userID := range uniqueInt64(userIDs) {
		if err := r.queries.AddGroupMember(ctx, sqlc.AddGroupMemberParams{
			GroupID: groupID, UserID: userID, BusinessID: businessID,
			AddedBy:  sql.NullInt64{Int64: addedBy, Valid: addedBy > 0},
			UserID_2: userID, ID: groupID, BusinessID_2: businessID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) RemoveGroupMember(ctx context.Context, businessID, groupID, userID int64) error {
	return r.queries.RemoveGroupMember(ctx, sqlc.RemoveGroupMemberParams{GroupID: groupID, UserID: userID, BusinessID: businessID})
}

func (r *Repository) GetGroupDetail(ctx context.Context, businessID, groupID int64) (*domain.GroupDetail, error) {
	group, err := r.GetGroup(ctx, businessID, groupID)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListGroupMembers(ctx, sqlc.ListGroupMembersParams{GroupID: groupID, BusinessID: businessID})
	if err != nil {
		return nil, err
	}
	detail := &domain.GroupDetail{Group: *group, Members: make([]domain.GroupMember, 0, len(rows))}
	for _, row := range rows {
		member := domain.GroupMember{UserID: row.UserID, Email: row.Email, FullName: row.FullName}
		if row.AddedAt.Valid {
			member.AddedAt = row.AddedAt.Time
		}
		detail.Members = append(detail.Members, member)
	}
	return detail, nil
}

func (r *Repository) GetUserGroups(ctx context.Context, businessID, userID int64) ([]domain.GroupSummary, error) {
	rows, err := r.queries.ListUserGroups(ctx, sqlc.ListUserGroupsParams{UserID: userID, BusinessID: businessID})
	if err != nil {
		return nil, err
	}
	groups := make([]domain.GroupSummary, 0, len(rows))
	for _, row := range rows {
		groups = append(groups, domain.GroupSummary{ID: row.ID, Name: row.Name, Color: row.Color.String})
	}
	return groups, nil
}
