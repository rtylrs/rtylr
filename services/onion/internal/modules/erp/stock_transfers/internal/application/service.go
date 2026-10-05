package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/erp/stock_transfers/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const stockTransfersCacheTTL = 5 * time.Minute

type repository interface {
	ListStockTransfers(ctx context.Context, businessID int64, limit, offset int) ([]domain.StockTransfer, bool, error)
	GetByID(ctx context.Context, businessID, id int64) (domain.StockTransfer, error)
	CreateStockTransfer(ctx context.Context, businessID int64, input domain.StockTransferInput, events ...audit.Event) (int64, error)
	UpdateStockTransferStatus(ctx context.Context, businessID, id int64, status domain.StockTransferStatus, events ...audit.Event) error
}

type Service struct {
	repo repository
}

type stockTransfersPageResult struct {
	Items   []domain.StockTransfer
	HasMore bool
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
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
	cacheKey := cacheKeys.Scoped("stock_transfers", businessID, fmt.Sprintf("p=%d:ps=%d", page, pageSize)) + "|domain=v2"

	res, err := cacher.Of[stockTransfersPageResult]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) (stockTransfersPageResult, error) {
		items, hasMore, err := s.repo.ListStockTransfers(ctx, businessID, limit, offset)
		return stockTransfersPageResult{Items: items, HasMore: hasMore}, err
	}, cacher.WithExpiry(ctx, stockTransfersCacheTTL))
	if err != nil {
		return nil, false, err
	}
	return res.Items, res.HasMore, nil
}

func (s *Service) GetByID(ctx context.Context, businessID, id int64) (domain.StockTransfer, error) {
	return s.repo.GetByID(ctx, businessID, id)
}

func (s *Service) CreateStockTransfer(ctx context.Context, businessID int64, input domain.StockTransferInput) (int64, error) {
	if err := domain.ValidateStockTransferInput(input); err != nil {
		return 0, fmt.Errorf("%w: %w", domain.ErrInvalidStockTransfer, err)
	}
	id, err := s.repo.CreateStockTransfer(ctx, businessID, input, audit.Event{
		BusinessID: businessID,
		EntityType: "stock_transfer",
		Action:     "created",
		NewValues:  input,
	})
	if err != nil {
		return 0, err
	}

	s.invalidateCache(ctx, businessID)

	return id, nil
}

func (s *Service) UpdateStockTransferStatus(ctx context.Context, businessID, id int64, status domain.StockTransferStatus) error {
	old, err := s.repo.GetByID(ctx, businessID, id)
	if err != nil {
		return err
	}
	err = s.repo.UpdateStockTransferStatus(ctx, businessID, id, status, audit.Event{
		BusinessID: businessID,
		EntityType: "stock_transfer",
		EntityID:   id,
		Action:     "status_changed",
		OldValues:  old,
		NewValues:  map[string]domain.StockTransferStatus{"status": status},
	})
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, businessID)

	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	cacher.InvalidateTypedScope[stockTransfersPageResult](ctx, cacheKeys.ScopedPrefix("stock_transfers", businessID))
}
