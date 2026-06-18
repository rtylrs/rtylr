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

func (s *Service) GetByID(ctx context.Context, ownerID, id int64) (domain.PurchaseOrder, error) {
	return s.repo.GetByID(ctx, ownerID, id)
}

func (s *Service) ListPurchaseOrders(ctx context.Context, ownerID int64, status string, page, pageSize int64) ([]domain.PurchaseOrder, bool, error) {
	limit, offset := paginate(page, pageSize)
	cacheKey := fmt.Sprintf("purchase_orders:%d:%s|p=%d|ps=%d", ownerID, status, page, pageSize)

	rows, err := cacher.Of[[]domain.PurchaseOrder]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.PurchaseOrder, error) {
		return s.repo.ListPurchaseOrders(ctx, ownerID, status, int32(limit+1), int32(offset))
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

func (s *Service) CreatePurchaseOrder(ctx context.Context, ownerID int64, input domain.PurchaseOrderInput, createdBy int64) (int64, error) {
	id, err := s.repo.CreatePurchaseOrder(ctx, ownerID, input, createdBy)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, ownerID)
	return id, nil
}

// ApprovePurchaseOrder transitions a purchase order to the `approved` status.
// Returns domain.ErrPurchaseOrderNotFound when the PO doesn't exist for the caller,
// or domain.ErrPurchaseOrderAlreadyApproved when the PO is already approved.
func (s *Service) ApprovePurchaseOrder(ctx context.Context, ownerID, poID int64) error {
	status, err := s.repo.GetPurchaseOrderStatus(ctx, ownerID, poID)
	if err != nil {
		return err
	}

	if status == string(domain.POStatusApproved) {
		return domain.ErrPurchaseOrderAlreadyApproved
	}

	if err := s.repo.ApprovePurchaseOrder(ctx, ownerID, poID); err != nil {
		return err
	}

	s.invalidateCache(ctx, ownerID)

	return nil
}

func (s *Service) ReceivePurchaseOrder(ctx context.Context, ownerID, poID, receivedBy int64) error {
	if err := s.repo.ReceivePurchaseOrder(ctx, ownerID, poID, receivedBy); err != nil {
		return err
	}
	s.invalidateCache(ctx, ownerID)
	return nil
}

func (s *Service) UpdatePurchaseOrderStatus(ctx context.Context, ownerID, id int64, status string) error {
	// Receiving is special: it must verify ownership, increment stock for every
	// line item, and ledger the movement — all atomically and idempotently.
	// ReceivePurchaseOrder owns that whole transaction.
	if status == string(domain.POStatusReceived) {
		if err := s.ReceivePurchaseOrder(ctx, ownerID, id, ownerID); err != nil {
			return err
		}
		return nil
	}

	if err := s.repo.UpdatePurchaseOrderStatus(ctx, ownerID, id, status); err != nil {
		return err
	}

	s.invalidateCache(ctx, ownerID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, ownerID int64) {
	prefix := fmt.Sprintf("purchase_orders:%d:", ownerID)
	cacher.Of[[]domain.PurchaseOrder]().Invalidate(ctx, []string{prefix})
}
