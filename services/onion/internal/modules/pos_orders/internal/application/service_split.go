package application

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/paddle"
	"github.com/voxire/onion/internal/platform/wsbus"
)

func (s *Service) SplitOrder(ctx context.Context, ownerID, orderID int64, splits [][]int64) ([]int64, error) {
	ids, err := s.repo.SplitOrder(ctx, ownerID, orderID, splits)
	if err != nil {
		return nil, err
	}
	invalidateOrderScope(ctx, ownerID)
	return ids, nil
}

func (s *Service) ApplyDiscount(ctx context.Context, ownerID, staffID int64, input domain.DiscountInput) error {
	if err := domain.ValidateDiscountInput(input); err != nil {
		return err
	}
	if err := s.repo.ApplyDiscount(ctx, ownerID, staffID, input); err != nil {
		return err
	}
	invalidateOrderScope(ctx, ownerID)
	return nil
}

func (s *Service) VoidOrder(ctx context.Context, ownerID, orderID int64, reason string) error {
	if err := s.repo.VoidOrder(ctx, ownerID, orderID, reason); err != nil {
		return err
	}

	wsbus.Publish(wsbus.Event{
		Topic:    ownerTopic(ownerID),
		Type:     "pos:order:voided",
		EntityID: orderID,
	})

	invalidateOrderScope(ctx, ownerID)
	return nil
}

func (s *Service) HoldOrder(ctx context.Context, ownerID, orderID int64, reason string) error {
	if err := s.repo.HoldOrder(ctx, ownerID, orderID, reason); err != nil {
		return err
	}
	invalidateOrderScope(ctx, ownerID)
	return nil
}

func (s *Service) ListHeldOrders(ctx context.Context, ownerID int64) ([]domain.Order, error) {
	key := fmt.Sprintf("%s|held", cacheKeys.OrdersList(ownerID))
	return cacher.Of[[]domain.Order]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.Order, error) {
			return s.repo.ListHeldOrders(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, ordersCacheTTL),
	)
}

func (s *Service) SendToKitchen(ctx context.Context, ownerID, orderID int64) error {
	order, err := s.repo.GetOrder(ctx, ownerID, orderID)
	if err != nil {
		return err
	}
	if order.Status != domain.OrderStatusPending && order.Status != domain.OrderStatusConfirmed {
		return errors.New("order is not in a sendable state")
	}
	if err := s.repo.UpdateOrderStatus(ctx, ownerID, orderID, domain.OrderStatusPreparing); err != nil {
		return err
	}
	invalidateOrderScope(ctx, ownerID)
	wsbus.Publish(wsbus.Event{
		Topic:    ownerTopic(ownerID),
		Type:     "pos:order:preparing",
		EntityID: orderID,
	})
	return nil
}

func (s *Service) VoidOrderItem(ctx context.Context, ownerID, orderID, itemID int64) error {
	// Verify the order belongs to this owner before touching its items — the
	// item-level cancel isn't owner-scoped, so without this a caller could void
	// an item on another tenant's order. Mirrors TransferTable's guard.
	if _, err := s.repo.GetOrder(ctx, ownerID, orderID); err != nil {
		return err
	}
	if err := s.repo.VoidOrderItem(ctx, ownerID, orderID, itemID); err != nil {
		return err
	}
	invalidateOrderScope(ctx, ownerID)
	return nil
}

func (s *Service) TransferTable(ctx context.Context, ownerID, orderID, newTableID int64) error {
	order, err := s.repo.GetOrder(ctx, ownerID, orderID)
	if err != nil {
		return err
	}
	if order.Status == domain.OrderStatusPaid || order.Status == domain.OrderStatusVoided || order.Status == domain.OrderStatusCancelled {
		return domain.ErrOrderCannotBeTransferred
	}
	if err := s.repo.TransferTable(ctx, ownerID, orderID, order.TableID, newTableID); err != nil {
		return err
	}
	invalidateOrderScope(ctx, ownerID)
	return nil
}

// ─── Paddle billing ─────────────────────────────────────────────────────────

// CreatePaddlePayment creates a Paddle transaction for the given order and
// persists the returned transaction id on the order row. The actual transition
// to status=paid happens when the matching transaction.completed webhook
// arrives — not here.
func (s *Service) CreatePaddlePayment(ctx context.Context, ownerID, orderID int64, customerEmail, currency string) (PaddlePaymentResult, error) {
	order, err := s.repo.GetOrder(ctx, ownerID, orderID)
	if err != nil {
		return PaddlePaymentResult{}, err
	}
	if order == nil {
		return PaddlePaymentResult{}, domain.ErrOrderNotFound
	}
	if order.PaymentStatus == domain.PaymentStatus("paid") {
		return PaddlePaymentResult{}, errors.New("order already paid")
	}

	amountMinor := int64(math.Round(order.Total * 100))
	if amountMinor <= 0 {
		return PaddlePaymentResult{}, errors.New("order total must be positive")
	}
	if currency == "" {
		currency = "USD"
	}

	// Bind the idempotency key to the full request body: a static per-order
	// key would make Paddle reject (or replay) a legitimately re-issued
	// checkout after the amount, currency, or email changed, while retries
	// of an identical body still dedupe.
	result, err := paddle.CreatePOSTransaction(ctx, paddle.POSPaymentRequest{
		OwnerID:        ownerID,
		OrderID:        orderID,
		AmountMinor:    amountMinor,
		Currency:       currency,
		CustomerEmail:  customerEmail,
		Description:    "POS order " + order.OrderNumber,
		IdempotencyKey: fmt.Sprintf("order:%d:%d:%s:%s:%s", orderID, ownerID, currency, customerEmail, strconv.FormatInt(amountMinor, 10)),
	})
	if err != nil {
		return PaddlePaymentResult{}, err
	}

	if setErr := s.repo.SetPaddleTransactionID(ctx, ownerID, orderID, result.TransactionID); setErr != nil {
		return PaddlePaymentResult{}, setErr
	}

	return PaddlePaymentResult{
		TransactionID: result.TransactionID,
		CheckoutURL:   result.CheckoutURL,
	}, nil
}

