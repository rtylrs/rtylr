package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/user/internal/domain"
	"github.com/voxire/onion/internal/modules/user/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/convert"
)

// userFromRow maps the explicit columns selected by GetUserByEmail/GetUserByID
// into a domain.User. Both queries select the same column set, so they share
// this mapper.
func userFromRow(id int64, email, passwordHash string, fullName, signupIP, deviceInfo sql.NullString, createdAt, updatedAt sql.NullTime) *domain.User {
	u := &domain.User{
		ID:               id,
		Email:            email,
		PasswordHash:     passwordHash,
		FullName:         convert.ToString(fullName),
		SignupDeviceInfo: convert.ToString(deviceInfo),
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
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
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
		row.ID, row.Email, row.PasswordHash,
		row.FullName, row.SignupIpAddress, row.SignupDeviceInfo,
		row.CreatedAt, row.UpdatedAt,
	), nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get user by ID", logger.Error(err))

		return nil, err
	}

	return userFromRow(
		row.ID, row.Email, row.PasswordHash,
		row.FullName, row.SignupIpAddress, row.SignupDeviceInfo,
		row.CreatedAt, row.UpdatedAt,
	), nil
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

// GetUserBusinessID returns the business_id for a user. Returns 0 if not set.
func (r *Repository) GetUserBusinessID(ctx context.Context, userID int64) (int64, error) {
	var businessID sql.NullInt64
	err := r.pool.QueryRowContext(ctx, "SELECT business_id FROM users WHERE id = ?", userID).Scan(&businessID)
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
	var ownerID int64
	err := r.pool.QueryRowContext(ctx, "SELECT owner_id FROM businesses WHERE id = ?", businessID).Scan(&ownerID)
	if err == nil && ownerID == userID {
		return "owner", nil
	}

	// Otherwise check permissions table
	var role string
	err = r.pool.QueryRowContext(ctx,
		"SELECT role FROM user_app_permissions WHERE business_id = ? AND user_id = ?",
		businessID, userID,
	).Scan(&role)
	if err != nil {
		return "staff", nil // default
	}
	return role, nil
}

// CreateBusiness creates a new business and returns its ID.
func (r *Repository) CreateBusiness(ctx context.Context, name string, ownerID int64) (int64, error) {
	res, err := r.pool.ExecContext(ctx,
		"INSERT INTO businesses (name, owner_id) VALUES (?, ?)",
		name, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create business", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

// SetUserBusinessID sets the business_id for a user.
func (r *Repository) SetUserBusinessID(ctx context.Context, userID, businessID int64) error {
	_, err := r.pool.ExecContext(ctx, "UPDATE users SET business_id = ? WHERE id = ?", businessID, userID)
	return err
}

// RecordLogin writes a login history entry via sqlc.
func (r *Repository) RecordLogin(ctx context.Context, userID int64, email, method, ip string, success bool) error {
	return r.queries.RecordLoginHistory(ctx, sqlc.RecordLoginHistoryParams{
		UserID:    userID,
		Email:     email,
		Method:    method,
		IpAddress: sql.NullString{String: ip, Valid: ip != ""},
		Success:   sql.NullBool{Bool: success, Valid: true},
	})
}

// AssignSubscription looks up the plan by slug and creates a trial subscription for the owner.
func (r *Repository) AssignSubscription(ctx context.Context, ownerID int64, planSlug string) error {
	var planID int64
	// cross-schema read from plans — approved for subscription assignment on registration
	err := r.pool.QueryRowContext(ctx,
		`SELECT id FROM plans WHERE slug = ? AND is_active = TRUE`,
		planSlug).Scan(&planID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("plan not found: %s", planSlug)
		}

		return err
	}

	trialEnd := time.Now().AddDate(0, 0, 14)
	// cross-schema write to subscriptions — approved for trial creation on registration
	_, execErr := r.pool.ExecContext(ctx,
		`INSERT INTO subscriptions (owner_id, plan_id, status, trial_ends_at)
		 VALUES (?, ?, 'trial', ?)
		 ON DUPLICATE KEY UPDATE plan_id = VALUES(plan_id), updated_at = CURRENT_TIMESTAMP`,
		ownerID, planID, trialEnd)

	return execErr
}
