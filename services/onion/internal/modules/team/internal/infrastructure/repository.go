package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	sqlnull "github.com/voxire/onion/internal/platform/utils/queries"
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

func (r *Repository) DeactivateUser(ctx context.Context, ownerID, userID int64) error {
	err := r.queries.DeactivateUser(ctx, sqlc.DeactivateUserParams{
		OwnerID: ownerID,
		UserID:  userID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to deactivate user", logger.Error(err))

		return err
	}

	return nil
}

// ─── Users ────────────────────────────────────────────────────────────────────

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (int64, string, string, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return 0, "", "", err
	}

	return row.ID, row.Email, row.FullName, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (string, string, bool, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", false, domain.ErrUserNotFound
		}

		return "", "", false, err
	}

	return row.Email, row.FullName, row.IsConsoleAdmin, nil
}

func (r *Repository) GetUserBusinessID(ctx context.Context, id int64) (int64, error) {
	businessID, err := r.queries.GetUserBusinessID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrUserNotFound
		}
		return 0, err
	}
	if !businessID.Valid {
		return 0, nil
	}
	return businessID.Int64, nil
}

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, fullName string) (int64, error) {
	res, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     sql.NullString{String: fullName, Valid: fullName != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create user via hub", logger.Error(err))

		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) EnsureEmployeeForUser(ctx context.Context, ownerID, userID int64) error {
	return r.queries.EnsureEmployeeForUser(ctx, sqlc.EnsureEmployeeForUserParams{
		OwnerID: ownerID,
		ID:      userID,
	})
}

func (r *Repository) UpdateUserProfile(ctx context.Context, userID int64, fullName string) error {
	query, args, err := sq.
		Update("users").
		Set("full_name", strings.TrimSpace(fullName)).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": userID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	if _, err := r.pool.ExecContext(ctx, query, args...); err != nil {
		logger.Ctx(ctx).Error("Failed to update user profile via team", logger.Error(err))
		return err
	}

	return nil
}

// ─── Staff listing (join users + permissions) ─────────────────────────────────

