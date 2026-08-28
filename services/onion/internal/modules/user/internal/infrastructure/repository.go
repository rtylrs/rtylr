package infrastructure

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
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

// userFromRow maps the explicit columns selected by GetUserByEmail/GetUserByID
// into a domain.User. Both queries select the same column set, so they share
// this mapper.
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
		row.IsConsoleAdmin, false,
		row.TokenEpoch,
	), nil
}

func (r *Repository) CreateEmailVerificationCode(ctx context.Context, email string) (string, error) {
	user, err := r.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return "", err
	}
	var randomBytes [4]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", err
	}
	code := fmt.Sprintf("%06d", binary.BigEndian.Uint32(randomBytes[:])%1_000_000)
	hash := sha256.Sum256([]byte(code))
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO user_email_verification_tokens (user_id, token_hash, expires_at)
		VALUES (?, ?, DATE_ADD(NOW(), INTERVAL 10 MINUTE))`,
		user.ID, hex.EncodeToString(hash[:]))
	if err != nil {
		return "", err
	}
	return code, nil
}

func (r *Repository) ConsumeEmailVerificationCode(ctx context.Context, token string) error {
	code := strings.TrimSpace(token)
	if len(code) != 6 {
		return errors.New("invalid email verification code")
	}
	if _, err := strconv.Atoi(code); err != nil {
		return errors.New("invalid email verification code")
	}
	hash := sha256.Sum256([]byte(code))
	hashText := hex.EncodeToString(hash[:])
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(cause error) error {
		_ = tx.Rollback()
		return cause
	}
	var userID int64
	var tokenID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id, user_id
		FROM user_email_verification_tokens
		WHERE token_hash = ? AND consumed_at IS NULL AND expires_at > NOW()
		ORDER BY id DESC
		LIMIT 1
		FOR UPDATE`, hashText).Scan(&tokenID, &userID)
	if errors.Is(err, sql.ErrNoRows) {
		return rollback(errors.New("invalid or expired email verification code"))
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
	return tx.Commit()
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
	if err != nil {
		return "staff", nil // default
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

func (r *Repository) EnsureEmployeeForUser(ctx context.Context, ownerID, userID int64) error {
	return r.queries.EnsureEmployeeForUser(ctx, sqlc.EnsureEmployeeForUserParams{
		OwnerID: ownerID,
		ID:      userID,
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
