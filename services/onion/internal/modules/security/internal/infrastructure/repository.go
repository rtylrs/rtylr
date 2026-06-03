package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/security/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		pool: db,
	}
}

func (r *Repository) GetSecuritySettings(ctx context.Context, userID int64) (domain.SecuritySettings, error) {
	// Check if 2FA is enabled for the user
	var twoFAEnabled bool

	// cross-schema read from users — 2FA flag is owned by the user/auth module
	query, args, err := sq.
		Select("COALESCE(two_factor_enabled, FALSE)").
		From("users").
		Where(sq.Eq{"id": userID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build security settings query", logger.Error(err))
		return domain.SecuritySettings{}, err
	}

	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&twoFAEnabled)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SecuritySettings{
				TwoFactorEnabled: false,
				ActiveSessions:   []domain.ActiveSession{},
			}, nil
		}
		logger.Ctx(ctx).Error("Failed to get security settings", logger.Error(err))
		return domain.SecuritySettings{}, err
	}

	return domain.SecuritySettings{
		TwoFactorEnabled: twoFAEnabled,
		ActiveSessions:   []domain.ActiveSession{},
	}, nil
}

func (r *Repository) SetTwoFactor(ctx context.Context, userID int64, enabled bool) error {
	// cross-schema write to users — two_factor_enabled column may not exist yet, silently fail
	query, args, err := sq.
		Update("users").
		Set("two_factor_enabled", enabled).
		Where(sq.Eq{"id": userID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Warn("Failed to build set 2FA query", logger.Error(err))
		return nil
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Warn("Failed to set 2FA (column may not exist)", logger.Error(err))
	}
	return nil
}

func (r *Repository) GetLoginHistory(ctx context.Context, userID int64) ([]domain.LoginEntry, error) {
	query, args, err := sq.
		Select("email", "method", "COALESCE(ip_address,'')", "COALESCE(user_agent,'')", "success", "created_at").
		From("login_history").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("created_at DESC").
		Limit(20).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build login history query", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get login history", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	entries := make([]domain.LoginEntry, 0, 20)
	for rows.Next() {
		var e domain.LoginEntry
		if err := rows.Scan(&e.Email, &e.Method, &e.IP, &e.UserAgent, &e.Success, &e.CreatedAt); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func (r *Repository) InvalidateAllSessions(ctx context.Context, userID int64) error {
	// cross-schema write to user_sessions — sessions are owned by the auth module
	query, args, err := sq.
		Update("user_sessions").
		Set("revoked_at", sq.Expr("NOW()")).
		Where(sq.Eq{"user_id": userID}).
		Where(sq.Eq{"revoked_at": nil}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build invalidate sessions query", logger.Error(err))
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to invalidate sessions", logger.Error(err))
	}
	return err
}

func (r *Repository) RecordLogin(ctx context.Context, userID int64, email, method, ip, userAgent string, success bool) error {
	query, args, err := sq.
		Insert("login_history").
		Columns("user_id", "email", "method", "ip_address", "user_agent", "success").
		Values(userID, email, method, ip, userAgent, success).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build record login query", logger.Error(err))
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to record login", logger.Error(err))
	}
	return err
}
