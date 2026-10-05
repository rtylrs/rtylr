package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/erp/stock_takes/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const stockTakesCacheTTL = 5 * time.Minute

type repository interface {
	ListStockTakes(ctx context.Context, businessID int64, limit, offset int) ([]domain.StockTake, bool, error)
	CreateStockTake(ctx context.Context, businessID int64, input domain.StockTakeInput, events ...audit.Event) (int64, error)
	UpdateStockTake(ctx context.Context, businessID int64, input domain.StockTakeUpdateInput, events ...audit.Event) error
}

type Service struct {
	repo repository
}

type stockTakesPageResult struct {
	Items   []domain.StockTake
	HasMore bool
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func paginateStockTakes(page, pageSize int64) (int, int) {
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

func (s *Service) ListStockTakes(ctx context.Context, businessID, page, pageSize int64) ([]domain.StockTake, bool, error) {
	limit, offset := paginateStockTakes(page, pageSize)
	cacheKey := cacheKeys.Scoped("stock_takes", businessID, fmt.Sprintf("p=%d:ps=%d", page, pageSize)) + "|domain=v2"

	res, err := cacher.Of[stockTakesPageResult]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) (stockTakesPageResult, error) {
		items, hasMore, err := s.repo.ListStockTakes(ctx, businessID, limit, offset)
		return stockTakesPageResult{Items: items, HasMore: hasMore}, err
	}, cacher.WithExpiry(ctx, stockTakesCacheTTL))
	if err != nil {
		return nil, false, err
	}
	return res.Items, res.HasMore, nil
}

func (s *Service) CreateStockTake(ctx context.Context, businessID int64, input domain.StockTakeInput) (int64, error) {
	id, err := s.repo.CreateStockTake(ctx, businessID, input, audit.Event{
		BusinessID: businessID,
		EntityType: "stock_take",
		Action:     "created",
		NewValues:  input,
	})
	if err != nil {
		return 0, err
	}

	s.invalidateCache(ctx, businessID)
	return id, nil
}

func (s *Service) UpdateStockTake(ctx context.Context, businessID int64, input domain.StockTakeUpdateInput) error {
	err := s.repo.UpdateStockTake(ctx, businessID, input, audit.Event{
		BusinessID: businessID,
		EntityType: "stock_take",
		EntityID:   input.ID,
		Action:     "updated",
		NewValues:  input,
	})
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	cacher.InvalidateTypedScope[stockTakesPageResult](ctx, cacheKeys.ScopedPrefix("stock_takes", businessID))
}
