package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/pos_member_locations/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_member_locations/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	posMemberLocsScope = "pos_member_locations"
	posMemberLocsTTL   = 5 * time.Minute
)

type Service struct{ repo *infrastructure.Repository }

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidatePosMemberLocsScope(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(posMemberLocsScope, businessID))
	}
}

func (s *Service) ListForUser(ctx context.Context, businessID, userID int64) ([]domain.MemberLocation, error) {
	key := cacheKeys.Scoped(posMemberLocsScope, businessID, fmt.Sprintf("for|u=%d", userID))
	return cacher.Of[[]domain.MemberLocation]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.MemberLocation, error) {
			return s.repo.ListForUser(ctx, businessID, userID)
		},
		cacher.WithExpiry(ctx, posMemberLocsTTL),
	)
}

func (s *Service) GetMyLocations(ctx context.Context, businessID, userID int64) ([]infrastructure.MyLocation, error) {
	key := cacheKeys.Scoped(posMemberLocsScope, businessID, fmt.Sprintf("mine|u=%d", userID))
	return cacher.Of[[]infrastructure.MyLocation]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]infrastructure.MyLocation, error) {
			return s.repo.GetMyLocations(ctx, businessID, userID)
		},
		cacher.WithExpiry(ctx, posMemberLocsTTL),
	)
}

// GetLocationRoster returns the POS sign-in roster for a location. Short TTL:
// it reflects has_pin which flips as soon as an operator onboards.
func (s *Service) GetLocationRoster(ctx context.Context, businessID, locationID int64) ([]infrastructure.RosterEntry, error) {
	key := cacheKeys.Scoped(posMemberLocsScope, businessID, fmt.Sprintf("roster|l=%d", locationID))
	return cacher.Of[[]infrastructure.RosterEntry]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]infrastructure.RosterEntry, error) {
			return s.repo.GetLocationRoster(ctx, businessID, locationID)
		},
		cacher.WithExpiry(ctx, 30*time.Second),
	)
}

func (s *Service) Assign(ctx context.Context, businessID, userID, locationID int64, canManagePOSUsers, canManageTerminals, canViewReports bool) (int64, error) {
	id, err := s.repo.Assign(ctx, businessID, userID, locationID, canManagePOSUsers, canManageTerminals, canViewReports)
	if err != nil {
		return 0, err
	}
	invalidatePosMemberLocsScope(ctx, businessID)
	return id, nil
}

func (s *Service) Remove(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Remove(ctx, businessID, id); err != nil {
		return err
	}
	invalidatePosMemberLocsScope(ctx, businessID)
	return nil
}
