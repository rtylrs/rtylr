package domain

import "time"

type KDSOrder struct {
	ID           int64     `json:"id"`
	OrderNumber  string    `json:"order_number"`
	Type         string    `json:"type"`
	Status       string    `json:"status"`
	TableID      int64     `json:"table_id"`
	CustomerName string    `json:"customer_name"`
	KitchenNotes string    `json:"kitchen_notes"`
	CreatedAt    time.Time `json:"created_at"`
	OwnerID      int64     `json:"owner_id"`
	BusinessID   int64     `json:"business_id"`
	Items        []KDSItem `json:"items"`
}

type KDSItem struct {
	ID       int64   `json:"id"`
	OrderID  int64   `json:"order_id"`
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Notes    string  `json:"notes"`
	Status   string  `json:"status"`
}

type UpdateItemStatusInput struct {
	OwnerID    int64
	BusinessID int64
	OrderID    int64
	ItemID     int64
	Status     string
}
