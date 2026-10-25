package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/wsbus"
)

func paginateOrders(page, pageSize int64) (int, int) {
	normalized := ordersPagePolicy.Normalize(page, pageSize)
	return int(normalized.Size), int(normalized.Offset)
}

func (s *Service) ListOrdersPage(ctx context.Context, ownerID, sessionID int64, status, dateFrom, dateTo, search string, page, pageSize int64) ([]domain.Order, int64, bool, error) {
	limit, offset := paginateOrders(page, pageSize)
	key := fmt.Sprintf("%s|sess=%d|st=%s|f=%s|t=%s|q=%s|p=%d|ps=%d",
		cacheKeys.OrdersList(ownerID), sessionID, status, dateFrom, dateTo, search, page, pageSize)
	result, err := cacher.Of[ordersListResult]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (ordersListResult, error) {
			orders, total, hasMore, err := s.repo.ListOrders(ctx, ownerID, sessionID, status, dateFrom, dateTo, search, limit, offset)
			return ordersListResult{Orders: orders, Total: total, HasMore: hasMore}, err
		},
		cacher.WithExpiry(ctx, ordersCacheTTL),
	)
	return result.Orders, result.Total, result.HasMore, err
}

type ordersListResult struct {
	Orders  []domain.Order
	Total   int64
	HasMore bool
}

func (s *Service) GetOrder(ctx context.Context, ownerID, orderID int64) (*domain.Order, error) {
	return cacher.Of[*domain.Order]().GetOrComputeSingle(ctx,
		cacheKeys.Order(ownerID, orderID)+"|domain=v2",
		func(ctx context.Context) (*domain.Order, error) {
			return s.repo.GetOrder(ctx, ownerID, orderID)
		},
		cacher.WithExpiry(ctx, ordersCacheTTL),
	)
}

func (s *Service) CreateOrder(ctx context.Context, businessID, ownerID, staffID int64, input domain.OrderInput) (int64, error) {
	if len(input.Items) == 0 {
		return 0, errors.New("order must have at least one item")
	}

	if input.Type == "" {
		input.Type = s.defaultOrderType(ctx, businessID)
	}

	if input.PaymentMethod == "" {
		input.PaymentMethod = domain.PaymentMethodCash
	}

	// The order number is generated inside the insert transaction, not here:
	// generating it on a separate connection let two concurrent sales pick the
	// same number, and the loser of uq_pos_orders_number lost the whole create.
	orderID, err := s.repo.CreateOrder(ctx, ownerID, staffID, "", input)
	if err != nil {
		return 0, err
	}

	invalidateOrderScope(ctx, ownerID)

	wsbus.Publish(wsbus.Event{
		Topic:    businessTopic(ownerID),
		Type:     "pos:order:created",
		EntityID: orderID,
	})

	return orderID, nil
}

// ResumeOrder atomically replaces the lines of an editable held/pending order
// and returns it to the pending workflow. Paid, refunded, voided, and
// kitchen-started orders are rejected by the repository before any write.
func (s *Service) ResumeOrder(ctx context.Context, ownerID, staffID, orderID int64, input domain.OrderInput) error {
	if err := s.repo.ResumeOrder(ctx, ownerID, staffID, orderID, input); err != nil {
		return err
	}
	invalidateOrderScope(ctx, ownerID)
	wsbus.Publish(wsbus.Event{
		Topic:    businessTopic(ownerID),
		Type:     "pos:order:resumed",
		EntityID: orderID,
	})
	return nil
}

func (s *Service) UpdateOrderStatus(ctx context.Context, ownerID, orderID int64, status domain.OrderStatus) error {
	if !domain.IsOrderStatus(status) {
		return domain.ErrInvalidOrderStatus
	}
	if !domain.CanSetOrderStatusDirectly(status) {
		// paid/refunded/voided carry companion writes; writing the status
		// alone would leave payment_status, refunds, or the void trail
		// inconsistent.
		return domain.ErrOrderStatusRequiresWorkflow
	}
	if err := s.repo.UpdateOrderStatus(ctx, ownerID, orderID, status); err != nil {
		return err
	}
	invalidateOrderScope(ctx, ownerID)
	return nil
}

