package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/security/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

func (r *Repository) GetSecuritySettings(ctx context.Context, userID int64) (domain.SecuritySettings, error) {
	// Check if 2FA is enabled for the user
	var twoFAEnabled bool

	err := r.pool.QueryRowContext(ctx,
		`SELECT COALESCE(two_factor_enabled, FALSE) FROM users WHERE id = ?`, userID,
	).Scan(&twoFAEnabled)
	if err != nil {
		if err == sql.ErrNoRows {
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
	query := `UPDATE users SET two_factor_enabled = ? WHERE id = ?`

	_, err := r.pool.ExecContext(ctx, query, enabled, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to set 2FA", logger.Error(err))
		return err
	}

	return nil
}
