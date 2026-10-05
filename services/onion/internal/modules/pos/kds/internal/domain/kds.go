package domain

import "time"

type KDSOrder struct {
	ID           int64
	OrderNumber  string
	Type         string
	Status       string
	TableID      int64
	CustomerName string
	KitchenNotes string
	CreatedAt    time.Time
	OwnerID      int64
	BusinessID   int64
	Items        []KDSItem
}

type KDSItem struct {
	ID       int64
	OrderID  int64
	Name     string
	Quantity float64
	Notes    string
	Status   string
}

type UpdateItemStatusInput struct {
	OwnerID    int64
	BusinessID int64
	OrderID    int64
	ItemID     int64
	Status     string
}
