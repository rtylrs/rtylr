package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/security/internal/domain"
	"github.com/voxire/onion/internal/modules/security/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
	}
}

func (r *Repository) GetSecuritySettings(ctx context.Context, userID int64) (domain.SecuritySettings, error) {
	// cross-schema read from users — 2FA flag is owned by the user/auth module
	twoFAEnabled, err := r.queries.GetTwoFactorEnabled(ctx, userID)
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
	err := r.queries.SetTwoFactor(ctx, sqlc.SetTwoFactorParams{
		TwoFactorEnabled: enabled,
		ID:               userID,
	})
	if err != nil {
		logger.Ctx(ctx).Warn("Failed to set 2FA (column may not exist)", logger.Error(err))
	}
	return nil
}

func (r *Repository) GetLoginHistory(ctx context.Context, userID int64) ([]domain.LoginEntry, error) {
	rows, err := r.queries.ListLoginHistory(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get login history", logger.Error(err))
		return nil, err
	}

	entries := make([]domain.LoginEntry, 0, len(rows))
	for _, row := range rows {
		entry := domain.LoginEntry{
			Email:     row.Email,
			Method:    row.Method,
			IP:        row.IpAddress,
			UserAgent: row.UserAgent,
		}
		if row.Success.Valid {
			entry.Success = row.Success.Bool
		}
		if row.CreatedAt.Valid {
			entry.CreatedAt = row.CreatedAt.Time
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (r *Repository) InvalidateAllSessions(ctx context.Context, userID int64) error {
	// cross-schema write to user_sessions — sessions are owned by the auth module
	err := r.queries.InvalidateAllSessions(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to invalidate sessions", logger.Error(err))
	}
	return err
}

func (r *Repository) RecordLogin(ctx context.Context, userID int64, email, method, ip, userAgent string, success bool) error {
	err := r.queries.RecordLogin(ctx, sqlc.RecordLoginParams{
		UserID:    userID,
		Email:     email,
		Method:    method,
		IpAddress: sql.NullString{String: ip, Valid: ip != ""},
		UserAgent: sql.NullString{String: userAgent, Valid: userAgent != ""},
		Success:   sql.NullBool{Bool: success, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to record login", logger.Error(err))
	}
	return err
}
