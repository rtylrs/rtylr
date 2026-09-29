package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/user/internal/domain"
	"github.com/voxire/onion/internal/modules/user/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/convert"
)

// userFromRow maps the explicit columns selected by
// GetUserByEmail/GetUserByID/GetUserByIdentifier into a domain.User. The three
// queries select the same columns except business_id, which
// GetUserByIdentifier does not need; its caller passes sql.NullInt64{}.
func userFromRow(id int64, businessID sql.NullInt64, email, passwordHash string, fullName, signupIP, deviceInfo sql.NullString, createdAt, updatedAt sql.NullTime, isConsoleAdmin, isBanned bool, tokenEpoch int64) *domain.User {
	u := &domain.User{
		ID:               id,
		Email:            email,
		PasswordHash:     passwordHash,
		FullName:         convert.ToString(fullName),
		SignupDeviceInfo: convert.ToString(deviceInfo),
		IsConsoleAdmin:   isConsoleAdmin,
		IsBanned:         isBanned,
		TokenEpoch:       tokenEpoch,
	}
	if businessID.Valid {
		u.BusinessID = businessID.Int64
	}
	if signupIP.Valid {
		u.SignupIPAddress = signupIP.String
	}
	if createdAt.Valid {
		u.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		u.UpdatedAt = updatedAt.Time
	}
	return u
}

type Repository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		db:      db,
	}
}

func (r *Repository) CreateUser(ctx context.Context, user *domain.User) (int64, error) {
	arg := sqlc.CreateUserParams{
		Email:            user.Email,
		PasswordHash:     user.PasswordHash,
		FullName:         sql.NullString{String: user.FullName, Valid: user.FullName != ""},
		SignupIpAddress:  sql.NullString{String: user.SignupIPAddress, Valid: user.SignupIPAddress != ""},
		SignupDeviceInfo: sql.NullString{String: user.SignupDeviceInfo, Valid: user.SignupDeviceInfo != ""},
	}

	result, err := r.queries.CreateUser(ctx, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create user", logger.Error(err))

		return 0, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read created user ID", logger.Error(err))

		return 0, err
	}

	return userID, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return userFromRow(
		row.ID, row.BusinessID, row.Email, row.PasswordHash,
		row.FullName, row.SignupIpAddress, row.SignupDeviceInfo,
		row.CreatedAt, row.UpdatedAt,
		row.IsConsoleAdmin, row.IsBanned,
		row.TokenEpoch,
	), nil
}

func (r *Repository) GetUserByIdentifier(ctx context.Context, identifier string) (*domain.User, error) {
	row, err := r.queries.GetUserByIdentifier(ctx, sqlc.GetUserByIdentifierParams{
		Email: identifier,
		LOWER: identifier,
	})
	if err != nil {
		return nil, err
	}

	return userFromRow(
		row.ID, sql.NullInt64{}, row.Email, row.PasswordHash,
		row.FullName, row.SignupIpAddress, row.SignupDeviceInfo,
		row.CreatedAt, row.UpdatedAt,
		row.IsConsoleAdmin, row.IsBanned,
		row.TokenEpoch,
	), nil
}

