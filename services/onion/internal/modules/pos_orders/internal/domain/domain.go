package domain

import (
	"errors"
	"time"
)

var (
	ErrSessionAlreadyOpen       = errors.New("staff already has an open session")
	ErrSessionNotFound          = errors.New("session not found")
	ErrSessionNotOpen           = errors.New("session is not open")
	ErrOrderNotFound            = errors.New("order not found")
	ErrOrderAlreadyPaid         = errors.New("order is already paid")
	ErrOrderCannotBeDiscounted  = errors.New("order cannot be discounted after payment")
	ErrDiscountNotAllowed       = errors.New("staff member cannot apply discounts")
	ErrDiscountExceedsLimit     = errors.New("discount exceeds staff limit")
	ErrPricedItemNotFound       = errors.New("order item is not an active menu item or combo")
	ErrPaymentInsufficient      = errors.New("amount paid is less than total")
	ErrOrderNotPaid             = errors.New("order has not been paid")
	ErrRefundExceedsTotal       = errors.New("refund exceeds order total")
	ErrLocationNotFound         = errors.New("location not found")
	ErrTableNotFound            = errors.New("table not found")
	ErrOrderCannotBeTransferred = errors.New("order cannot be transferred in current status")
	ErrOrderNotEditable         = errors.New("order cannot be edited in current status")
	ErrOrderCannotBeSplit       = errors.New("order cannot be split in current status")
	ErrItemNotVoidable          = errors.New("order item cannot be voided in current status")
	ErrInvalidOrderStatus       = errors.New("invalid order status")
	ErrInvalidItemStatus        = errors.New("invalid order item status")
	// ErrOrderStatusRequiresWorkflow is returned when a caller tries to set a
	// status that has companion writes and must go through its own flow
	// (paid -> payment, refunded -> refund, voided -> void).
	ErrOrderStatusRequiresWorkflow = errors.New("order status must be changed through its dedicated flow")
)

// ─── Location ────────────────────────────────────────────────────────────────

type Location struct {
	ID         int64
	OwnerID    int64
	BusinessID int64 `json:"business_id"`
	Name       string
	Address    string
	Phone      string
	IsActive   bool
	CreatedAt  time.Time
}

type LocationInput struct {
	Name    string
	Address string
	Phone   string
}

// ─── Table ───────────────────────────────────────────────────────────────────

type TableStatus string

const (
	TableStatusAvailable TableStatus = "available"
	TableStatusOccupied  TableStatus = "occupied"
	TableStatusReserved  TableStatus = "reserved"
	TableStatusCleaning  TableStatus = "cleaning"
)

type Table struct {
	ID                    int64
	OwnerID               int64
	BusinessID            int64 `json:"business_id"`
	LocationID            int64
	Name                  string
	Capacity              int
	Status                TableStatus
	CurrentSessionOrderID int64
	QRCodeURL             string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type TableInput struct {
	ID         int64
	LocationID int64
	Name       string
	Capacity   int
	QRCodeURL  string
}

// ─── Session ─────────────────────────────────────────────────────────────────

type SessionStatus string

const (
	SessionStatusOpen   SessionStatus = "open"
	SessionStatusClosed SessionStatus = "closed"
)

type Session struct {
	ID               int64
	OwnerID          int64
	BusinessID       int64 `json:"business_id"`
	LocationID       int64
	StaffID          int64
	Status           SessionStatus
	OpeningCash      float64
	ClosingCash      float64
	ExpectedCash     float64
	CashDifference   float64
	TotalSales       float64
	TotalOrders      int
	TotalCashSales   float64
	TotalCardSales   float64
	TotalOnlineSales float64
	TotalRefunds     float64
	OpenedAt         time.Time
	ClosedAt         *time.Time
	Notes            string
}

type OpenSessionInput struct {
	LocationID  int64
	OpeningCash float64
}

type CloseSessionInput struct {
	SessionID   int64
	ClosingCash float64
	Notes       string
}

// ─── Order ───────────────────────────────────────────────────────────────────

type OrderType string

const (
	OrderTypeDineIn   OrderType = "dine_in"
	OrderTypeTakeaway OrderType = "takeaway"
	OrderTypeDelivery OrderType = "delivery"
	OrderTypeOnline   OrderType = "online"
)

func IsOrderType(orderType OrderType) bool {
	switch orderType {
	case OrderTypeDineIn, OrderTypeTakeaway, OrderTypeDelivery, OrderTypeOnline:
		return true
	default:
		return false
	}
}

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusPreparing OrderStatus = "preparing"
	OrderStatusReady     OrderStatus = "ready"
	OrderStatusServed    OrderStatus = "served"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusRefunded  OrderStatus = "refunded"
	OrderStatusVoided    OrderStatus = "voided"
	OrderStatusHeld      OrderStatus = "held"
)

