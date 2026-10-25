package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/organization/team/internal/domain"
	"github.com/voxire/onion/internal/modules/organization/team/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
)

// ListStaffUsersByBusinessID returns members from the current business.
func (r *Repository) ListStaffUsersByBusinessID(ctx context.Context, businessID int64, limit, offset int32) ([]domain.StaffUser, error) {
	rows, err := r.queries.ListStaffUsersByBusinessID(ctx, sqlc.ListStaffUsersByBusinessIDParams{
		BusinessID: businessID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	users := make([]domain.StaffUser, 0, len(rows))
	for _, row := range rows {
		users = append(users, domain.StaffUser{
			ID: row.ID, Email: row.Email, FullName: row.FullName,
			CreatedAt: row.CreatedAt.Time, MemberType: string(row.MemberType),
			IsActive: string(row.MemberStatus) == "active",
		})
	}
	userIDs := make([]int64, 0, len(users))
	for i := range users {
		userIDs = append(userIDs, users[i].ID)
		users[i].Groups = []domain.GroupSummary{}
	}
	groupsByUser, err := r.groupsForUsers(ctx, businessID, userIDs)
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i].Groups = groupsByUser[users[i].ID]
	}
	businessApps, err := r.GetBusinessApps(ctx, businessID)
	if err != nil {
		return nil, err
	}
	for i := range users {
		snapshot, found, err := r.GetAuthorizationSnapshot(ctx, businessID, users[i].ID, businessApps)
		if err != nil {
			return nil, err
		}
		if found {
			applyStaffAuthorization(&users[i], snapshot)
		}
	}
	return users, nil
}

func (r *Repository) CountStaffUsersByBusinessID(ctx context.Context, businessID int64) (int64, error) {
	return r.queries.CountStaffUsersByBusinessID(ctx, businessID)
}

// groupsForUsers batches group lookups for many users into one query.
func (r *Repository) groupsForUsers(ctx context.Context, businessID int64, userIDs []int64) (map[int64][]domain.GroupSummary, error) {
	result := make(map[int64][]domain.GroupSummary, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	rows, err := r.queries.ListUserGroupsForUsers(ctx, sqlc.ListUserGroupsForUsersParams{UserIds: userIDs, BusinessID: businessID})
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.UserID] = append(result[row.UserID], domain.GroupSummary{ID: row.ID, Name: row.Name, Color: row.Color.String})
	}
	return result, nil
}

func (r *Repository) GetStaffUserByID(ctx context.Context, businessID, userID int64) (*domain.StaffUser, error) {
	row, err := r.queries.GetStaffUserByID(ctx, sqlc.GetStaffUserByIDParams{BusinessID: businessID, UserID: userID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	member := &domain.StaffUser{
		ID: row.ID, Email: row.Email, FullName: row.FullName,
		MemberType: string(row.MemberType), IsActive: string(row.MemberStatus) == "active",
	}
	if row.CreatedAt.Valid {
		member.CreatedAt = row.CreatedAt.Time
	}
	businessApps, err := r.GetBusinessApps(ctx, businessID)
	if err != nil {
		return nil, err
	}
	snapshot, found, err := r.GetAuthorizationSnapshot(ctx, businessID, userID, businessApps)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, domain.ErrUserNotFound
	}
	applyStaffAuthorization(member, snapshot)
	member.Groups, err = r.GetUserGroups(ctx, businessID, userID)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func applyStaffAuthorization(member *domain.StaffUser, snapshot authorization.Snapshot) {
	resolved := authorization.Resolve(snapshot)
	member.EffectiveAccess = &resolved
	member.PermissionProfileAssignments = make([]authorization.RoleAssignment, 0, len(snapshot.Assignments))
	for _, assignment := range snapshot.Assignments {
		if assignment.Active && assignment.GroupID == nil && assignment.Source == authorization.SourceExplicit {
			member.PermissionProfileAssignments = append(member.PermissionProfileAssignments, assignment)
		}
	}
	member.LicensedAppIDs = make([]string, 0, len(snapshot.Licenses))
	for _, license := range snapshot.Licenses {
		if license.Active && license.GroupID == nil && license.Source == authorization.SourceExplicit {
			member.LicensedAppIDs = appendUniqueAuthorizationApp(member.LicensedAppIDs, string(license.AppID))
		}
	}
}

// ─── Invite Tokens ──────────────────────────────────────────────────────────

// CreateInviteToken records the business the invitation grants access in, so
// the accept path resolves the tenant from the invitation itself.
func (r *Repository) CreateInviteToken(ctx context.Context, userID, businessID int64, token string, expiresAt time.Time) error {
	err := r.queries.CreateInviteToken(ctx, sqlc.CreateInviteTokenParams{
		UserID:     userID,
		BusinessID: businessID,
		Token:      token,
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create invite token", logger.Error(err))
	}
	return err
}

func (r *Repository) GetInviteToken(ctx context.Context, token string) (domain.InviteToken, error) {
	row, err := r.queries.GetInviteToken(ctx, token)
	if err != nil {
		return domain.InviteToken{}, err
	}

	if row.UsedAt.Valid {
		return domain.InviteToken{}, domain.ErrTokenAlreadyUsed
	}

	if time.Now().After(row.ExpiresAt) {
		return domain.InviteToken{}, domain.ErrTokenExpired
	}

	return domain.InviteToken{
		UserID:     row.UserID,
		BusinessID: row.BusinessID,
		ExpiresAt:  row.ExpiresAt,
	}, nil
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

// CountActiveMembers returns the billable active seat count for a business.
func (r *Repository) CountActiveMembers(ctx context.Context, businessID int64) (int, error) {
	count, err := r.queries.CountActiveMembers(ctx, sqlc.CountActiveMembersParams{BusinessID: businessID})
	if err != nil {
		return 0, err
	}
	if count < 1 {
		return 1, nil
	}
	return int(count), nil
}

// ─── Groups ──────────────────────────────────────────────────────────────────
