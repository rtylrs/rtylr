package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/security/internal/domain"
	"github.com/voxire/onion/internal/modules/security/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

// Security data is per-user; everything lives under cacheKeys.UserPrefix so
// the public /invalidate-cache?type=user&id=X path wipes it for free.
const securityTTL = 5 * time.Minute

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func invalidateSecurityScope(ctx context.Context, userID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.UserPrefix(userID))
	}
}

func (s *Service) GetSettings(ctx context.Context, userID int64) (domain.SecuritySettings, error) {
	key := cacheKeys.Scoped("user", userID, "security_settings")
	return cacher.Of[domain.SecuritySettings]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.SecuritySettings, error) {
			return s.repo.GetSecuritySettings(ctx, userID)
		},
		cacher.WithExpiry(ctx, securityTTL),
	)
}

func (s *Service) SetTwoFactor(ctx context.Context, userID int64, enabled bool) error {
	if err := s.repo.SetTwoFactor(ctx, userID, enabled); err != nil {
		return err
	}
	invalidateSecurityScope(ctx, userID)
	return nil
}

func (s *Service) GetLoginHistory(ctx context.Context, userID int64) ([]domain.LoginEntry, error) {
	key := cacheKeys.Scoped("user", userID, "login_history")
	return cacher.Of[[]domain.LoginEntry]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.LoginEntry, error) {
			return s.repo.GetLoginHistory(ctx, userID)
		},
		cacher.WithExpiry(ctx, 1*time.Minute),
	)
}

func (s *Service) InvalidateAllSessions(ctx context.Context, userID int64) error {
	if err := s.repo.InvalidateAllSessions(ctx, userID); err != nil {
		return err
	}
	invalidateSecurityScope(ctx, userID)
	return nil
}

// RecordLogin stores a login event in the history table; invalidates the
// per-user login_history cache so the next GET is fresh.
func (s *Service) RecordLogin(ctx context.Context, userID int64, email, method, ip, userAgent string, success bool) {
	_ = s.repo.RecordLogin(ctx, userID, email, method, ip, userAgent, success)
	invalidateSecurityScope(ctx, userID)
}
