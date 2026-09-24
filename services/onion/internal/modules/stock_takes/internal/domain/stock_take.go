package domain

import (
	"database/sql"
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

func (s *StockTake) FromRow(row *sql.Rows) error {
	var (
		warehouseID sql.NullInt64
		notes       sql.NullString
		startedAt   sql.NullTime
		completedAt sql.NullTime
	)

	err := row.Scan(
		&s.ID,
		&s.OwnerID,
		&s.BusinessID,
		&warehouseID,
		&s.Status,
		&notes,
		&startedAt,
		&completedAt,
		&s.CreatedAt,
	)
	if err != nil {
		return err
	}

	if warehouseID.Valid {
		s.WarehouseID = warehouseID.Int64
	}

	if notes.Valid {
		s.Notes = notes.String
	}

	if startedAt.Valid {
		s.StartedAt = startedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}

	if completedAt.Valid {
		s.CompletedAt = completedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}

	return nil
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

func (i *StockTakeItem) FromRow(row *sql.Rows) error {
	var createdAt time.Time

	err := row.Scan(
		&i.ID,
		&i.StockTakeID,
		&i.ProductID,
		&i.ExpectedQty,
		&i.CountedQty,
		&i.Variance,
		&createdAt,
	)
	if err != nil {
		return err
	}

	i.CreatedAt = createdAt.Format("2006-01-02T15:04:05Z07:00")

	return nil
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
