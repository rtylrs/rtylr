package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/pos_floors/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_floors/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	posFloorsScope = "pos_floors"
	posFloorsTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID, branchID int64) ([]domain.Floor, error)
	Create(ctx context.Context, input domain.FloorInput) (int64, error)
	Update(ctx context.Context, input domain.FloorUpdateInput) error
	Delete(ctx context.Context, businessID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidatePosFloorsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(posFloorsScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID, branchID int64) ([]domain.Floor, error) {
	key := cacheKeys.Scoped(posFloorsScope, businessID, fmt.Sprintf("list|b=%d", branchID))
	return cacher.Of[[]domain.Floor]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.Floor, error) {
			return s.repo.List(ctx, businessID, branchID)
		},
		cacher.WithExpiry(ctx, posFloorsTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.FloorInput) (int64, error) {
	if err := domain.ValidateFloorInput(input); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidatePosFloorsScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.FloorUpdateInput) error {
	if err := domain.ValidateFloorUpdateInput(input); err != nil {
		return err
	}
	if err := s.repo.Update(ctx, input); err != nil {
		return err
	}
	invalidatePosFloorsScope(ctx, input.BusinessID)
	return nil
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Delete(ctx, businessID, id); err != nil {
		return err
	}
	invalidatePosFloorsScope(ctx, businessID)
	return nil
}
