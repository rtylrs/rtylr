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
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := r.queries.WithTx(tx)
	if err := qtx.DeactivateUser(ctx, sqlc.DeactivateUserParams{OwnerID: ownerID, UserID: userID}); err != nil {
		logger.Ctx(ctx).Error("Failed to deactivate user", logger.Error(err))
		return err
	}
	newCount, err := qtx.CountActiveMembers(ctx, sqlc.CountActiveMembersParams{OwnerID: ownerID, OwnerID_2: ownerID, OwnerID_3: ownerID})
	if err != nil {
		return err
	}
	if err := r.recordSeatChangeTx(ctx, qtx, ownerID, int(newCount)); err != nil {
		return err
	}
	return tx.Commit()
}

// RevokeBusinessUserPermissions deactivates one user's permission rows in one
// business when hr_core marks that user's employee record terminated. The
// employment decision stays in hr_employees; this module only owns the write.
func (r *Repository) RevokeBusinessUserPermissions(ctx context.Context, businessID, userID int64) (int64, error) {
	rows, err := r.queries.RevokeBusinessUserPermissions(ctx, sqlc.RevokeBusinessUserPermissionsParams{
		BusinessID: nullBusinessID(businessID),
		UserID:     userID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to revoke permissions for terminated employee",
			logger.Any("business_id", businessID), logger.Any("user_id", userID), logger.Error(err))
		return 0, err
	}
	return rows, nil
}

