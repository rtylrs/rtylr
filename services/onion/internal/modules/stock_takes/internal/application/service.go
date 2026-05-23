package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/stock_takes/internal/domain"
	"github.com/voxire/onion/internal/modules/stock_takes/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const stockTakesCacheTTL = 5 * time.Minute

type repository interface {
	ListStockTakes(ctx context.Context, ownerID int64) ([]domain.StockTake, error)
	CreateStockTake(ctx context.Context, ownerID int64, input domain.StockTakeInput) (int64, error)
	UpdateStockTake(ctx context.Context, ownerID int64, input domain.StockTakeUpdateInput) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	repo := infrastructure.NewRepository(db)

	return &Service{
		repo: repo,
	}
}

func (s *Service) ListStockTakes(ctx context.Context, ownerID int64) ([]domain.StockTake, error) {
	cacheKey := fmt.Sprintf("stock_takes:%d", ownerID)

	return cacher.Of[[]domain.StockTake]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.StockTake, error) {
		return s.repo.ListStockTakes(ctx, ownerID)
	}, cacher.WithExpiry(ctx, stockTakesCacheTTL))
}

func (s *Service) CreateStockTake(ctx context.Context, ownerID int64, input domain.StockTakeInput) (int64, error) {
	id, err := s.repo.CreateStockTake(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}

	s.invalidateCache(ctx, ownerID)

	return id, nil
}

func (s *Service) UpdateStockTake(ctx context.Context, ownerID int64, input domain.StockTakeUpdateInput) error {
	err := s.repo.UpdateStockTake(ctx, ownerID, input)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, ownerID)

	return nil
}

func (s *Service) invalidateCache(ctx context.Context, ownerID int64) {
	prefix := fmt.Sprintf("stock_takes:%d", ownerID)
	cacher.Of[[]domain.StockTake]().Invalidate(ctx, []string{prefix})
}