func (r *Repository) ListStaffUsers(ctx context.Context, ownerID int64, limit, offset int32) ([]domain.StaffUser, error) {
	rows, err := r.queries.ListStaffUsers(ctx, sqlc.ListStaffUsersParams{
		OwnerID: ownerID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list staff users", logger.Error(err))

		return nil, err
	}

	users := make([]domain.StaffUser, 0, len(rows))

	for _, row := range rows {
		var su domain.StaffUser
		su.ID = row.ID
		su.Email = row.Email
		su.FullName = row.FullName
		if row.CreatedAt.Valid {
			su.CreatedAt = row.CreatedAt.Time
		}
		su.Role = domain.Role(row.Role.UserAppPermissionsRole)
		if row.IsActive.Valid {
			su.IsActive = row.IsActive.Bool
		}

		if err := json.Unmarshal(row.Apps, &su.Apps); err != nil {
			su.Apps = []string{}
		}

		users = append(users, su)
	}

	userIDs := make([]int64, len(users))
	for i := range users {
		userIDs[i] = users[i].ID
		users[i].Groups = []domain.GroupSummary{}
	}

	groupsByUser, err := r.groupsForUsers(ctx, ownerID, userIDs)
	if err == nil {
		for i := range users {
			if groups := groupsByUser[users[i].ID]; groups != nil {
				users[i].Groups = groups
			}
		}
	}

	return users, nil
}

// ListStaffUsersByBusinessID lists team members scoped to a single business
// (via user_app_permissions.business_id), so each business sees only its own
// team. Group enrichment stays owner-scoped (groups are owner-level), resolved
// from the business's owner.
func (r *Repository) ListStaffUsersByBusinessID(ctx context.Context, businessID int64, limit, offset int32) ([]domain.StaffUser, error) {
	rows, err := r.queries.ListStaffUsersByBusinessID(ctx, sqlc.ListStaffUsersByBusinessIDParams{
		BusinessID: sqlnull.Int64(businessID),
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list staff users by business", logger.Error(err))

		return nil, err
	}

	users := make([]domain.StaffUser, 0, len(rows))

	for _, row := range rows {
		var su domain.StaffUser
		su.ID = row.ID
		su.Email = row.Email
		su.FullName = row.FullName
		if row.CreatedAt.Valid {
			su.CreatedAt = row.CreatedAt.Time
		}
		su.Role = domain.Role(row.Role.UserAppPermissionsRole)
		if row.IsActive.Valid {
			su.IsActive = row.IsActive.Bool
		}

		if err := json.Unmarshal(row.Apps, &su.Apps); err != nil {
			su.Apps = []string{}
		}

		users = append(users, su)
	}

	userIDs := make([]int64, len(users))
	for i := range users {
		userIDs[i] = users[i].ID
		users[i].Groups = []domain.GroupSummary{}
	}

	if _, ownerID, ownerErr := r.GetBusinessByID(ctx, businessID); ownerErr == nil {
		if groupsByUser, gErr := r.groupsForUsers(ctx, ownerID, userIDs); gErr == nil {
			for i := range users {
				if groups := groupsByUser[users[i].ID]; groups != nil {
					users[i].Groups = groups
				}
			}
		}
	}

	return users, nil
}

// groupsForUsers batches group lookups for many users into a single query,
// mapping the result back per user ID in memory (avoids an N+1 over users).
func (r *Repository) groupsForUsers(ctx context.Context, ownerID int64, userIDs []int64) (map[int64][]domain.GroupSummary, error) {
	result := make(map[int64][]domain.GroupSummary, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	query, args, err := sq.
		Select("m.user_id", "g.id", "g.name", "g.color").
		From("user_group_members m").
		Join("user_groups g ON g.id = m.group_id").
		Where(sq.Eq{"m.user_id": userIDs}).
		Where(sq.Eq{"m.owner_id": ownerID}).
		OrderBy("g.name").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build batch list user groups query", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to batch list user groups", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID int64
		var color sql.NullString
		var summary domain.GroupSummary
		if err := rows.Scan(&userID, &summary.ID, &summary.Name, &color); err != nil {
			return nil, err
		}
		summary.Color = color.String
		result[userID] = append(result[userID], summary)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) GetStaffUserByID(ctx context.Context, ownerID, businessID, userID int64) (*domain.StaffUser, error) {
	row, err := r.queries.GetStaffUserByID(ctx, sqlc.GetStaffUserByIDParams{
		OwnerID:    ownerID,
		BusinessID: sql.NullInt64{Int64: businessID, Valid: true},
		UserID:     userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		logger.Ctx(ctx).Error("Failed to get staff user by id", logger.Error(err))

		return nil, err
	}

	su := &domain.StaffUser{
		ID:       row.ID,
		Email:    row.Email,
		FullName: row.FullName,
		Role:     domain.Role(row.Role.UserAppPermissionsRole),
	}

	if row.CreatedAt.Valid {
		su.CreatedAt = row.CreatedAt.Time
	}

	if row.IsActive.Valid {
		su.IsActive = row.IsActive.Bool
	}

	if err := json.Unmarshal(row.Apps, &su.Apps); err != nil {
		su.Apps = []string{}
	}

	groups, err := r.GetUserGroups(ctx, ownerID, su.ID)
	if err != nil || groups == nil {
		groups = []domain.GroupSummary{}
	}
	su.Groups = groups

	return su, nil
}

// ─── Invite Tokens ──────────────────────────────────────────────────────────

func (r *Repository) CreateInviteToken(ctx context.Context, userID, ownerID int64, token string, expiresAt time.Time) error {
	err := r.queries.CreateInviteToken(ctx, sqlc.CreateInviteTokenParams{
		UserID:    userID,
		OwnerID:   ownerID,
		Token:     token,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create invite token", logger.Error(err))
	}
	return err
}

func (r *Repository) GetInviteToken(ctx context.Context, token string) (int64, int64, time.Time, error) {
	row, err := r.queries.GetInviteToken(ctx, token)
	if err != nil {
		return 0, 0, time.Time{}, err
	}

	if row.UsedAt.Valid {
		return 0, 0, time.Time{}, domain.ErrTokenAlreadyUsed
	}

	if time.Now().After(row.ExpiresAt) {
		return 0, 0, time.Time{}, domain.ErrTokenExpired
	}

	return row.UserID, row.OwnerID, row.ExpiresAt, nil
}

func (r *Repository) MarkInviteTokenUsed(ctx context.Context, token string) error {
	return r.queries.MarkInviteTokenUsed(ctx, token)
}

func (r *Repository) SetPassword(ctx context.Context, userID int64, passwordHash string) error {
	return r.queries.SetPassword(ctx, sqlc.SetPasswordParams{
		PasswordHash: passwordHash,
		ID:           userID,
	})
}

// GetBusinessByID returns the business name and owner ID.
func (r *Repository) GetBusinessByID(ctx context.Context, businessID int64) (string, int64, error) {
	row, err := r.queries.GetBusinessByID(ctx, businessID)
	if err != nil {
		return "", 0, err
	}
	return row.Name, row.OwnerID, nil
}

func (r *Repository) GetBusinessIDByOwnerID(ctx context.Context, ownerID int64) (int64, error) {
	q, args, err := sq.
		Select("id").
		From("businesses").
		Where(sq.Eq{"owner_id": ownerID}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}

	var businessID int64
	if err := r.pool.QueryRowContext(ctx, q, args...).Scan(&businessID); err != nil {
		return 0, err
	}
	return businessID, nil
}

// ─── Access Requests ────────────────────────────────────────────────────────

func (r *Repository) CreateAccessRequest(ctx context.Context, businessID, userID int64, app string) error {
	return r.queries.CreateAccessRequest(ctx, sqlc.CreateAccessRequestParams{
		BusinessID: businessID,
		UserID:     userID,
		App:        app,
	})
}

// ─── Onboarding ─────────────────────────────────────────────────────────────
//
// The `users` table (schemas/shared/users.sql) is a GLOBAL identity table keyed
// by `id` — it has no `owner_id` column (tenant association lives in the nullable
// `business_id`). Onboarding state belongs to the user identity, not to a tenant,
// so these queries intentionally key on `id` only and do NOT filter by owner_id.
// If `users` is ever re-scoped to a tenant during a schema refactor, revisit these
// two queries (and add an owner_id filter) before that migration ships.

func (r *Repository) IsOnboarded(ctx context.Context, userID int64) bool {
	onboarded, err := r.queries.IsOnboarded(ctx, userID)
	if err != nil {
		// A missing row (sql.ErrNoRows) legitimately means "not onboarded"; any
		// other error indicates a DB problem and must not be swallowed silently.
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Warn("Failed to read onboarding state, defaulting to not onboarded",
				logger.Any("user_id", userID), logger.Error(err))
		}
		return false
	}
	return onboarded
}

func (r *Repository) MarkOnboarded(ctx context.Context, userID int64) error {
	return r.queries.MarkOnboarded(ctx, userID)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func permissionFromRow(id, ownerID, userID int64, apps json.RawMessage, role sqlc.NullUserAppPermissionsRole, isActive sql.NullBool, createdAt, updatedAt sql.NullTime) *domain.UserAppPermission {
	p := &domain.UserAppPermission{
		ID:      id,
		OwnerID: ownerID,
		UserID:  userID,
	}

	if role.Valid {
		p.Role = domain.Role(role.UserAppPermissionsRole)
	}

	if isActive.Valid {
		p.IsActive = isActive.Bool
	}

	if createdAt.Valid {
		p.CreatedAt = createdAt.Time
	}

	if updatedAt.Valid {
		p.UpdatedAt = updatedAt.Time
	}

	if err := json.Unmarshal(apps, &p.Apps); err != nil {
		p.Apps = []string{}
	}

	return p
}

// CountActiveMembers returns the billable active seat count for an owner.
func (r *Repository) CountActiveMembers(ctx context.Context, ownerID int64) (int, error) {
	const query = `
SELECT COUNT(DISTINCT user_id)
FROM (
	SELECT b.owner_id AS user_id
	FROM businesses b
	JOIN users u ON u.id = b.owner_id
	WHERE b.owner_id = ? AND COALESCE(b.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE
	UNION
	SELECT p.user_id AS user_id
	FROM user_app_permissions p
	JOIN users u ON u.id = p.user_id
	WHERE p.owner_id = ? AND COALESCE(p.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE
	UNION
	SELECT u.id AS user_id
	FROM users u
	JOIN businesses b ON b.id = u.business_id
	WHERE b.owner_id = ? AND COALESCE(b.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE
) active_users`

	var count int
	if err := r.pool.QueryRowContext(ctx, query, ownerID, ownerID, ownerID).Scan(&count); err != nil {
		return 0, err
	}
	if count < 1 {
		return 1, nil
	}
	return count, nil
}

// ─── Groups ──────────────────────────────────────────────────────────────────

func (r *Repository) CreateGroup(ctx context.Context, ownerID int64, input domain.CreateGroupInput) (int64, error) {
	res, err := r.queries.CreateGroup(ctx, sqlc.CreateGroupParams{
		OwnerID:     ownerID,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Color:       sql.NullString{String: input.Color, Valid: input.Color != ""},
	})
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := r.setGroupResourcePermissions(ctx, id, input.ResourcePermissions); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) setGroupResourcePermissions(ctx context.Context, groupID int64, perms []domain.ResourcePermission) error {
	if err := r.queries.DeleteGroupResourcePermissions(ctx, groupID); err != nil {
		return err
	}
	for _, p := range perms {
		if err := r.queries.UpsertGroupResourcePermission(ctx, sqlc.UpsertGroupResourcePermissionParams{
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

	query, args, err := sq.
		Select("group_id", "resource_key", "permission_level").
		From("user_group_resource_permissions").
		Where(sq.Eq{"group_id": groupIDs}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build batch list group resource permissions query", logger.Error(err))
		return result
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to batch list group resource permissions", logger.Error(err))
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var groupID int64
		var p domain.ResourcePermission
		var level string
		if err := rows.Scan(&groupID, &p.ResourceKey, &level); err != nil {
			logger.Ctx(ctx).Error("Failed to scan group resource permission", logger.Error(err))
			return result
		}
		p.PermissionLevel = domain.PermissionLevel(level)
		result[groupID] = append(result[groupID], p)
	}
	if err := rows.Err(); err != nil {
		logger.Ctx(ctx).Error("Failed to iterate group resource permissions", logger.Error(err))
	}
	return result
}

func (r *Repository) UpdateGroup(ctx context.Context, ownerID, groupID int64, input domain.UpdateGroupInput) error {
	if err := r.queries.UpdateGroup(ctx, sqlc.UpdateGroupParams{
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Color:       sql.NullString{String: input.Color, Valid: input.Color != ""},
		ID:          groupID,
		OwnerID:     ownerID,
	}); err != nil {
		return err
	}
	return r.setGroupResourcePermissions(ctx, groupID, input.ResourcePermissions)
}

func (r *Repository) DeleteGroup(ctx context.Context, ownerID, groupID int64) error {
	if err := r.queries.DeleteGroupResourcePermissions(ctx, groupID); err != nil {
		return err
	}
	return r.queries.DeleteGroup(ctx, sqlc.DeleteGroupParams{ID: groupID, OwnerID: ownerID})
}

func (r *Repository) AddGroupMembers(ctx context.Context, ownerID, groupID, addedBy int64, userIDs []int64) error {
	for _, uid := range userIDs {
		if err := r.queries.AddGroupMember(ctx, sqlc.AddGroupMemberParams{
			GroupID: groupID,
			UserID:  uid,
			OwnerID: ownerID,
			AddedBy: sql.NullInt64{Int64: addedBy, Valid: addedBy > 0},
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
	var appsJSON string
	// cross-schema read from subscriptions + plans — approved for owner app resolution
	query, args, err := sq.
		Select("p.apps").
		From("subscriptions s").
		Join("plans p ON p.id = s.plan_id").
		Where(sq.Eq{"s.owner_id": ownerID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&appsJSON)
	if err != nil {
		return nil, err
	}

	var apps []string
	if jsonErr := json.Unmarshal([]byte(appsJSON), &apps); jsonErr != nil {
		return []string{}, nil
	}

	return apps, nil
}

// GetOwnerSubscriptionInfo returns the plan slug and status for the owner's subscription.
func (r *Repository) GetOwnerSubscriptionInfo(ctx context.Context, ownerID int64) (string, string, error) {
	var planSlug, status string
	// cross-schema read from subscriptions + plans — approved for owner subscription info
	query, args, err := sq.
		Select("p.slug", "s.status").
		From("subscriptions s").
		Join("plans p ON p.id = s.plan_id").
		Where(sq.Eq{"s.owner_id": ownerID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return "", "", err
	}

	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&planSlug, &status)
	return planSlug, status, err
}

// RecordSeatChange updates the billing seat_periods table after a team member is added or removed.
// All three writes (close the active period, open a new one, update the subscription seat count)
// run inside a single transaction so billing state can never end up partially applied.
func (r *Repository) RecordSeatChange(ctx context.Context, ownerID int64, newCount int) error {
	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1)
	billingMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	var subscriptionID int64
	var pricePerSeatCents int
	// cross-schema read from subscriptions — approved for seat period management
	selectQuery, selectArgs, err := sq.
		Select("id", "price_per_seat_cents").
		From("subscriptions").
		Where(sq.Eq{"owner_id": ownerID}).
		Suffix("FOR UPDATE").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.QueryRowContext(ctx, selectQuery, selectArgs...).
		Scan(&subscriptionID, &pricePerSeatCents); err != nil {
		_ = tx.Rollback()
		return err
	}

	// cross-schema write to seat_periods — approved for seat billing accuracy
	closeQuery, closeArgs, err := sq.
		Update("seat_periods").
		Set("period_end", yesterday).
		Where(sq.Eq{"subscription_id": subscriptionID}).
		Where(sq.Eq{"period_end": nil}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, execErr := tx.ExecContext(ctx, closeQuery, closeArgs...); execErr != nil {
		_ = tx.Rollback()
		return execErr
	}

	// cross-schema write to seat_periods — approved for seat billing accuracy
	insertQuery, insertArgs, err := sq.
		Insert("seat_periods").
		Columns("subscription_id", "owner_id", "seat_count", "price_per_seat_cents", "period_start", "billing_month").
		Values(subscriptionID, ownerID, newCount, pricePerSeatCents, now, billingMonth).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, execErr := tx.ExecContext(ctx, insertQuery, insertArgs...); execErr != nil {
		_ = tx.Rollback()
		return execErr
	}

	// cross-schema write to subscriptions — approved for seat billing accuracy
	updateQuery, updateArgs, err := sq.
		Update("subscriptions").
		Set("seat_count", newCount).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"owner_id": ownerID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, execErr := tx.ExecContext(ctx, updateQuery, updateArgs...); execErr != nil {
		_ = tx.Rollback()
		return execErr
	}

	return tx.Commit()
}

// ─── Access Requests ──────────────────────────────────────────────────────────

func (r *Repository) ListAccessRequests(ctx context.Context, businessID int64) ([]domain.AccessRequest, error) {
	rows, err := r.queries.ListAccessRequestsByBusiness(ctx, businessID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.AccessRequest, 0, len(rows))
	for _, row := range rows {
		status := "pending"
		if row.Status.Valid {
			status = string(row.Status.AccessRequestsStatus)
		}
		out = append(out, domain.AccessRequest{
			ID:             row.ID,
			UserID:         row.UserID,
			App:            row.App,
			Status:         status,
			RequesterName:  row.RequesterName,
			RequesterEmail: row.RequesterEmail,
			CreatedAt:      row.CreatedAt.Time,
		})
	}
	return out, nil
}

// GetAdminsAndOwnerUserIDs returns all user IDs that should receive management
// notifications for a business: the primary owner plus any staff with
// role in ('admin','owner') in user_app_permissions.
func (r *Repository) GetAdminsAndOwnerUserIDs(ctx context.Context, businessID int64) ([]int64, error) {
	// cross-schema read from hub — approved for business-wide admin notification fan-out
	sqlRows, err := r.pool.QueryContext(ctx, `
		SELECT b.owner_id
		FROM businesses b
		WHERE b.id = ?
		UNION
		SELECT p.user_id
		FROM user_app_permissions p
		WHERE p.business_id = ?
		  AND p.role IN ('admin', 'owner')
		  AND p.is_active = TRUE
	`, businessID, businessID)
	if err != nil {
		return nil, err
	}
	defer sqlRows.Close()
	var ids []int64
	for sqlRows.Next() {
		var id int64
		if err := sqlRows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, sqlRows.Err()
}

// BulkCreateHubNotification writes an in-app notification to hub_notifications for
// each recipient in userIDs. Best-effort: errors are logged, not propagated.
func (r *Repository) ResolveAccessRequest(ctx context.Context, id, businessID int64, status string) error {
	return r.queries.ResolveAccessRequest(ctx, sqlc.ResolveAccessRequestParams{
		ID:         id,
		BusinessID: businessID,
		Status: sqlc.NullAccessRequestsStatus{
			AccessRequestsStatus: sqlc.AccessRequestsStatus(status),
			Valid:                true,
		},
	})
}

func (r *Repository) BulkCreateHubNotification(ctx context.Context, ownerID int64, userIDs []int64, title, body, notifType, link string) error {
	for _, uid := range userIDs {
		// cross-schema write to hub_notifications — approved for cross-module access request alerts
		if err := r.queries.CreateHubNotification(ctx, sqlc.CreateHubNotificationParams{
			OwnerID: ownerID,
			UserID:  uid,
			Title:   title,
			Body:    sql.NullString{String: body, Valid: body != ""},
			Type:    notifType,
			Link:    sql.NullString{String: link, Valid: link != ""},
		}); err != nil {
			logger.Ctx(ctx).Warn("failed to create hub notification",
				logger.Int("user_id", int(uid)), logger.Error(err))
		}
	}
	return nil
}
