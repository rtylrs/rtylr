package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/organization/user/internal/domain"
	"github.com/voxire/onion/internal/modules/organization/user/internal/infrastructure/persistence/sqlc"
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
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

func (r *Repository) GetUserByIdentifier(ctx context.Context, identifier string) (*domain.User, error) {
	row, err := r.queries.GetUserByIdentifier(ctx, sqlc.GetUserByIdentifierParams{
		Email: identifier,
		LOWER: identifier,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
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
func (r *Repository) CreateEmailVerificationCode(ctx context.Context, email string) (*domain.EmailVerification, error) {
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

	return &domain.EmailVerification{
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
func (r *Repository) ConsumeEmailVerificationCode(ctx context.Context, token string) (int64, error) {
	userID, code, ok := parseEmailVerificationToken(token)
	if !ok {
		if isLegacyEmailVerificationCode(token) {
			return 0, ErrSupersededEmailVerificationCode
		}
		return 0, ErrInvalidEmailVerificationCode
	}

	hash, err := emailVerificationHash(userID, code)
	if err != nil {
		return 0, err
	}
	userKey := strconv.FormatInt(userID, 10)
	codeKey := userKey + "|" + hash
	if !emailVerificationAttempts.check(userKey, codeKey) {
		return 0, ErrTooManyEmailVerificationAttempts
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
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
		return 0, rollback(ErrInvalidEmailVerificationCode)
	}
	if err != nil {
		return 0, rollback(err)
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE user_email_verification_tokens SET consumed_at = NOW() WHERE id = ?`, tokenID); err != nil {
		return 0, rollback(err)
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE user_provisioning_states
		SET email_verified = TRUE,
		    state = CASE WHEN password_change_required THEN 'pending_activation' ELSE 'active' END,
		    updated_at = NOW()
		WHERE user_id = ?`, userID); err != nil {
		return 0, rollback(err)
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE authorization_business_members
		SET status = 'active', authorization_version = authorization_version + 1, updated_at = NOW()
		WHERE user_id = ? AND status IN ('pending', 'suspended')`, userID); err != nil {
		return 0, rollback(err)
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	emailVerificationAttempts.resetUser(userKey)

	return userID, nil
}

// CreateSignupProvisioning records a self-signup account. It starts active, so
// /users/me loads, and the auth middleware keeps it to the verification
// endpoints until emailVerified is true.
func (r *Repository) CreateSignupProvisioning(ctx context.Context, userID, businessID int64, emailVerified bool) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_provisioning_states (user_id, business_id, mode, state, email_verified, password_change_required)
		VALUES (?, ?, 'self_signup', 'active', ?, FALSE)`, userID, businessID, emailVerified)
	return err
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
		// tokens. Keep it out of the server-error stream.
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("Failed to get user by ID", logger.Error(err))
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
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
func (r *Repository) GetUserAccessState(ctx context.Context, userID int64) (domain.UserAccessState, error) {
	row, err := r.queries.GetUserAccessState(ctx, userID)
	if err != nil {
		return domain.UserAccessState{}, err
	}
	return domain.UserAccessState{
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `
		UPDATE users SET password_hash = ?, token_epoch = token_epoch + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, passwordHash, id); err != nil {
		logger.Ctx(ctx).Error("Failed to update user password", logger.Error(err))
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE user_sessions SET revoked_at = CURRENT_TIMESTAMP WHERE user_id = ? AND revoked_at IS NULL`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) CreateSession(ctx context.Context, userID int64, tokenHash, ip, userAgent, method string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_sessions (user_id, token_hash, ip_address, user_agent, method, expires_at)
		VALUES (?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?)`, userID, tokenHash, ip, userAgent, method, expiresAt)
	return err
}

func (r *Repository) GetSessionUserID(ctx context.Context, tokenHash string) (int64, error) {
	var userID int64
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id FROM user_sessions
		WHERE token_hash = ? AND revoked_at IS NULL AND expires_at > CURRENT_TIMESTAMP
		ORDER BY id DESC LIMIT 1`, tokenHash).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrSessionNotFound
	}
	return userID, err
}

func (r *Repository) RevokeSession(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE user_sessions SET revoked_at = CURRENT_TIMESTAMP WHERE token_hash = ? AND revoked_at IS NULL`, tokenHash)
	return err
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
// or has an active authorization_business_members row.
func (r *Repository) CheckBusinessAccess(ctx context.Context, userID, businessID int64) (bool, error) {
	// cross-schema read from normalized tenant memberships — business switch validation
	cnt, err := r.queries.CheckBusinessAccess(ctx, sqlc.CheckBusinessAccessParams{
		ID:         businessID,
		OwnerID:    userID,
		UserID:     userID,
		BusinessID: businessID,
	})
	return cnt > 0, err
}

// AssignSubscription looks up the plan by slug and creates the subscription for
// the business the user registered with. A user without a business has nothing
// to bill yet; a business created later starts on the default plan.
func (r *Repository) AssignSubscription(ctx context.Context, userID int64, planSlug string) error {
	businessID, err := r.queries.GetOwnedDefaultBusinessID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

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
		BusinessID:  businessID,
		PlanID:      plan.ID,
		Status:      sqlc.NullSubscriptionsStatus{SubscriptionsStatus: status, Valid: true},
		TrialEndsAt: trialEnd,
	})
}
