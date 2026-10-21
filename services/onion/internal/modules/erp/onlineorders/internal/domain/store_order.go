package domain

import (
	"errors"
	"time"
)

// ErrInvalidAmount is returned when an order monetary field is negative.
var ErrInvalidAmount = errors.New("invalid amount: monetary values cannot be negative")

var ErrInvalidStatus = errors.New("invalid order status transition")

type StoreOrder struct {
	ID              int64
	OwnerID         int64
	BusinessID      int64
	CustomerEmail   string
	CustomerName    string
	CustomerPhone   string
	Status          string
	Subtotal        float64
	Tax             float64
	ShippingCost    float64
	Total           float64
	ShippingAddress string
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// StoreOrderPage is a single page of store orders plus the pagination
// metadata that was applied to produce it.
type StoreOrderPage struct {
	Orders []StoreOrder
	Total  int64
	Limit  int32
	Offset int32
}

type StoreOrderItem struct {
	ID        int64
	OrderID   int64
	ProductID int64
	Name      string
	Quantity  int
	Price     float64
	Total     float64
	CreatedAt time.Time
}

type StoreOrderInput struct {
	CustomerEmail   string
	CustomerName    string
	CustomerPhone   string
	Status          string
	Subtotal        float64
	Tax             float64
	ShippingCost    float64
	Total           float64
	ShippingAddress string
	Notes           string
	Items           []OrderItemInput
}

type OrderItemInput struct {
	ProductID int64
	Name      string
	Quantity  int
	Price     float64
	Total     float64
}

type StoreOrderDetail struct {
	Order StoreOrder
	Items []StoreOrderItem
}

// CentsToFloat converts integer cents to standard float amount.
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}
