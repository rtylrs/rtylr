package application

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/pos_settings"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheInvalidate "github.com/voxire/onion/internal/platform/cache/invalidate"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/paddle"
	"github.com/voxire/onion/internal/platform/pagination"
	"github.com/voxire/onion/internal/platform/wsbus"
)

const ordersCacheTTL = 30 * time.Second

const publicOrderTokenTTL = 30 * 24 * time.Hour

// ownerTopic builds the wsbus topic for an owner without the reflection and
// format-parsing overhead of fmt.Sprintf; it runs on every order mutation.
func ownerTopic(ownerID int64) string {
	return "owner:" + strconv.FormatInt(ownerID, 10)
}

func invalidateOrderScope(ctx context.Context, ownerID int64) {
	cacheInvalidate.InvalidateAsync(ctx, cacheKeys.InvalidateOrder, strconv.FormatInt(ownerID, 10))
}

// PaddlePaymentResult mirrors paddle.POSPaymentResult for the service surface.
type PaddlePaymentResult struct {
	TransactionID string
	CheckoutURL   string
}

type Service struct {
	repo     *infrastructure.Repository
	settings *pos_settings.Service
}

func NewService(ctx context.Context, db *sql.DB) *Service {
	return &Service{
		repo:     infrastructure.NewRepository(db),
		settings: pos_settings.NewService(ctx, db),
	}
}

func defaultOrderTypeForProfile(profile string) domain.OrderType {
	if profile == pos_settings.ProfileRestaurant {
		return domain.OrderTypeDineIn
	}
	return domain.OrderTypeTakeaway
}

func (s *Service) defaultOrderType(ctx context.Context, businessID int64) domain.OrderType {
	if s.settings == nil || businessID <= 0 {
		return domain.OrderTypeTakeaway
	}
	settings, err := s.settings.GetSettings(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Warn("failed to load POS settings for order type default",
			logger.Error(err),
			logger.Any("business_id", businessID),
		)
		return domain.OrderTypeTakeaway
	}
	return defaultOrderTypeForProfile(settings.Profile)
}

// ─── Locations ───────────────────────────────────────────────────────────────

func (s *Service) ListLocations(ctx context.Context, ownerID int64) ([]domain.Location, error) {
	return s.repo.ListLocations(ctx, ownerID)
}

func (s *Service) CreateLocation(ctx context.Context, ownerID int64, input domain.LocationInput) (int64, error) {
	if input.Name == "" {
		return 0, errors.New("location name is required")
	}

	return s.repo.CreateLocation(ctx, ownerID, input)
}

// ─── Tables ──────────────────────────────────────────────────────────────────

func (s *Service) ListTables(ctx context.Context, ownerID, locationID int64, status string) ([]domain.Table, error) {
	return s.repo.ListTables(ctx, ownerID, locationID, status)
}

func (s *Service) CreateTable(ctx context.Context, ownerID int64, input domain.TableInput) (int64, error) {
	if input.Name == "" {
		return 0, errors.New("table name is required")
	}

	if input.Capacity <= 0 {
		input.Capacity = 4
	}

	return s.repo.CreateTable(ctx, ownerID, input)
}

func (s *Service) UpdateTable(ctx context.Context, ownerID int64, input domain.TableInput) error {
	if input.ID == 0 {
		return errors.New("table id is required")
	}

	if input.Name == "" {
		return errors.New("table name is required")
	}

	return s.repo.UpdateTable(ctx, ownerID, input)
}

func (s *Service) UpdateTableStatus(ctx context.Context, ownerID, tableID int64, status domain.TableStatus) error {
	if tableID == 0 {
		return errors.New("table id is required")
	}

	if err := s.repo.UpdateTableStatus(ctx, ownerID, tableID, status); err != nil {
		return err
	}

	wsbus.Publish(wsbus.Event{
		Topic:    ownerTopic(ownerID),
		Type:     "pos:table:status",
		EntityID: tableID,
	})

	return nil
}

func (s *Service) DeleteTable(ctx context.Context, ownerID, tableID int64) error {
	if tableID == 0 {
		return errors.New("table id is required")
	}

	return s.repo.DeleteTable(ctx, ownerID, tableID)
}

// ─── Sessions ─────────────────────────────────────────────────────────────────