// ReconcileTerminatedEmployeeAccess sweeps every business and deactivates any
// still-active permission row whose user is a terminated employee of that
// same business. Idempotent; the daily job runs it so a missed synchronous
// revoke is corrected within a day.
func (r *Repository) ReconcileTerminatedEmployeeAccess(ctx context.Context) (int64, error) {
	rows, err := r.queries.ReconcileTerminatedEmployeeAccess(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("terminated-employee access reconciliation failed", logger.Error(err))
		return 0, err
	}
	return rows, nil
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

// CreateInviteTransaction atomically provisions the user, permission, employee,
// invite token, and corresponding billing seat state. A failed employee or
// billing write rolls back the entire invite instead of leaving partial access.
func (r *Repository) CreateInviteTransaction(ctx context.Context, businessID, ownerID int64, email, passwordHash, fullName string, apps []string, role domain.Role, token string, expiresAt time.Time) (int64, error) {
	appsJSON, err := json.Marshal(apps)
	if err != nil {
		return 0, err
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := r.queries.WithTx(tx)
	// Lock the subscription before provisioning anything. This serializes
	// concurrent invitations for the same owner and makes the limit check
	// authoritative rather than a racy count-then-insert sequence.
	subscription, err := qtx.LockSubscriptionForSeatChange(ctx, ownerID)
	if err != nil {
		return 0, err
	}
	user, err := qtx.GetUserByEmail(ctx, email)
	var userID int64
	if err == nil {
		userID = user.ID
	} else if errors.Is(err, sql.ErrNoRows) {
		result, createErr := qtx.CreateUser(ctx, sqlc.CreateUserParams{
			Email:        email,
			PasswordHash: passwordHash,
			FullName:     sql.NullString{String: fullName, Valid: fullName != ""},
		})
		if createErr != nil {
			return 0, createErr
		}
		userID, err = result.LastInsertId()
		if err != nil {
			return 0, err
		}
	} else {
		return 0, err
	}

	if err := qtx.UpsertPermission(ctx, sqlc.UpsertPermissionParams{
		BusinessID: sql.NullInt64{Int64: businessID, Valid: businessID > 0},
		OwnerID:    ownerID,
		UserID:     userID,
		Apps:       appsJSON,
		Role:       sqlc.NullUserAppPermissionsRole{UserAppPermissionsRole: sqlc.UserAppPermissionsRole(role), Valid: true},
	}); err != nil {
		return 0, err
	}
	if err := qtx.EnsureEmployeeForUser(ctx, sqlc.EnsureEmployeeForUserParams{OwnerID: ownerID, ID: userID}); err != nil {
		return 0, err
	}
	if err := qtx.CreateInviteToken(ctx, sqlc.CreateInviteTokenParams{UserID: userID, OwnerID: ownerID, Token: token, ExpiresAt: expiresAt}); err != nil {
		return 0, err
	}

	newCount, err := qtx.CountActiveMembers(ctx, sqlc.CountActiveMembersParams{OwnerID: ownerID, OwnerID_2: ownerID, OwnerID_3: ownerID})
	if err != nil {
		return 0, err
	}
	if subscription.MaxSeats.Valid && int32(newCount) > subscription.MaxSeats.Int32 {
		return 0, domain.ErrSeatLimitExceeded
	}
	if err := r.recordSeatChangeTx(ctx, qtx, ownerID, int(newCount)); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *Repository) EnsureEmployeeForUser(ctx context.Context, ownerID, userID int64) error {
	return r.queries.EnsureEmployeeForUser(ctx, sqlc.EnsureEmployeeForUserParams{
		OwnerID: ownerID,
		ID:      userID,
	})
}

func (r *Repository) UpdateUserProfile(ctx context.Context, userID int64, fullName string) error {
	err := r.queries.UpdateUserProfile(ctx, sqlc.UpdateUserProfileParams{
		FullName: sql.NullString{String: strings.TrimSpace(fullName), Valid: strings.TrimSpace(fullName) != ""},
		ID:       userID,
	})
	if err != nil {
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

	rows, err := r.queries.ListUserGroupsForUsers(ctx, sqlc.ListUserGroupsForUsersParams{
		UserIds: userIDs,
		OwnerID: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to batch list user groups", logger.Error(err))
		return nil, err
	}

	for _, row := range rows {
		summary := domain.GroupSummary{ID: row.ID, Name: row.Name, Color: row.Color.String}
		result[row.UserID] = append(result[row.UserID], summary)
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

// SetPasswordAndConsumeInviteToken makes invite acceptance one transaction.
// The conditional token update wins the race between concurrent accept calls;
// only the caller that consumes the unused, unexpired token may set a password.
func (r *Repository) SetPasswordAndConsumeInviteToken(ctx context.Context, userID int64, token, passwordHash string) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := r.queries.WithTx(tx)
	result, err := qtx.ConsumeInviteToken(ctx, token)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		row, lookupErr := qtx.GetInviteToken(ctx, token)
		if lookupErr != nil {
			return lookupErr
		}
		if row.UsedAt.Valid {
			return domain.ErrTokenAlreadyUsed
		}
		if time.Now().After(row.ExpiresAt) {
			return domain.ErrTokenExpired
		}
		return domain.ErrTokenAlreadyUsed
	}

	if err := qtx.SetPassword(ctx, sqlc.SetPasswordParams{PasswordHash: passwordHash, ID: userID}); err != nil {
		return err
	}
	return tx.Commit()
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
	return r.queries.GetBusinessIDByOwnerID(ctx, ownerID)
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
	count, err := r.queries.CountActiveMembers(ctx, sqlc.CountActiveMembersParams{
		OwnerID:   ownerID,
		OwnerID_2: ownerID,
		OwnerID_3: ownerID,
	})
	if err != nil {
		return 0, err
	}
	if count < 1 {
		return 1, nil
	}
	return int(count), nil
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

func (r *Repository) UpdateGroup(ctx context.Context, ownerID, groupID int64, input domain.UpdateGroupInput) error {
	res, err := r.queries.UpdateGroup(ctx, sqlc.UpdateGroupParams{
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Color:       sql.NullString{String: input.Color, Valid: input.Color != ""},
		ID:          groupID,
		OwnerID:     ownerID,
	})
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	// A foreign or stale group id must never fall through to the unscoped
	// resource-permission rewrite below.
	if affected == 0 {
		return domain.ErrGroupNotFound
	}
	return r.setGroupResourcePermissions(ctx, groupID, input.ResourcePermissions)
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

// GetOwnerSubscriptionInfo returns the plan slug and status for the owner's subscription.
func (r *Repository) GetOwnerSubscriptionInfo(ctx context.Context, ownerID int64) (string, string, error) {
	// cross-schema read from subscriptions + plans — approved for owner subscription info
	row, err := r.queries.GetOwnerSubscriptionInfo(ctx, ownerID)
	if err != nil {
		return "", "", err
	}
	status := ""
	if row.Status.Valid {
		status = string(row.Status.SubscriptionsStatus)
	}
	return row.Slug, status, nil
}

// RecordSeatChange updates the billing seat_periods table after a team member is added or removed.
// All three writes (close the active period, open a new one, update the subscription seat count)
// run inside a single transaction so billing state can never end up partially applied.
func (r *Repository) RecordSeatChange(ctx context.Context, ownerID int64, newCount int) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := r.recordSeatChangeTx(ctx, r.queries.WithTx(tx), ownerID, newCount); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) recordSeatChangeTx(ctx context.Context, qtx *sqlc.Queries, ownerID int64, newCount int) error {
	now := time.Now().UTC()
	sub, err := qtx.LockSubscriptionForSeatChange(ctx, ownerID)
	if err != nil {
		return err
	}
	if err := qtx.CloseActiveSeatPeriod(ctx, sqlc.CloseActiveSeatPeriodParams{
		PeriodEnd:      sql.NullTime{Time: now.AddDate(0, 0, -1), Valid: true},
		SubscriptionID: sub.ID,
	}); err != nil {
		return err
	}
	if err := qtx.OpenSeatPeriod(ctx, sqlc.OpenSeatPeriodParams{
		SubscriptionID:    sub.ID,
		OwnerID:           ownerID,
		SeatCount:         int32(newCount),
		PricePerSeatCents: sub.PricePerSeatCents,
		PeriodStart:       now,
		BillingMonth:      time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		return err
	}
	return qtx.UpdateSubscriptionSeatCount(ctx, sqlc.UpdateSubscriptionSeatCountParams{
		SeatCount: int32(newCount),
		OwnerID:   ownerID,
	})
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
	return r.queries.ListAdminsAndOwnerUserIDs(ctx, sqlc.ListAdminsAndOwnerUserIDsParams{
		ID:         businessID,
		BusinessID: sql.NullInt64{Int64: businessID, Valid: true},
	})
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
	if len(userIDs) == 0 {
		return nil
	}

	// cross-schema write to hub_notifications — approved for cross-module access request alerts
	insert := sq.Insert("hub_notifications").Columns("owner_id", "user_id", "title", "body", "type", "link")
	for _, uid := range userIDs {
		insert = insert.Values(
			ownerID,
			uid,
			title,
			sql.NullString{String: body, Valid: body != ""},
			notifType,
			sql.NullString{String: link, Valid: link != ""},
		)
	}
	query, args, err := insert.ToSql()
	if err != nil {
		return err
	}
	if _, err := r.pool.ExecContext(ctx, query, args...); err != nil {
		logger.Ctx(ctx).Warn("failed to create hub notifications",
			logger.Int("recipient_count", len(userIDs)), logger.Error(err))
		return err
	}
	return nil
}

// nullBusinessID wraps the join-side business id so sqlc's named parameter
// (shared by both placeholders in RevokeBusinessUserPermissions) scans as
// NOT NULL int64.
func nullBusinessID(v int64) sql.NullInt64 { return sql.NullInt64{Int64: v, Valid: true} }
