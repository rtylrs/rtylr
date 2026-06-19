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
	ListStockTransfers(ctx context.Context, businessID int64, limit, offset int) ([]domain.StockTransfer, bool, error)
	GetByID(ctx context.Context, businessID, id int64) (domain.StockTransfer, error)
	CreateStockTransfer(ctx context.Context, businessID int64, input domain.StockTransferInput) (int64, error)
	UpdateStockTransferStatus(ctx context.Context, businessID, id int64, status domain.StockTransferStatus) error
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

func (s *Service) ListStockTransfers(ctx context.Context, businessID, page, pageSize int64) ([]domain.StockTransfer, bool, error) {
	limit, offset := paginateStockTransfers(page, pageSize)
	cacheKey := fmt.Sprintf("stock_transfers:%d|p=%d|ps=%d", businessID, page, pageSize)

	type pageResult struct {
		items   []domain.StockTransfer
		hasMore bool
	}

	res, err := cacher.Of[pageResult]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) (pageResult, error) {
		items, hasMore, err := s.repo.ListStockTransfers(ctx, businessID, limit, offset)
		return pageResult{items: items, hasMore: hasMore}, err
	}, cacher.WithExpiry(ctx, stockTransfersCacheTTL))
	if err != nil {
		return nil, false, err
	}
	return res.items, res.hasMore, nil
}

func (s *Service) GetByID(ctx context.Context, businessID, id int64) (domain.StockTransfer, error) {
	return s.repo.GetByID(ctx, businessID, id)
}

func (s *Service) CreateStockTransfer(ctx context.Context, businessID int64, input domain.StockTransferInput) (int64, error) {
	id, err := s.repo.CreateStockTransfer(ctx, businessID, input)
	if err != nil {
		return 0, err
	}

	s.invalidateCache(ctx, businessID)

	return id, nil
}

func (s *Service) UpdateStockTransferStatus(ctx context.Context, businessID, id int64, status domain.StockTransferStatus) error {
	err := s.repo.UpdateStockTransferStatus(ctx, businessID, id, status)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, businessID)

	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	prefix := fmt.Sprintf("stock_transfers:%d", businessID)
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, prefix)
	}
}