// HandlePaddleTransactionCompleted marks the matching order as paid. Uses the
// pos_paddle_events_processed table to dedupe repeat deliveries.
func (s *Service) HandlePaddleTransactionCompleted(ctx context.Context, txID, eventType string) error {
	if txID == "" {
		return errors.New("paddle transaction id required")
	}

	processed, err := s.repo.IsPaddleEventProcessed(ctx, txID)
	if err != nil {
		return err
	}
	if processed {
		return nil
	}

	_, ownerID, getErr := s.repo.GetOrderIDByPaddleTx(ctx, txID)
	if getErr != nil {
		return getErr
	}

	if _, markErr := s.repo.MarkOrderPaidByPaddleTx(ctx, txID); markErr != nil {
		return markErr
	}

	return s.repo.RecordPaddleEventProcessed(ctx, txID, eventType, ownerID)
}

// ─── Dashboard ───────────────────────────────────────────────────────────────

func (s *Service) GetDashboardStats(ctx context.Context, ownerID int64) (domain.DashboardStats, error) {
	key := fmt.Sprintf("%s|dashboard", cacheKeys.OrdersList(ownerID))
	return cacher.Of[domain.DashboardStats]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.DashboardStats, error) {
			return s.repo.GetDashboardStats(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, ordersCacheTTL),
	)
}

// ─── Public Online Ordering ──────────────────────────────────────────────────

type PublicOrderItemInput struct {
	MenuItemID int64
	Name       string
	Quantity   float64
	UnitPrice  float64
	Notes      string
}

type PublicOrderInput struct {
	CustomerName  string
	CustomerPhone string
	CustomerEmail string
	Notes         string
	Items         []PublicOrderItemInput
}

func (s *Service) CreatePublicOrder(ctx context.Context, ownerID int64, input PublicOrderInput) (int64, string, string, error) {
	if ownerID == 0 {
		return 0, "", "", errors.New("owner_id is required")
	}
	if len(input.Items) == 0 {
		return 0, "", "", errors.New("order must have at least one item")
	}
	for i, item := range input.Items {
		if item.MenuItemID <= 0 {
			return 0, "", "", fmt.Errorf("item[%d]: menu_item_id is required", i)
		}
		if item.Quantity <= 0 {
			return 0, "", "", fmt.Errorf("item[%d]: quantity must be positive", i)
		}
	}

	enabled, err := s.repo.IsOnlineOrderingEnabled(ctx, ownerID)
	if err != nil {
		return 0, "", "", fmt.Errorf("check online ordering entitlement: %w", err)
	}
	if !enabled {
		return 0, "", "", errors.New("online ordering is not enabled")
	}

	orderNumber, err := s.repo.GenerateOrderNumber(ctx, ownerID)
	if err != nil {
		return 0, "", "", fmt.Errorf("generate order number: %w", err)
	}

	items := make([]domain.OrderItemInput, 0, len(input.Items))
	for i, it := range input.Items {
		item, err := s.repo.GetPublicMenuItem(ctx, ownerID, it.MenuItemID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, "", "", fmt.Errorf("item[%d]: menu item is not available", i)
			}
			return 0, "", "", fmt.Errorf("load menu item: %w", err)
		}
		item.Quantity = it.Quantity
		item.Notes = it.Notes
		items = append(items, item)
	}

	orderID, err := s.repo.CreatePublicOrder(ctx, ownerID, orderNumber, input.CustomerName, input.CustomerPhone, input.Notes, items)
	if err != nil {
		return 0, "", "", fmt.Errorf("create public order: %w", err)
	}

	publicToken := uuid.NewString()
	cacher.Of[string]().Put(ctx, publicOrderTokenCacheKey(ownerID, orderID), publicToken,
		cacher.WithExpiry(ctx, publicOrderTokenTTL),
	)

	return orderID, orderNumber, publicToken, nil
}

func (s *Service) GetPublicOrderStatus(ctx context.Context, businessID, orderID int64, publicToken string) (string, string, error) {
	if businessID <= 0 {
		return "", "", errors.New("business_id is required")
	}

	cachedToken, found := cacher.Of[string]().Get(ctx, publicOrderTokenCacheKey(businessID, orderID))
	if !found || subtle.ConstantTimeCompare([]byte(cachedToken), []byte(publicToken)) != 1 {
		return "", "", sql.ErrNoRows
	}

	return s.repo.GetPublicOrderStatus(ctx, businessID, orderID)
}

func publicOrderTokenCacheKey(businessID, orderID int64) string {
	return fmt.Sprintf("pos:public-order-token:%d:%d", businessID, orderID)
}
