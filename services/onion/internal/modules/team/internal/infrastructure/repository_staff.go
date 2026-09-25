package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	sqlnull "github.com/voxire/onion/internal/platform/utils/queries"
)

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

func (r *Repository) CountStaffUsersByBusinessID(ctx context.Context, businessID int64) (int64, error) {
	return r.queries.CountStaffUsersByBusinessID(ctx, sqlnull.Int64(businessID))
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
	// Atomic single-use claim: only an unused, unexpired token is marked, so
	// concurrent accepts cannot both succeed.
	res, err := r.queries.ConsumeInviteToken(ctx, token)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrTokenAlreadyUsed
	}
	return nil
}

// ReleaseInviteToken reverts the single-use claim taken by MarkInviteTokenUsed.
// AcceptInvite calls it when a later step fails, so the invitee can retry the
// accept instead of being left with a consumed token and no activation.
func (r *Repository) ReleaseInviteToken(ctx context.Context, token string) error {
	return r.queries.ReleaseInviteToken(ctx, token)
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
