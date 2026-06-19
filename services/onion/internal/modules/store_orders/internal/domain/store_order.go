package domain

import (
	"errors"
	"time"
)

// ErrInvalidAmount is returned when an order monetary field is negative.
var ErrInvalidAmount = errors.New("invalid amount: monetary values cannot be negative")

type StoreOrder struct {
	ID              int64     `json:"id"`
	OwnerID         int64     `json:"owner_id"`
	BusinessID      int64     `json:"business_id"`
	CustomerEmail   string    `json:"customer_email"`
	CustomerName    string    `json:"customer_name"`
	CustomerPhone   string    `json:"customer_phone"`
	Status          string    `json:"status"`
	Subtotal        float64   `json:"subtotal"`
	Tax             float64   `json:"tax"`
	ShippingCost    float64   `json:"shipping_cost"`
	Total           float64   `json:"total"`
	ShippingAddress string    `json:"shipping_address"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type StoreOrderItem struct {
	ID        int64     `json:"id"`
	OrderID   int64     `json:"order_id"`
	ProductID int64     `json:"product_id"`
	Name      string    `json:"name"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}

type StoreOrderInput struct {
	CustomerEmail   string           `json:"customer_email"`
	CustomerName    string           `json:"customer_name"`
	CustomerPhone   string           `json:"customer_phone"`
	Status          string           `json:"status"`
	Subtotal        float64          `json:"subtotal"`
	Tax             float64          `json:"tax"`
	ShippingCost    float64          `json:"shipping_cost"`
	Total           float64          `json:"total"`
	ShippingAddress string           `json:"shipping_address"`
	Notes           string           `json:"notes"`
	Items           []OrderItemInput `json:"items"`
}

type OrderItemInput struct {
	ProductID int64   `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Total     float64 `json:"total"`
}

type StoreOrderDetail struct {
	Order StoreOrder       `json:"order"`
	Items []StoreOrderItem `json:"items"`
}
