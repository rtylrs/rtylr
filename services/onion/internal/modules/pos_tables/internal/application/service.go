package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/pos_tables/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_tables/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	posTablesScope = "pos_tables"
	posTablesTTL   = 1 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.Table, error)
	Create(ctx context.Context, input domain.TableInput) (int64, error)
	Update(ctx context.Context, input domain.TableUpdateInput) error
	UpdateStatus(ctx context.Context, input domain.TableStatusInput) error
	Delete(ctx context.Context, businessID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidatePosTablesScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(posTablesScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Table, error) {
	return cacher.Of[[]domain.Table]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(posTablesScope, businessID, "list"),
		func(ctx context.Context) ([]domain.Table, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, posTablesTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.TableInput) (int64, error) {
	if err := domain.ValidateTableInput(input); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidatePosTablesScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.TableUpdateInput) error {
	if err := domain.ValidateTableUpdateInput(input); err != nil {
		return err
	}
	if err := s.repo.Update(ctx, input); err != nil {
		return err
	}
	invalidatePosTablesScope(ctx, input.BusinessID)
	return nil
}

func (s *Service) UpdateStatus(ctx context.Context, input domain.TableStatusInput) error {
	if err := domain.ValidateTableStatusInput(input); err != nil {
		return err
	}
	if err := s.repo.UpdateStatus(ctx, input); err != nil {
		return err
	}
	invalidatePosTablesScope(ctx, input.BusinessID)
	return nil
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Delete(ctx, businessID, id); err != nil {
		return err
	}
	invalidatePosTablesScope(ctx, businessID)
	return nil
}
