package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/purchase_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/audit"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const poCacheTTL = 5 * time.Minute

type Service struct {
	repo *infrastructure.Repository
	db   *sql.DB
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db), db: db}
}

func (s *Service) GetByID(ctx context.Context, businessID, id int64) (domain.PurchaseOrder, error) {
	return s.repo.GetByID(ctx, businessID, id)
}

func (s *Service) ListPurchaseOrders(ctx context.Context, businessID int64, status string, page, pageSize int64) ([]domain.PurchaseOrder, bool, error) {
	limit, offset := paginate(page, pageSize)
	cacheKey := fmt.Sprintf("purchase_orders:%d:%s|p=%d|ps=%d", businessID, status, page, pageSize)

	rows, err := cacher.Of[[]domain.PurchaseOrder]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.PurchaseOrder, error) {
		return s.repo.ListPurchaseOrders(ctx, businessID, status, int32(limit+1), int32(offset)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
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
	if err := input.Validate(); err != nil {
		return 0, fmt.Errorf("%w: %w", domain.ErrInvalidPurchaseOrder, err)
	}

	id, err := s.repo.CreatePurchaseOrder(ctx, businessID, input, createdBy)
	if err != nil {
		return 0, err
	}
	if err := audit.Append(ctx, s.db, audit.Event{BusinessID: businessID, ActorID: createdBy, EntityType: "purchase_order", EntityID: id, Action: "created", NewValues: input}); err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, businessID)
	return id, nil
}

// ApprovePurchaseOrder transitions a purchase order when the caller has the
// app permission enforced at the API or tool boundary. It also enforces
// segregation of duties for every entry point.
func (s *Service) ApprovePurchaseOrder(ctx context.Context, businessID, poID, actorID int64) error {
	po, err := s.repo.GetByID(ctx, businessID, poID)
	if err != nil {
		return err
	}
	if po.CreatedBy == actorID {
		return domain.ErrPurchaseOrderSelfApproval
	}
	status := string(po.Status)

	if status == string(domain.POStatusApproved) {
		return domain.ErrPurchaseOrderAlreadyApproved
	}
	if status == string(domain.POStatusReceived) {
		return domain.ErrPurchaseOrderAlreadyReceived
	}
	old := map[string]string{"status": status}

	if err := s.repo.ApprovePurchaseOrder(ctx, businessID, poID); err != nil {
		return err
	}
	if err := audit.Append(ctx, s.db, audit.Event{BusinessID: businessID, ActorID: actorID, EntityType: "purchase_order", EntityID: poID, Action: "approved", OldValues: old, NewValues: map[string]string{"status": "approved"}}); err != nil {
		return err
	}

	s.invalidateCache(ctx, businessID)

	return nil
}

func (s *Service) ReceivePurchaseOrder(ctx context.Context, businessID, poID, receivedBy int64, submitted []domain.PurchaseOrderItem) error {
	old, err := s.repo.GetByID(ctx, businessID, poID)
	if err != nil {
		return err
	}
	if err := s.repo.ReceivePurchaseOrder(ctx, businessID, poID, receivedBy, submitted); err != nil {
		return err
	}
	if err := audit.Append(ctx, s.db, audit.Event{BusinessID: businessID, ActorID: receivedBy, EntityType: "purchase_order", EntityID: poID, Action: "received", OldValues: old, NewValues: map[string]string{"status": "received"}}); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) UpdatePurchaseOrderStatus(ctx context.Context, businessID, id int64, status string) error {
	oldStatus, err := s.repo.GetPurchaseOrderStatus(ctx, businessID, id)
	if err != nil {
		return err
	}
	// Receiving is special: it must verify ownership, increment stock for every
	// line item, and ledger the movement — all atomically and idempotently.
	// ReceivePurchaseOrder owns that whole transaction.
	if status == string(domain.POStatusReceived) {
		return s.ReceivePurchaseOrder(ctx, businessID, id, businessID, nil)
	}

	if err := s.repo.UpdatePurchaseOrderStatus(ctx, businessID, id, status); err != nil {
		return err
	}
	if err := audit.Append(ctx, s.db, audit.Event{BusinessID: businessID, ActorID: businessID, EntityType: "purchase_order", EntityID: id, Action: "status_changed", OldValues: map[string]string{"status": oldStatus}, NewValues: map[string]string{"status": status}}); err != nil {
		return err
	}

	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	cacher.InvalidateTypedScope[[]domain.PurchaseOrder](ctx, fmt.Sprintf("purchase_orders:%d:*", businessID))
}
