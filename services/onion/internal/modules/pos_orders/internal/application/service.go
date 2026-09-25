package application

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/pos_settings"
	cacheInvalidate "github.com/voxire/onion/internal/platform/cache/invalidate"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
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