func (s *Service) UpdateOrderItemStatus(ctx context.Context, ownerID, orderID, itemID int64, status domain.ItemStatus) error {
	if !domain.IsItemStatus(status) {
		return domain.ErrInvalidItemStatus
	}
	// The underlying item-status query isn't owner-scoped, so verify the order
	// belongs to this owner first — otherwise a caller could flip KDS item
	// status on another tenant's order.
	if _, err := s.repo.GetOrder(ctx, ownerID, orderID); err != nil {
		return err
	}
	if err := s.repo.UpdateOrderItemStatus(ctx, orderID, itemID, status); err != nil {
		return err
	}

	wsbus.Publish(wsbus.Event{
		Topic:    businessTopic(ownerID),
		Type:     "pos:kds:item:updated",
		EntityID: itemID,
	})

	return nil
}

type paymentSessionTotalsRefresher interface {
	GetOrderSessionID(context.Context, int64) (int64, error)
	GetSessionTotals(context.Context, int64) (float64, float64, float64, float64, int, error)
	UpdateSessionTotals(context.Context, int64, float64, float64, float64, float64, int) error
}

// refreshSessionTotalsAfterPayment is deliberately best effort. The payment
// transaction has already committed when this runs, so a totals refresh must
// not turn a durable payment into a failed request.
func refreshSessionTotalsAfterPayment(ctx context.Context, repo paymentSessionTotalsRefresher, orderID int64) {
	sessionID, err := repo.GetOrderSessionID(ctx, orderID)
	if err != nil {
		logger.Ctx(ctx).Warn("payment recorded but session lookup failed",
			logger.Any("order_id", orderID),
			logger.Error(err),
		)
		return
	}
	if sessionID == 0 {
		return
	}

	totalSales, cashSales, cardSales, onlineSales, totalOrders, err := repo.GetSessionTotals(ctx, sessionID)
	if err != nil {
		logger.Ctx(ctx).Warn("payment recorded but session totals read failed",
			logger.Any("order_id", orderID),
			logger.Any("session_id", sessionID),
			logger.Error(err),
		)
		return
	}

	if err := repo.UpdateSessionTotals(ctx, sessionID, totalSales, cashSales, cardSales, onlineSales, totalOrders); err != nil {
		logger.Ctx(ctx).Warn("payment recorded but session totals update failed",
			logger.Any("order_id", orderID),
			logger.Any("session_id", sessionID),
			logger.Error(err),
		)
	}
}

func (s *Service) ProcessPayment(ctx context.Context, ownerID int64, input domain.PaymentInput) error {
	if input.OrderID == 0 {
		return errors.New("order id is required")
	}

	if err := s.repo.ProcessPayment(ctx, ownerID, input.OrderID, input); err != nil {
		return err
	}

	wsbus.Publish(wsbus.Event{
		Topic:    businessTopic(ownerID),
		Type:     "pos:order:paid",
		EntityID: input.OrderID,
	})

	invalidateOrderScope(ctx, ownerID)

	refreshSessionTotalsAfterPayment(ctx, s.repo, input.OrderID)
	return nil
}

func (s *Service) ProcessRefund(ctx context.Context, ownerID, refundedBy int64, input domain.RefundInput) (int64, error) {
	if input.OrderID == 0 {
		return 0, errors.New("order id is required")
	}

	if input.Amount <= 0 {
		return 0, errors.New("refund amount must be positive")
	}

	// Confirm the order is this owner's before refunding — GetOrderSessionID and
	// UpdateSessionRefundTotals below act on the order/session by id with no
	// owner scope, so without this a refund could bump another tenant's session.
	if _, err := s.repo.GetOrder(ctx, ownerID, input.OrderID); err != nil {
		return 0, err
	}

	refundID, err := s.repo.CreateRefund(ctx, ownerID, input.OrderID, refundedBy, input.Amount, input.Reason)
	if err != nil {
		return 0, err
	}

	invalidateOrderScope(ctx, ownerID)

	sessionID, err := s.repo.GetOrderSessionID(ctx, input.OrderID)
	if err == nil && sessionID > 0 {
		if updateErr := s.repo.UpdateSessionRefundTotals(ctx, sessionID, input.Amount); updateErr != nil {
			logger.Ctx(ctx).Warn("refund recorded but session totals update failed",
				logger.Any("session_id", sessionID),
				logger.Any("order_id", input.OrderID),
				logger.Any("refund_id", refundID),
				logger.Float64("amount", input.Amount),
				logger.Error(updateErr),
			)
		}
	}

	return refundID, nil
}

// ─── Split / Discount / Void / Hold ─────────────────────────────────────────
