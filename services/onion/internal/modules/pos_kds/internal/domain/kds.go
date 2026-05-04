package domain

import "time"

type KDSOrder struct {
	ID           int64      `json:"id"`
	OrderNumber  string     `json:"order_number"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	TableID      int64      `json:"table_id"`
	CustomerName string     `json:"customer_name"`
	KitchenNotes string     `json:"kitchen_notes"`
	CreatedAt    time.Time  `json:"created_at"`
	Items        []KDSItem  `json:"items"`
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
	OwnerID int64
	OrderID int64
	ItemID  int64
	Status  string
}
