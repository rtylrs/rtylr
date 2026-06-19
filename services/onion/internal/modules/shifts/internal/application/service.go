package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/shifts/internal/domain"
	"github.com/voxire/onion/internal/modules/shifts/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	shiftsScope = "shifts"
	shiftsTTL   = 2 * time.Minute
)

type repository interface {
	ListShifts(ctx context.Context, businessID, staffID int64, date string) ([]domain.Shift, error)
	CreateShift(ctx context.Context, input domain.ShiftInput) (int64, error)
	UpdateShift(ctx context.Context, businessID int64, input domain.ShiftUpdateInput) error
	DeleteShift(ctx context.Context, businessID, shiftID int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateShiftsScope(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(shiftsScope, businessID))
	}
}

func (s *Service) ListShifts(ctx context.Context, businessID, staffID int64, date string) ([]domain.Shift, error) {
	key := cacheKeys.Scoped(shiftsScope, businessID, fmt.Sprintf("list|s=%d|d=%s", staffID, date))
	return cacher.Of[[]domain.Shift]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.Shift, error) {
			return s.repo.ListShifts(ctx, businessID, staffID, date)
		},
		cacher.WithExpiry(ctx, shiftsTTL),
	)
}

func (s *Service) CreateShift(ctx context.Context, input domain.ShiftInput) (int64, error) {
	id, err := s.repo.CreateShift(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateShiftsScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) UpdateShift(ctx context.Context, businessID int64, input domain.ShiftUpdateInput) error {
	if err := s.repo.UpdateShift(ctx, businessID, input); err != nil {
		return err
	}
	invalidateShiftsScope(ctx, businessID)
	return nil
}

func (s *Service) DeleteShift(ctx context.Context, businessID, shiftID int64) error {
	if err := s.repo.DeleteShift(ctx, businessID, shiftID); err != nil {
		return err
	}
	invalidateShiftsScope(ctx, businessID)
	return nil
}
