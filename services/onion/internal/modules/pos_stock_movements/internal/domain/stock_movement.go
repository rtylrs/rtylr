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
	ID            int64     `json:"id"`
	BusinessID    int64     `json:"businessId"`
	BranchID      int64     `json:"branchId"`
	ProductID     int64     `json:"productId"`
	Type          string    `json:"type"`
	Quantity      float64   `json:"quantity"`
	ReferenceID   *int64    `json:"referenceId"`
	ReferenceType *string   `json:"referenceType"`
	Note          *string   `json:"note"`
	EmployeeID    *int64    `json:"employeeId"`
	CreatedAt     time.Time `json:"createdAt"`
}

type BranchInventory struct {
	ID              int64     `json:"id"`
	BusinessID      int64     `json:"businessId"`
	BranchID        int64     `json:"branchId"`
	ProductID       int64     `json:"productId"`
	Quantity        float64   `json:"quantity"`
	ReorderPoint    *float64  `json:"reorderPoint"`
	ReorderQuantity *float64  `json:"reorderQuantity"`
	UpdatedAt       time.Time `json:"updatedAt"`
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
