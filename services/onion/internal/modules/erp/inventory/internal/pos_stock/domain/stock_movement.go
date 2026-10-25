package domain

import (
	"errors"
	"time"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock for operation")
	ErrInventoryNotFound = errors.New("inventory item not found")
)

type StockMovement struct {
	ID            int64
	BusinessID    int64
	BranchID      int64
	ProductID     int64
	Type          string
	Quantity      float64
	ReferenceID   *int64
	ReferenceType *string
	Note          *string
	EmployeeID    *int64
	CreatedAt     time.Time
}

type BranchInventory struct {
	ID                int64
	BusinessID        int64
	BranchID          int64
	ProductID         int64
	Name              string
	SKU               string
	Unit              string
	Quantity          float64
	LowStockThreshold *float64
	ReorderPoint      *float64
	ReorderQuantity   *float64
	UpdatedAt         time.Time
}

type ReceiveStockInput struct {
	BusinessID int64
	BranchID   int64
	ProductID  int64
	Quantity   float64
	Note       *string
	EmployeeID *int64
}

type AdjustStockInput struct {
	BusinessID int64
	BranchID   int64
	ProductID  int64
	Quantity   float64
	Note       *string
	EmployeeID *int64
}

type WasteStockInput struct {
	BusinessID int64
	BranchID   int64
	ProductID  int64
	Quantity   float64
	Note       *string
	EmployeeID *int64
}

type TransferStockInput struct {
	BusinessID   int64
	FromBranchID int64
	ToBranchID   int64
	ProductID    int64
	Quantity     float64
	Note         *string
	EmployeeID   *int64
}

type RecordedOrderInventory struct {
	BusinessID   int64
	SourceType   string
	SourceID     int64
	LocationKind string
	LocationID   int64
	Channel      string
	Return       bool
	Items        []OrderInventoryItem
}

type OrderInventoryItem struct {
	ProductID int64
	Quantity  float64
}
