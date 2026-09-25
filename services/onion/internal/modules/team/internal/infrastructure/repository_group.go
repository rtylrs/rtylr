package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

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
