package domain

import (
	"errors"
	"time"
)

var (
	ErrStockTakeNotFound          = errors.New("stock take not found")
	ErrStockTakeReferenceNotFound = errors.New("stock take reference not found in business")
	ErrStockTakeIncomplete        = errors.New("every stock-take item requires a physical count")
)

type StockTakeStatus string

const (
	StockTakeStatusDraft      StockTakeStatus = "draft"
	StockTakeStatusInProgress StockTakeStatus = "in_progress"
	StockTakeStatusCompleted  StockTakeStatus = "completed"
)

type StockTake struct {
	ID          int64           `json:"id"`
	OwnerID     int64           `json:"owner_id"`
	BusinessID  int64           `json:"business_id"`
	WarehouseID int64           `json:"warehouse_id"`
	Status      StockTakeStatus `json:"status"`
	Notes       string          `json:"notes"`
	Items       []StockTakeItem `json:"items"`
	StartedAt   string          `json:"started_at"`
	CompletedAt string          `json:"completed_at"`
	CreatedAt   time.Time       `json:"created_at"`
}
type StockTakeItem struct {
	ID          int64   `json:"id"`
	StockTakeID int64   `json:"stock_take_id"`
	ProductID   int64   `json:"product_id"`
	ExpectedQty float64 `json:"expected_qty"`
	CountedQty  float64 `json:"counted_qty"`
	Variance    float64 `json:"variance"`
	CreatedAt   string  `json:"created_at"`
}
type StockTakeInput struct {
	WarehouseID int64                `json:"warehouse_id"`
	Notes       string               `json:"notes"`
	Items       []StockTakeItemInput `json:"items"`
}

type StockTakeItemInput struct {
	ProductID   int64   `json:"product_id"`
	ExpectedQty float64 `json:"expected_qty"`
	CountedQty  float64 `json:"counted_qty"`
}

type StockTakeUpdateInput struct {
	ID     int64                `json:"id"`
	Status StockTakeStatus      `json:"status"`
	Notes  string               `json:"notes"`
	Items  []StockTakeItemInput `json:"items"`
}
