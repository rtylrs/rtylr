package domain

import "time"

type StoreOrder struct {
	ID              int64     `json:"id"`
	OwnerID         int64     `json:"ownerId"`
	CustomerEmail   string    `json:"customerEmail"`
	CustomerName    string    `json:"customerName"`
	CustomerPhone   string    `json:"customerPhone"`
	Status          string    `json:"status"`
	Subtotal        float64   `json:"subtotal"`
	Tax             float64   `json:"tax"`
	ShippingCost    float64   `json:"shippingCost"`
	Total           float64   `json:"total"`
	ShippingAddress string    `json:"shippingAddress"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type StoreOrderItem struct {
	ID        int64     `json:"id"`
	OrderID   int64     `json:"orderId"`
	ProductID int64     `json:"productId"`
	Name      string    `json:"name"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"createdAt"`
}

type StoreOrderInput struct {
	CustomerEmail   string           `json:"customerEmail"`
	CustomerName    string           `json:"customerName"`
	CustomerPhone   string           `json:"customerPhone"`
	Status          string           `json:"status"`
	Subtotal        float64          `json:"subtotal"`
	Tax             float64          `json:"tax"`
	ShippingCost    float64          `json:"shippingCost"`
	Total           float64          `json:"total"`
	ShippingAddress string           `json:"shippingAddress"`
	Notes           string           `json:"notes"`
	Items           []OrderItemInput `json:"items"`
}

type OrderItemInput struct {
	ProductID int64   `json:"productId"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Total     float64 `json:"total"`
}

type StoreOrderDetail struct {
	Order StoreOrder      `json:"order"`
	Items []StoreOrderItem `json:"items"`
}
