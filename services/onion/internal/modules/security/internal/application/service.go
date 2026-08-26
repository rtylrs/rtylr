package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/security/internal/domain"
	"github.com/voxire/onion/internal/modules/security/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

// Security data is per-user; everything lives under cacheKeys.UserPrefix so
// the public /invalidate-cache?type=user&id=X path wipes it for free.
const securityTTL = 5 * time.Minute

const (
	defaultLoginHistoryPageSize int64 = 20
	maxLoginHistoryPageSize     int64 = 100
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func invalidateSecurityScope(ctx context.Context, userID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.UserPrefix(userID))
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

type loginHistoryPage struct {
	Entries []domain.LoginEntry
	Total   int64
	HasMore bool
}

// NormalizeLoginHistoryPagination applies the public defaults and upper bound
// used by both the API and the cached repository query.
func NormalizeLoginHistoryPagination(page, pageSize int64) (int64, int64) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultLoginHistoryPageSize
	}
	if pageSize > maxLoginHistoryPageSize {
		pageSize = maxLoginHistoryPageSize
	}
	return page, pageSize
}

func (s *Service) GetLoginHistory(ctx context.Context, userID, page, pageSize int64) ([]domain.LoginEntry, int64, bool, error) {
	page, pageSize = NormalizeLoginHistoryPagination(page, pageSize)
	key := cacheKeys.Scoped("user", userID, fmt.Sprintf("login_history|p=%d|ps=%d", page, pageSize))
	result, err := cacher.Of[loginHistoryPage]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (loginHistoryPage, error) {
			entries, total, err := s.repo.GetLoginHistory(
				ctx,
				userID,
				int(pageSize+1),
				int((page-1)*pageSize),
			)
			if err != nil {
				return loginHistoryPage{}, err
			}

			hasMore := len(entries) > int(pageSize)
			if hasMore {
				entries = entries[:pageSize]
			}
			return loginHistoryPage{Entries: entries, Total: total, HasMore: hasMore}, nil
		},
		cacher.WithExpiry(ctx, 1*time.Minute),
	)
	return result.Entries, result.Total, result.HasMore, err
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