func (s *Service) GetCurrentOpenSession(ctx context.Context, ownerID, staffID int64) (*domain.Session, error) {
	return s.repo.GetCurrentOpenSession(ctx, ownerID, staffID)
}

func (s *Service) ListSessions(ctx context.Context, ownerID, locationID int64, dateFrom, dateTo string) ([]domain.Session, error) {
	return s.repo.ListSessions(ctx, ownerID, locationID, dateFrom, dateTo)
}

func (s *Service) OpenSession(ctx context.Context, ownerID, staffID int64, input domain.OpenSessionInput) (int64, error) {
	existing, err := s.repo.GetCurrentOpenSession(ctx, ownerID, staffID)
	if err != nil {
		return 0, err
	}

	if existing != nil {
		return 0, domain.ErrSessionAlreadyOpen
	}

	return s.repo.OpenSession(ctx, ownerID, staffID, input.LocationID, input.OpeningCash)
}

func (s *Service) CloseSession(ctx context.Context, ownerID, staffID int64, input domain.CloseSessionInput) error {
	sess, err := s.repo.GetSessionByID(ctx, ownerID, input.SessionID)
	if err != nil {
		return err
	}

	if sess.Status != domain.SessionStatusOpen {
		return domain.ErrSessionNotOpen
	}
	if sess.StaffID != staffID {
		return domain.ErrSessionNotFound
	}

	totalSales, cashSales, cardSales, onlineSales, totalOrders, err := s.repo.GetSessionTotals(ctx, input.SessionID)
	if err != nil {
		return err
	}

	cashMovements, refunds, err := s.repo.GetSessionCashAdjustments(ctx, ownerID, input.SessionID)
	if err != nil {
		return err
	}

	expectedCash := sess.OpeningCash + cashSales + cashMovements - refunds
	cashDiff := input.ClosingCash - expectedCash

	// Close and totals travel in one UPDATE so a failed close leaves the
	// session open and retryable; a separate total write after the close
	// committed could fail and strand a closed session with stale counters.
	return s.repo.CloseSessionForStaff(ctx, ownerID, staffID, input.SessionID,
		input.ClosingCash, expectedCash, cashDiff, input.Notes,
		totalSales, cashSales, cardSales, onlineSales, totalOrders)
}

// ─── Orders ──────────────────────────────────────────────────────────────────

func (s *Service) ListOrders(ctx context.Context, ownerID, sessionID int64, status, dateFrom, dateTo string) ([]domain.Order, error) {
	orders, _, _, err := s.ListOrdersPage(ctx, ownerID, sessionID, status, dateFrom, dateTo, "", 1, 500)
	return orders, err
}

var ordersPagePolicy = pagination.Policy{
	DefaultPageSize: 50,
	MaxPageSize:     500,
	MaxOffset:       10000,
}

func paginateOrders(page, pageSize int64) (int, int) {
	normalized := ordersPagePolicy.Normalize(page, pageSize)
	return int(normalized.Size), int(normalized.Offset)
}

func (s *Service) ListOrdersPage(ctx context.Context, ownerID, sessionID int64, status, dateFrom, dateTo, search string, page, pageSize int64) ([]domain.Order, int64, bool, error) {
	limit, offset := paginateOrders(page, pageSize)
	key := fmt.Sprintf("%s|sess=%d|st=%s|f=%s|t=%s|q=%s|p=%d|ps=%d",
		cacheKeys.OrdersList(ownerID), sessionID, status, dateFrom, dateTo, search, page, pageSize)
	result, err := cacher.Of[ordersListResult]().GetOrComputeSingle(ctx, key,
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
		cacheKeys.Order(ownerID, orderID),
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

	orderNumber, err := s.repo.GenerateOrderNumber(ctx, ownerID)
	if err != nil {
		return 0, err
	}

	orderID, err := s.repo.CreateOrder(ctx, ownerID, staffID, orderNumber, input)
	if err != nil {
		return 0, err
	}

	invalidateOrderScope(ctx, ownerID)

	wsbus.Publish(wsbus.Event{
		Topic:    ownerTopic(ownerID),
		Type:     "pos:order:created",
		EntityID: orderID,
	})

	return orderID, nil
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
		Topic:    ownerTopic(ownerID),
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
		Topic:    ownerTopic(ownerID),
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
