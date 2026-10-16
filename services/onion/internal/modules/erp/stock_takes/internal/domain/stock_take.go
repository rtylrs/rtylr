package domain

import (
	"errors"
	"time"
)

var (
	ErrStockTakeNotFound          = errors.New("stock take not found")
	ErrStockTakeReferenceNotFound = errors.New("stock take reference not found in business")
	ErrStockTakeIncomplete        = errors.New("every stock-take item requires a physical count")
	ErrStockTakeAlreadyCompleted  = errors.New("stock take is already completed and cannot be modified")
)

type StockTakeStatus string

const (
	StockTakeStatusDraft      StockTakeStatus = "draft"
	StockTakeStatusInProgress StockTakeStatus = "in_progress"
	StockTakeStatusCompleted  StockTakeStatus = "completed"
)

type StockTake struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	WarehouseID int64
	Status      StockTakeStatus
	Notes       string
	Items       []StockTakeItem
	StartedAt   string
	CompletedAt string
	CreatedAt   time.Time
}
type StockTakeItem struct {
	ID          int64
	StockTakeID int64
	ProductID   int64
	ExpectedQty float64
	CountedQty  float64
	Variance    float64
	CreatedAt   string
}
type StockTakeInput struct {
	WarehouseID int64
	Notes       string
	Items       []StockTakeItemInput
}

type StockTakeItemInput struct {
	ProductID   int64
	ExpectedQty float64
	CountedQty  float64
}

type StockTakeUpdateInput struct {
	ID     int64
	Status StockTakeStatus
	Notes  string
	Items  []StockTakeItemInput
}
