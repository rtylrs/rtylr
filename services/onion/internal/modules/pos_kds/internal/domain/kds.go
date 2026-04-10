package domain

import "time"

type KDSOrder struct {
	ID           int64      `json:"id"`
	OrderNumber  string     `json:"orderNumber"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	TableID      int64      `json:"tableId"`
	CustomerName string     `json:"customerName"`
	KitchenNotes string     `json:"kitchenNotes"`
	CreatedAt    time.Time  `json:"createdAt"`
	Items        []KDSItem  `json:"items"`
}

type KDSItem struct {
	ID       int64   `json:"id"`
	OrderID  int64   `json:"orderId"`
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Notes    string  `json:"notes"`
	Status   string  `json:"status"`
}

type UpdateItemStatusInput struct {
	OrderID int64
	ItemID  int64
	Status  string
}
