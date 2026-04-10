package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
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

	return s.repo.UpdateTableStatus(ctx, ownerID, tableID, status)
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

func (s *Service) CloseSession(ctx context.Context, ownerID int64, input domain.CloseSessionInput) error {
	sess, err := s.repo.GetSessionByID(ctx, ownerID, input.SessionID)
	if err != nil {
		return err
	}

	if sess.Status != domain.SessionStatusOpen {
		return domain.ErrSessionNotOpen
	}

	totalSales, cashSales, cardSales, onlineSales, totalOrders, err := s.repo.GetSessionTotals(ctx, input.SessionID)
	if err != nil {
		return err
	}

	expectedCash := sess.OpeningCash + cashSales
	cashDiff := input.ClosingCash - expectedCash

	if err := s.repo.CloseSession(ctx, ownerID, input.SessionID, input.ClosingCash, expectedCash, cashDiff, input.Notes); err != nil {
		return err
	}

	return s.repo.UpdateSessionTotals(ctx, input.SessionID, totalSales, cashSales, cardSales, onlineSales, totalOrders)
}

// ─── Orders ──────────────────────────────────────────────────────────────────

func (s *Service) ListOrders(ctx context.Context, ownerID, sessionID int64, status, dateFrom, dateTo string) ([]domain.Order, error) {
	return s.repo.ListOrders(ctx, ownerID, sessionID, status, dateFrom, dateTo)
}

func (s *Service) GetOrder(ctx context.Context, ownerID, orderID int64) (*domain.Order, error) {
	return s.repo.GetOrder(ctx, ownerID, orderID)
}

func (s *Service) CreateOrder(ctx context.Context, ownerID, staffID int64, input domain.OrderInput) (int64, error) {
	if len(input.Items) == 0 {
		return 0, errors.New("order must have at least one item")
	}

	if input.Type == "" {
		input.Type = domain.OrderTypeDineIn
	}

	if input.PaymentMethod == "" {
		input.PaymentMethod = domain.PaymentMethodCash
	}

	orderNumber, err := s.repo.GenerateOrderNumber(ctx, ownerID)
	if err != nil {
		return 0, err
	}

	return s.repo.CreateOrder(ctx, ownerID, staffID, orderNumber, input)
}

func (s *Service) UpdateOrderStatus(ctx context.Context, ownerID, orderID int64, status domain.OrderStatus) error {
	return s.repo.UpdateOrderStatus(ctx, ownerID, orderID, status)
}

func (s *Service) UpdateOrderItemStatus(ctx context.Context, orderID, itemID int64, status domain.ItemStatus) error {
	return s.repo.UpdateOrderItemStatus(ctx, orderID, itemID, status)
}

func (s *Service) ProcessPayment(ctx context.Context, ownerID int64, input domain.PaymentInput) error {
	if input.OrderID == 0 {
		return errors.New("order id is required")
	}

	if err := s.repo.ProcessPayment(ctx, ownerID, input.OrderID, input); err != nil {
		return err
	}

	sessionID, err := s.repo.GetOrderSessionID(ctx, input.OrderID)
	if err != nil || sessionID == 0 {
		return nil
	}

	totalSales, cashSales, cardSales, onlineSales, totalOrders, err := s.repo.GetSessionTotals(ctx, sessionID)
	if err != nil {
		return nil
	}

	return s.repo.UpdateSessionTotals(ctx, sessionID, totalSales, cashSales, cardSales, onlineSales, totalOrders)
}

func (s *Service) ProcessRefund(ctx context.Context, ownerID, refundedBy int64, input domain.RefundInput) (int64, error) {
	if input.OrderID == 0 {
		return 0, errors.New("order id is required")
	}

	if input.Amount <= 0 {
		return 0, errors.New("refund amount must be positive")
	}

	refundID, err := s.repo.CreateRefund(ctx, ownerID, input.OrderID, refundedBy, input.Amount, input.Reason)
	if err != nil {
		return 0, err
	}

	sessionID, err := s.repo.GetOrderSessionID(ctx, input.OrderID)
	if err == nil && sessionID > 0 {
		_ = s.repo.UpdateSessionRefundTotals(ctx, sessionID, input.Amount)
	}

	return refundID, nil
}

// ─── Split / Discount / Void / Hold ─────────────────────────────────────────

func (s *Service) SplitOrder(ctx context.Context, ownerID, orderID int64, splits [][]int64) ([]int64, error) {
	return s.repo.SplitOrder(ctx, ownerID, orderID, splits)
}

func (s *Service) ApplyDiscount(ctx context.Context, ownerID int64, input domain.DiscountInput) error {
	return s.repo.ApplyDiscount(ctx, ownerID, input)
}

func (s *Service) VoidOrder(ctx context.Context, ownerID, orderID int64, reason string) error {
	return s.repo.VoidOrder(ctx, ownerID, orderID, reason)
}

func (s *Service) HoldOrder(ctx context.Context, ownerID, orderID int64, reason string) error {
	return s.repo.HoldOrder(ctx, ownerID, orderID, reason)
}

func (s *Service) ListHeldOrders(ctx context.Context, ownerID int64) ([]domain.Order, error) {
	return s.repo.ListHeldOrders(ctx, ownerID)
}

// ─── Dashboard ───────────────────────────────────────────────────────────────

func (s *Service) GetDashboardStats(ctx context.Context, ownerID int64) (domain.DashboardStats, error) {
	return s.repo.GetDashboardStats(ctx, ownerID)
}
