package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/stock_transfers/internal/domain"
	"github.com/voxire/onion/internal/modules/stock_transfers/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const stockTransfersCacheTTL = 5 * time.Minute

type repository interface {
	ListStockTransfers(ctx context.Context, ownerID int64, limit, offset int) ([]domain.StockTransfer, bool, error)
	CreateStockTransfer(ctx context.Context, ownerID int64, input domain.StockTransferInput) (int64, error)
	UpdateStockTransferStatus(ctx context.Context, ownerID, id int64, status domain.StockTransferStatus) error
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

func paginateStockTransfers(page, pageSize int64) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}
	return int(pageSize), int((page - 1) * pageSize)
}

func (s *Service) ListStockTransfers(ctx context.Context, ownerID, page, pageSize int64) ([]domain.StockTransfer, bool, error) {
	limit, offset := paginateStockTransfers(page, pageSize)
	cacheKey := fmt.Sprintf("stock_transfers:%d|p=%d|ps=%d", ownerID, page, pageSize)

	type pageResult struct {
		items   []domain.StockTransfer
		hasMore bool
	}

	res, err := cacher.Of[pageResult]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) (pageResult, error) {
		items, hasMore, err := s.repo.ListStockTransfers(ctx, ownerID, limit, offset)
		return pageResult{items: items, hasMore: hasMore}, err
	}, cacher.WithExpiry(ctx, stockTransfersCacheTTL))
	if err != nil {
		return nil, false, err
	}
	return res.items, res.hasMore, nil
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
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, prefix)
	}
}
