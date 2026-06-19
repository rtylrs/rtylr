package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrStockTransferNotFound = errors.New("stock transfer not found")

type StockTransferStatus string

const (
	StockTransferStatusDraft     StockTransferStatus = "draft"
	StockTransferStatusInTransit StockTransferStatus = "in_transit"
	StockTransferStatusReceived  StockTransferStatus = "received"
	StockTransferStatusCancelled StockTransferStatus = "cancelled"
)

type StockTransfer struct {
	ID              int64               `json:"id"`
	OwnerID         int64               `json:"owner_id"`
	BusinessID      int64               `json:"business_id"`
	FromWarehouseID int64               `json:"from_warehouse_id"`
	ToWarehouseID   int64               `json:"to_warehouse_id"`
	Status          StockTransferStatus `json:"status"`
	Notes           string              `json:"notes"`
	Items           []StockTransferItem `json:"items"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

func (s *StockTransfer) FromRow(row *sql.Rows) error {
	var notes sql.NullString

	err := row.Scan(
		&s.ID,
		&s.OwnerID,
		&s.BusinessID,
		&s.FromWarehouseID,
		&s.ToWarehouseID,
		&s.Status,
		&notes,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if notes.Valid {
		s.Notes = notes.String
	}

	return nil
}

type StockTransferItem struct {
	ID         int64     `json:"id"`
	TransferID int64     `json:"transfer_id"`
	ProductID  int64     `json:"product_id"`
	Quantity   float64   `json:"quantity"`
	CreatedAt  time.Time `json:"created_at"`
}

func (i *StockTransferItem) FromRow(row *sql.Rows) error {
	return row.Scan(
		&i.ID,
		&i.TransferID,
		&i.ProductID,
		&i.Quantity,
		&i.CreatedAt,
	)
}

type StockTransferInput struct {
	FromWarehouseID int64                    `json:"from_warehouse_id"`
	ToWarehouseID   int64                    `json:"to_warehouse_id"`
	Notes           string                   `json:"notes"`
	Items           []StockTransferItemInput `json:"items"`
}

type StockTransferItemInput struct {
	ProductID int64   `json:"product_id"`
	Quantity  float64 `json:"quantity"`
}

type StockTransferStatusInput struct {
	ID     int64               `json:"id"`
	Status StockTransferStatus `json:"status"`
}
