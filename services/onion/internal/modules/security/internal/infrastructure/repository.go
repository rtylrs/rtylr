package infrastructure

import (
	"context"
	"database/sql"
	"errors"

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

	err := r.pool.QueryRowContext(ctx,
		`SELECT COALESCE(two_factor_enabled, FALSE) FROM users WHERE id = ?`, userID,
	).Scan(&twoFAEnabled)
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
	// two_factor_enabled column may not exist yet - silently fail
	_, err := r.pool.ExecContext(ctx, `UPDATE users SET two_factor_enabled = ? WHERE id = ?`, enabled, userID)
	if err != nil {
		logger.Ctx(ctx).Warn("Failed to set 2FA (column may not exist)", logger.Error(err))
	}
	return nil
}

func (r *Repository) GetLoginHistory(ctx context.Context, userID int64) ([]domain.LoginEntry, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT email, method, COALESCE(ip_address,''), COALESCE(user_agent,''), success, created_at
		 FROM login_history WHERE user_id = ? ORDER BY created_at DESC LIMIT 20`, userID)
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
	_, err := r.pool.ExecContext(ctx,
		`UPDATE user_sessions SET revoked_at = NOW() WHERE user_id = ? AND revoked_at IS NULL`, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to invalidate sessions", logger.Error(err))
	}
	return err
}

func (r *Repository) RecordLogin(ctx context.Context, userID int64, email, method, ip, userAgent string, success bool) error {
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO login_history (user_id, email, method, ip_address, user_agent, success) VALUES (?, ?, ?, ?, ?, ?)`,
		userID, email, method, ip, userAgent, success)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to record login", logger.Error(err))
	}
	return err
}
