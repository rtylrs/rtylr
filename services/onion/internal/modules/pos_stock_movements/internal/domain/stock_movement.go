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
	BusinessID    int64     `json:"business_id"`
	BranchID      int64     `json:"branch_id"`
	ProductID     int64     `json:"product_id"`
	Type          string    `json:"type"`
	Quantity      float64   `json:"quantity"`
	ReferenceID   *int64    `json:"reference_id"`
	ReferenceType *string   `json:"reference_type"`
	Note          *string   `json:"note"`
	EmployeeID    *int64    `json:"employee_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type BranchInventory struct {
	ID              int64     `json:"id"`
	BusinessID      int64     `json:"business_id"`
	BranchID        int64     `json:"branch_id"`
	ProductID       int64     `json:"product_id"`
	Quantity        float64   `json:"quantity"`
	ReorderPoint    *float64  `json:"reorder_point"`
	ReorderQuantity *float64  `json:"reorder_quantity"`
	UpdatedAt       time.Time `json:"updated_at"`
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
