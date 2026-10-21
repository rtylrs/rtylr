package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/pos/memberlocations/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	posMemberLocsScope = "pos_member_locations"
	posMemberLocsTTL   = 5 * time.Minute
)

var ErrInvalidAssignment = errors.New("invalid POS location assignment")

type Service struct{ repo repository }

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

func invalidatePosMemberLocsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(posMemberLocsScope, businessID))
}

func (s *Service) ListForUser(ctx context.Context, businessID, userID int64) ([]domain.MemberLocation, error) {
	key := cacheKeys.Scoped(posMemberLocsScope, businessID, fmt.Sprintf("for|u=%d", userID))
	return cacher.Of[[]domain.MemberLocation]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) ([]domain.MemberLocation, error) {
			return s.repo.ListForUser(ctx, businessID, userID)
		},
		cacher.WithExpiry(ctx, posMemberLocsTTL),
	)
}

func (s *Service) GetMyLocations(ctx context.Context, businessID, userID int64) ([]domain.MyLocation, error) {
	key := cacheKeys.Scoped(posMemberLocsScope, businessID, fmt.Sprintf("mine|u=%d", userID))
	return cacher.Of[[]domain.MyLocation]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) ([]domain.MyLocation, error) {
			return s.repo.GetMyLocations(ctx, businessID, userID)
		},
		cacher.WithExpiry(ctx, posMemberLocsTTL),
	)
}

func (s *Service) GetBusinessLocations(ctx context.Context, businessID int64) ([]domain.MyLocation, error) {
	key := cacheKeys.Scoped(posMemberLocsScope, businessID, "business_locations|domain=v2")
	return cacher.Of[[]domain.MyLocation]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) ([]domain.MyLocation, error) {
			return s.repo.GetBusinessLocations(ctx, businessID)
		},
		cacher.WithExpiry(ctx, posMemberLocsTTL),
	)
}

// GetLocationRoster returns the POS sign-in roster for a location. Short TTL:
// it reflects has_pin which flips as soon as an operator onboards.
func (s *Service) GetLocationRoster(ctx context.Context, businessID, locationID int64) ([]domain.RosterEntry, error) {
	key := cacheKeys.Scoped(posMemberLocsScope, businessID, fmt.Sprintf("roster|l=%d", locationID))
	return cacher.Of[[]domain.RosterEntry]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) ([]domain.RosterEntry, error) {
			return s.repo.GetLocationRoster(ctx, businessID, locationID)
		},
		cacher.WithExpiry(ctx, 30*time.Second),
	)
}

// IsAssignedToLocation reports whether the user has a pos_member_locations row
// for the location. This is an authorization gate, so it reads the DB directly
// instead of trusting the cached per-user location list.
func (s *Service) IsAssignedToLocation(ctx context.Context, businessID, userID, locationID int64) (bool, error) {
	return s.repo.IsAssignedToLocation(ctx, businessID, userID, locationID)
}

func (s *Service) Assign(ctx context.Context, businessID, userID, locationID int64, canManagePOSUsers, canManageTerminals, canViewReports bool) (int64, error) {
	if businessID <= 0 || userID <= 0 || locationID <= 0 {
		return 0, ErrInvalidAssignment
	}
	userOK, err := s.repo.UserBelongsToBusiness(ctx, businessID, userID)
	if err != nil {
		return 0, err
	}
	locationOK, err := s.repo.LocationBelongsToBusiness(ctx, businessID, locationID)
	if err != nil {
		return 0, err
	}
	if !userOK || !locationOK {
		return 0, ErrInvalidAssignment
	}

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
