package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/purchase_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const poCacheTTL = 5 * time.Minute

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) GetByID(ctx context.Context, businessID, id int64) (domain.PurchaseOrder, error) {
	return s.repo.GetByID(ctx, businessID, id)
}

func (s *Service) ListPurchaseOrders(ctx context.Context, businessID int64, status string, page, pageSize int64) ([]domain.PurchaseOrder, bool, error) {
	limit, offset := paginate(page, pageSize)
	cacheKey := fmt.Sprintf("purchase_orders:%d:%s|p=%d|ps=%d", businessID, status, page, pageSize)

	rows, err := cacher.Of[[]domain.PurchaseOrder]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.PurchaseOrder, error) {
		return s.repo.ListPurchaseOrders(ctx, businessID, status, int32(limit+1), int32(offset))
	}, cacher.WithExpiry(ctx, poCacheTTL))
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	return rows, hasMore, nil
}

func paginate(page, pageSize int64) (int, int) {
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

func (s *Service) CreatePurchaseOrder(ctx context.Context, businessID int64, input domain.PurchaseOrderInput, createdBy int64) (int64, error) {
	id, err := s.repo.CreatePurchaseOrder(ctx, businessID, input, createdBy)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, businessID)
	return id, nil
}

// ApprovePurchaseOrder transitions a purchase order to the `approved` status.
// Returns domain.ErrPurchaseOrderNotFound when the PO doesn't exist for the caller,
// or domain.ErrPurchaseOrderAlreadyApproved when the PO is already approved.
func (s *Service) ApprovePurchaseOrder(ctx context.Context, businessID, poID int64) error {
	status, err := s.repo.GetPurchaseOrderStatus(ctx, businessID, poID)
	if err != nil {
		return err
	}

	if status == string(domain.POStatusApproved) {
		return domain.ErrPurchaseOrderAlreadyApproved
	}

	if err := s.repo.ApprovePurchaseOrder(ctx, businessID, poID); err != nil {
		return err
	}

	s.invalidateCache(ctx, businessID)

	return nil
}

func (s *Service) ReceivePurchaseOrder(ctx context.Context, businessID, poID, receivedBy int64) error {
	if err := s.repo.ReceivePurchaseOrder(ctx, businessID, poID, receivedBy); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) UpdatePurchaseOrderStatus(ctx context.Context, businessID, id int64, status string) error {
	// Receiving is special: it must verify ownership, increment stock for every
	// line item, and ledger the movement — all atomically and idempotently.
	// ReceivePurchaseOrder owns that whole transaction.
	if status == string(domain.POStatusReceived) {
		return s.ReceivePurchaseOrder(ctx, businessID, id, businessID)
	}

	if err := s.repo.UpdatePurchaseOrderStatus(ctx, businessID, id, status); err != nil {
		return err
	}

	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	cacher.Of[[]domain.PurchaseOrder]().DeleteByPattern(ctx, []string{fmt.Sprintf("purchase_orders:%d:*", businessID)})
}