// IsOrderStatus reports whether status is one of the values the pos_orders
// status column accepts.
func IsOrderStatus(status OrderStatus) bool {
	switch status {
	case OrderStatusPending, OrderStatusConfirmed, OrderStatusPreparing,
		OrderStatusReady, OrderStatusServed, OrderStatusPaid,
		OrderStatusCancelled, OrderStatusRefunded, OrderStatusVoided,
		OrderStatusHeld:
		return true
	default:
		return false
	}
}

// CanSetOrderStatusDirectly reports whether the generic status endpoint may
// write status. paid, refunded, and voided all have companion writes
// (payment row and payment_status, refund row, void trail) and must be set
// through their dedicated flows.
func CanSetOrderStatusDirectly(status OrderStatus) bool {
	switch status {
	case OrderStatusPaid, OrderStatusRefunded, OrderStatusVoided:
		return false
	default:
		return IsOrderStatus(status)
	}
}

type PaymentMethod string

const (
	PaymentMethodCash   PaymentMethod = "cash"
	PaymentMethodCard   PaymentMethod = "card"
	PaymentMethodOnline PaymentMethod = "online"
	PaymentMethodSplit  PaymentMethod = "split"
)

type PaymentStatus string

const (
	PaymentStatusUnpaid   PaymentStatus = "unpaid"
	PaymentStatusPartial  PaymentStatus = "partial"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

type Order struct {
	ID              int64
	OwnerID         int64
	BusinessID      int64 `json:"business_id"`
	SessionID       int64
	TableID         int64
	LocationID      int64
	OrderNumber     string
	Type            OrderType
	Status          OrderStatus
	CustomerID      int64
	CustomerName    string
	StaffID         int64
	Subtotal        float64
	DiscountAmount  float64
	TaxAmount       float64
	Total           float64
	PaymentMethod   PaymentMethod
	PaymentStatus   PaymentStatus
	AmountPaid      float64
	ChangeDue       float64
	Notes           string
	KitchenNotes    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	PaidAt          *time.Time
	Items           []OrderItem
	TerminalID      int64
	BranchID        int64
	TipAmount       float64
	VoidReason      string
	VoidedBy        int64
	VoidedAt        *time.Time
	ExchangeOrderID int64
}

type OrderInput struct {
	SessionID     int64
	TableID       int64
	LocationID    int64
	Type          OrderType
	CustomerID    int64
	CustomerName  string
	PaymentMethod PaymentMethod
	Notes         string
	KitchenNotes  string
	Items         []OrderItemInput
}

// ─── Order Item ──────────────────────────────────────────────────────────────

type ItemStatus string

const (
	ItemStatusPending   ItemStatus = "pending"
	ItemStatusPreparing ItemStatus = "preparing"
	ItemStatusReady     ItemStatus = "ready"
	ItemStatusServed    ItemStatus = "served"
	ItemStatusCancelled ItemStatus = "cancelled"
)

// IsItemStatus reports whether status is one of the values the
// pos_order_items status column accepts.
func IsItemStatus(status ItemStatus) bool {
	switch status {
	case ItemStatusPending, ItemStatusPreparing, ItemStatusReady,
		ItemStatusServed, ItemStatusCancelled:
		return true
	default:
		return false
	}
}

type OrderItem struct {
	ID             int64
	OrderID        int64
	MenuItemID     int64
	Name           string
	SKU            string
	Quantity       float64
	UnitPrice      float64
	DiscountPct    float64
	DiscountAmount float64
	TaxRate        float64
	Total          float64
	Notes          string
	Status         ItemStatus
	CreatedAt      time.Time
}

type OrderItemInput struct {
	MenuItemID     int64
	ComboID        int64
	ModifierIDs    []int64
	Name           string
	SKU            string
	Quantity       float64
	UnitPrice      float64
	DiscountPct    float64
	DiscountAmount float64
	TaxRate        float64
	Notes          string
}

// ─── Payment ─────────────────────────────────────────────────────────────────

type PaymentInput struct {
	OrderID       int64
	PaymentMethod PaymentMethod
	AmountPaid    float64
}

// ─── Refund ──────────────────────────────────────────────────────────────────

type Refund struct {
	ID         int64
	OwnerID    int64
	BusinessID int64 `json:"business_id"`
	OrderID    int64
	Amount     float64
	Reason     string
	RefundedBy int64
	RefundedAt time.Time
}

type RefundInput struct {
	OrderID int64
	Amount  float64
	Reason  string
}

// ─── Discount ───────────────────────────────────────────────────────────────

type DiscountInput struct {
	OrderID       int64
	DiscountType  string
	DiscountValue float64
	Reason        string
}

// ─── Transfer Table ──────────────────────────────────────────────────────────
// ─── Dashboard ───────────────────────────────────────────────────────────────

type DashboardStats struct {
	TotalSales    float64 `json:"total_sales"`
	OrderCount    int     `json:"order_count"`
	AvgOrder      float64 `json:"avg_order"`
	TableTurnover float64 `json:"table_turnover"`
}
