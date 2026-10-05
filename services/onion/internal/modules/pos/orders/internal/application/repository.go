package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/domain"
)

type repository interface {
	ApplyDiscount(ctx context.Context, businessID, staffID int64, input domain.DiscountInput) error
	VoidOrder(ctx context.Context, businessID, orderID int64, reason string) error
	HoldOrder(ctx context.Context, businessID, orderID int64, reason string) error
	ListHeldOrders(ctx context.Context, businessID int64) ([]domain.Order, error)
	SetPaddleTransactionID(ctx context.Context, businessID, orderID int64, txID string) error
	MarkOrderPaidByPaddleTx(ctx context.Context, txID string) (int64, error)
	IsPaddleEventProcessed(ctx context.Context, txID string) (bool, error)
	RecordPaddleEventProcessed(ctx context.Context, txID, eventType string, ownerID int64) error
	GetOrderIDByPaddleTx(ctx context.Context, txID string) (int64, int64, error)
	IsOnlineOrderingEnabled(ctx context.Context, businessID int64) (bool, error)
	GetPublicMenuItem(ctx context.Context, businessID, menuItemID int64) (domain.OrderItemInput, error)
	CreatePublicOrder(ctx context.Context, businessID int64, orderNumber, customerName, customerPhone, notes string, items []domain.OrderItemInput) (int64, string, error)
	GetPublicOrderStatus(ctx context.Context, businessID, orderID int64) (orderNumber, status string, err error)
	ListLocations(ctx context.Context, businessID int64) ([]domain.Location, error)
	CreateLocation(ctx context.Context, ownerID int64, input domain.LocationInput) (int64, error)
	ListTables(ctx context.Context, businessID, locationID int64, status string) ([]domain.Table, error)
	CreateTable(ctx context.Context, businessID int64, input domain.TableInput) (int64, error)
	UpdateTable(ctx context.Context, businessID int64, input domain.TableInput) error
	UpdateTableStatus(ctx context.Context, businessID, tableID int64, status domain.TableStatus) error
	DeleteTable(ctx context.Context, businessID, tableID int64) error
	GetCurrentOpenSession(ctx context.Context, businessID, staffID int64) (*domain.Session, error)
	ListSessions(ctx context.Context, businessID, locationID int64, dateFrom, dateTo string) ([]domain.Session, error)
	OpenSession(ctx context.Context, businessID, staffID, locationID int64, openingCash float64) (int64, error)
	CloseSessionForStaff(ctx context.Context, businessID, staffID, sessionID int64, closingCash, expectedCash, cashDiff float64, notes string, totalSales, cashSales, cardSales, onlineSales float64, totalOrders int) error
	GetSessionTotals(ctx context.Context, sessionID int64) (totalSales, cashSales, cardSales, onlineSales float64, totalOrders int, err error)
	GetSessionCashAdjustments(ctx context.Context, businessID, sessionID int64) (netMovement, refunds float64, err error)
	UpdateSessionTotals(ctx context.Context, sessionID int64, totalSales, cashSales, cardSales, onlineSales float64, totalOrders int) error
	GetSessionByID(ctx context.Context, businessID, sessionID int64) (*domain.Session, error)
	UpdateSessionRefundTotals(ctx context.Context, sessionID int64, refundAmount float64) error
	CreateRefund(ctx context.Context, businessID, orderID, refundedBy int64, amount float64, reason string) (int64, error)
	VoidOrderItem(ctx context.Context, businessID, orderID, itemID int64) error
	TransferTable(ctx context.Context, businessID, orderID, oldTableID, newTableID int64) error
	SplitOrder(ctx context.Context, businessID, orderID int64, splits [][]int64) ([]int64, error)
	ResumeOrder(ctx context.Context, businessID, staffID, orderID int64, input domain.OrderInput) error
	GetDashboardStats(ctx context.Context, businessID int64) (domain.DashboardStats, error)
	GetOrderSessionID(ctx context.Context, orderID int64) (int64, error)
	ListOrders(ctx context.Context, businessID, sessionID int64, status, dateFrom, dateTo, search string, limit, offset int) ([]domain.Order, int64, bool, error)
	GetOrder(ctx context.Context, businessID, orderID int64) (*domain.Order, error)
	CreateOrder(ctx context.Context, businessID, staffID int64, orderNumber string, input domain.OrderInput) (int64, error)
	UpdateOrderStatus(ctx context.Context, businessID, orderID int64, status domain.OrderStatus) error
	UpdateOrderItemStatus(ctx context.Context, orderID, itemID int64, status domain.ItemStatus) error
	ProcessPayment(ctx context.Context, businessID, orderID int64, input domain.PaymentInput) error
}

type settingsReader interface {
	GetProfile(ctx context.Context, businessID int64) (string, error)
}
