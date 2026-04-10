package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/stock_transfers/internal/domain"
	"github.com/voxire/onion/internal/modules/stock_transfers/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const stockTransfersCacheTTL = 5 * time.Minute

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	repo := infrastructure.NewRepository(ctx)

	return &Service{
		repo: repo,
	}
}

func (s *Service) ListStockTransfers(ctx context.Context, ownerID int64) ([]domain.StockTransfer, error) {
	cacheKey := fmt.Sprintf("stock_transfers:%d", ownerID)

	return cacher.Of[[]domain.StockTransfer]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.StockTransfer, error) {
		return s.repo.ListStockTransfers(ctx, ownerID)
	}, cacher.WithExpiry(ctx, stockTransfersCacheTTL))
}

func (s *Service) CreateStockTransfer(ctx context.Context, ownerID int64, input domain.StockTransferInput) (int64, error) {
	id, err := s.repo.CreateStockTransfer(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}

	s.invalidateCache(ctx, ownerID)

	return id, nil
}

func (s *Service) UpdateStockTransferStatus(ctx context.Context, ownerID, id int64, status domain.StockTransferStatus) error {
	err := s.repo.UpdateStockTransferStatus(ctx, ownerID, id, status)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, ownerID)

	return nil
}

func (s *Service) invalidateCache(ctx context.Context, ownerID int64) {
	prefix := fmt.Sprintf("stock_transfers:%d", ownerID)
	cacher.Of[[]domain.StockTransfer]().Invalidate(ctx, []string{prefix})
}