// CreateEmailVerificationCode issues a code for the account owning email. The
// stored digest is an HMAC over the user id and the code, so the code only ever
// verifies the account it was issued for.
func (r *Repository) CreateEmailVerificationCode(ctx context.Context, email string) (*EmailVerification, error) {
	user, err := r.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return nil, err
	}
	code, err := newEmailVerificationCode()
	if err != nil {
		return nil, err
	}
	hash, err := emailVerificationHash(user.ID, code)
	if err != nil {
		return nil, err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO user_email_verification_tokens (user_id, token_hash, expires_at)
		VALUES (?, ?, DATE_ADD(NOW(), INTERVAL 10 MINUTE))`,
		user.ID, hash)
	if err != nil {
		return nil, err
	}
	// A new code starts with a full attempt budget; a user who exhausted the
	// previous one is not locked out by asking for a fresh code.
	emailVerificationAttempts.resetUser(strconv.FormatInt(user.ID, 10))

	return &EmailVerification{
		UserID: user.ID,
		Code:   code,
		Token:  formatEmailVerificationToken(user.ID, code),
	}, nil
}

// ConsumeEmailVerificationCode spends the code carried by token. The token is
// self-describing ("<userID>.<code>"), which is what lets the anonymous
// endpoint recompute the user-bound digest: a code issued for user A cannot
// verify user B, and a code issued for A cannot be replayed after B's
// provisioning rows are touched.
//
// Codes issued before the user binding existed (six digits, stored as an
// unkeyed sha256) are refused outright rather than honoured, because the only
// way to verify them is a digest-only lookup — the exact lookup that lets a
// guessed code spend an unrelated account's row. They are at most ten minutes
// old, and resend-verification mints a replacement in one click.
func (r *Repository) ConsumeEmailVerificationCode(ctx context.Context, token string) error {
	userID, code, ok := parseEmailVerificationToken(token)
	if !ok {
		if isLegacyEmailVerificationCode(token) {
			return ErrSupersededEmailVerificationCode
		}
		return ErrInvalidEmailVerificationCode
	}

	hash, err := emailVerificationHash(userID, code)
	if err != nil {
		return err
	}
	userKey := strconv.FormatInt(userID, 10)
	codeKey := userKey + "|" + hash
	if !emailVerificationAttempts.check(userKey, codeKey) {
		return ErrTooManyEmailVerificationAttempts
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(cause error) error {
		_ = tx.Rollback()
		return cause
	}
	var tokenID int64
	// user_id is in the predicate as well as inside the digest, so the row this
	// spends is the one the code was issued for and not merely the one whose
	// digest collides.
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM user_email_verification_tokens
		WHERE user_id = ? AND token_hash = ? AND consumed_at IS NULL AND expires_at > NOW()
		ORDER BY id DESC
		LIMIT 1
		FOR UPDATE`, userID, hash).Scan(&tokenID)
	if errors.Is(err, sql.ErrNoRows) {
		emailVerificationAttempts.recordFailure(userKey, codeKey)
		return rollback(ErrInvalidEmailVerificationCode)
	}
	if err != nil {
		return rollback(err)
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE user_email_verification_tokens SET consumed_at = NOW() WHERE id = ?`, tokenID); err != nil {
		return rollback(err)
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE user_provisioning_states
		SET email_verified = TRUE,
		    state = CASE WHEN password_change_required THEN 'pending_activation' ELSE 'active' END,
		    updated_at = NOW()
		WHERE user_id = ?`, userID); err != nil {
		return rollback(err)
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE authorization_business_members
		SET status = 'active', authorization_version = authorization_version + 1, updated_at = NOW()
		WHERE user_id = ? AND status IN ('pending', 'suspended')`, userID); err != nil {
		return rollback(err)
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	emailVerificationAttempts.resetUser(userKey)

	return nil
}

// EnsureOwnerAuthorizationMembership writes the business owner's row in
// authorization_business_members.
//
// Creating a business did not write one, so GetAuthorizationSnapshot found no
// row for the owner and left EffectiveAccess.IsOwner false, and every check
// through requireHubPermission answered 403 to the person who had just created
// the business. Meanwhile GET /hub/user/v1/users/me reads businesses.owner_id
// directly and correctly reported is_owner true, so the two sources of truth
// disagreed. Several call sites work around this by short-circuiting on
// userID == ownerID; this writes the row those workarounds were standing in for.
//
// Idempotent, and it never demotes: an owner who somehow already has a row keeps
// their role and is only reactivated.
func (r *Repository) EnsureOwnerAuthorizationMembership(ctx context.Context, businessID, userID int64) error {
	if businessID <= 0 || userID <= 0 {
		return nil
	}

	const ownerRoleKey = "dash.owner"

	var dashRoleID int64
	err := r.db.QueryRowContext(ctx, `
		SELECT id
		FROM authorization_roles
		WHERE business_id = 0 AND app_id = 'dash' AND role_key = ?
		  AND lifecycle = 'active' AND is_archived = FALSE
		LIMIT 1`, ownerRoleKey).Scan(&dashRoleID)
	if err != nil {
		// A deployment without the v2 authorization tables, or without the
		// catalogue seeded, behaves exactly as it did before this existed
		// rather than failing business creation over it.
		if authorizationCatalogueUnavailable(err) {
			logger.Ctx(ctx).Warn("skipping owner authorization membership: dash owner role unavailable",
				logger.Int("business_id", int(businessID)), logger.Error(err))
			return nil
		}
		logger.Ctx(ctx).Error("failed to look up the dash owner role",
			logger.Int("business_id", int(businessID)), logger.Error(err))
		return err
	}

	if _, err = r.db.ExecContext(ctx, `
		INSERT INTO authorization_business_members
			(business_id, user_id, dash_role_id, dash_role_key, status, authorization_version)
		VALUES (?, ?, ?, ?, 'active', 1)
		ON DUPLICATE KEY UPDATE
			status = 'active',
			authorization_version = authorization_version + 1,
			updated_at = NOW()`,
		businessID, userID, dashRoleID, strings.TrimPrefix(ownerRoleKey, "dash."),
	); err != nil {
		if authorizationCatalogueUnavailable(err) {
			logger.Ctx(ctx).Warn("skipping owner authorization membership: table unavailable",
				logger.Int("business_id", int(businessID)), logger.Error(err))
			return nil
		}
		logger.Ctx(ctx).Error("failed to write the owner authorization membership",
			logger.Int("business_id", int(businessID)), logger.Int("user_id", int(userID)), logger.Error(err))
		return err
	}
	return nil
}

// authorizationCatalogueUnavailable reports a deployment where the v2
// authorization tables are absent or unseeded, which is a supported state
// rather than a fault. Mirrors the team module's helper of the same shape.
func authorizationCatalogueUnavailable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, sql.ErrNoRows) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "authorization_") &&
		(strings.Contains(message, "doesn't exist") || strings.Contains(message, "no such table"))
}

// EmailExists reports whether a user row with the given email exists. The
// email is matched as stored (callers normalize case before calling).
func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	exists, err := r.queries.EmailExists(ctx, email)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to check email existence", logger.Error(err))
		return false, err
	}
	return exists, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		// A missing user is a normal outcome while rejecting stale or forged
		// tokens. Keep it out of the server-error stream; callers still receive
		// sql.ErrNoRows and can map it to authentication/not-found behavior.
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("Failed to get user by ID", logger.Error(err))
		}

		return nil, err
	}

	return userFromRow(
		row.ID, row.BusinessID, row.Email, row.PasswordHash,
		row.FullName, row.SignupIpAddress, row.SignupDeviceInfo,
		row.CreatedAt, row.UpdatedAt,
		row.IsConsoleAdmin, row.IsBanned,
		row.TokenEpoch,
	), nil
}

// UserAccessState is intentionally read without the application cache so
// deactivation and bans take effect across processes without waiting for a
// profile-cache TTL.
type UserAccessState struct {
	HasPermissions  bool
	HasActiveAccess bool
}

func (r *Repository) GetUserAccessState(ctx context.Context, userID int64) (UserAccessState, error) {
	row, err := r.queries.GetUserAccessState(ctx, userID)
	if err != nil {
		return UserAccessState{}, err
	}
	return UserAccessState{
		HasPermissions:  row.HasPermissions,
		HasActiveAccess: row.HasActiveAccess.Valid && row.HasActiveAccess.Bool,
	}, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user *domain.User) error {
	arg := sqlc.UpdateUserProfileParams{
		ID:       user.ID,
		FullName: sql.NullString{String: user.FullName, Valid: user.FullName != ""},
	}

	err := r.queries.UpdateUserProfile(ctx, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update user", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) UpdateUserPassword(ctx context.Context, id int64, passwordHash string) error {
	arg := sqlc.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: passwordHash,
	}

	err := r.queries.UpdateUserPassword(ctx, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update user password", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteUser(ctx context.Context, id int64) error {
	err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete user", logger.Error(err))

		return err
	}

	return nil
}

// BusinessNameExists reports whether a business with the given (trimmed) name
// already exists. Business names are globally unique.
func (r *Repository) BusinessNameExists(ctx context.Context, name string) (bool, error) {
	exists, err := r.queries.BusinessNameExists(ctx, name)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to check business name existence", logger.Error(err))
		return false, err
	}
	return exists, nil
}

// GetUserBusinessID returns the business_id for a user. Returns 0 if not set.
func (r *Repository) GetUserBusinessID(ctx context.Context, userID int64) (int64, error) {
	businessID, err := r.queries.GetUserBusinessID(ctx, userID)
	if err != nil {
		return 0, err
	}
	if businessID.Valid {
		return businessID.Int64, nil
	}
	return 0, nil
}

// GetUserRole returns the role for a user within their business.
func (r *Repository) GetUserRole(ctx context.Context, businessID, userID int64) (string, error) {
	// Check if user is the business owner
	ownerID, err := r.queries.GetBusinessOwnerID(ctx, businessID)
	if err == nil && ownerID == userID {
		return "owner", nil
	}

	// Otherwise check permissions table
	role, err := r.queries.GetUserAppRole(ctx, sqlc.GetUserAppRoleParams{
		BusinessID: sql.NullInt64{Int64: businessID, Valid: true},
		UserID:     userID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return "staff", nil
	}
	if err != nil {
		return "", err
	}
	return string(role.UserAppPermissionsRole), nil
}

// CreateBusiness creates a new business and returns its ID.
func (r *Repository) CreateBusiness(ctx context.Context, name string, ownerID int64) (int64, error) {
	res, err := r.queries.CreateBusiness(ctx, sqlc.CreateBusinessParams{
		Name:    name,
		OwnerID: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create business", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

// SetUserBusinessID sets the business_id for a user.
func (r *Repository) SetUserBusinessID(ctx context.Context, userID, businessID int64) error {
	return r.queries.SetUserBusinessID(ctx, sqlc.SetUserBusinessIDParams{
		BusinessID: sql.NullInt64{Int64: businessID, Valid: true},
		ID:         userID,
	})
}

// EnsureEmployeeForUser creates the HR employee record for a user. The business
// is passed in rather than read from the user's own row: at invite time
// users.business_id is still NULL, so reading it there scoped the record to no
// business at all.
func (r *Repository) EnsureEmployeeForUser(ctx context.Context, ownerID, businessID, userID int64) error {
	return r.queries.EnsureEmployeeForUser(ctx, sqlc.EnsureEmployeeForUserParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		ID:         userID,
	})
}

// RecordLogin writes a login history entry via sqlc.
func (r *Repository) RecordLogin(ctx context.Context, userID int64, email, method, ip, userAgent string, success bool) error {
	return r.queries.RecordLoginHistory(ctx, sqlc.RecordLoginHistoryParams{
		UserID:    userID,
		Email:     email,
		Method:    method,
		IpAddress: sql.NullString{String: ip, Valid: ip != ""},
		UserAgent: sql.NullString{String: userAgent, Valid: userAgent != ""},
		Success:   sql.NullBool{Bool: success, Valid: true},
	})
}

// CheckBusinessAccess reports whether userID may switch into businessID.
// Access is granted if the user directly owns the business (businesses.owner_id = userID)
// or has been explicitly granted access via user_business_access.
func (r *Repository) CheckBusinessAccess(ctx context.Context, userID, businessID int64) (bool, error) {
	// cross-schema read from user_business_access — multi-business switch validation
	cnt, err := r.queries.CheckBusinessAccess(ctx, sqlc.CheckBusinessAccessParams{
		ID:         businessID,
		OwnerID:    userID,
		UserID:     userID,
		BusinessID: businessID,
	})
	return cnt > 0, err
}

// AssignSubscription looks up the plan by slug and creates the owner's initial subscription.
func (r *Repository) AssignSubscription(ctx context.Context, ownerID int64, planSlug string) error {
	// cross-schema read from plans — approved for subscription assignment on registration
	plan, err := r.queries.GetActivePlanForSignup(ctx, planSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("plan not found: %s", planSlug)
		}

		return err
	}

	status := sqlc.SubscriptionsStatusActive
	trialEnd := sql.NullTime{}
	if plan.TrialDays > 0 {
		status = sqlc.SubscriptionsStatusTrial
		trialEnd = sql.NullTime{Time: time.Now().AddDate(0, 0, int(plan.TrialDays)), Valid: true}
	}

	// cross-schema write to subscriptions — approved for subscription creation on registration
	return r.queries.UpsertSignupSubscription(ctx, sqlc.UpsertSignupSubscriptionParams{
		OwnerID:     ownerID,
		PlanID:      plan.ID,
		Status:      sqlc.NullSubscriptionsStatus{SubscriptionsStatus: status, Valid: true},
		TrialEndsAt: trialEnd,
	})
}
