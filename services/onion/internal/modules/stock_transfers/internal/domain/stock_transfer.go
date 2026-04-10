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
	OwnerID         int64               `json:"ownerId"`
	FromWarehouseID int64               `json:"fromWarehouseId"`
	ToWarehouseID   int64               `json:"toWarehouseId"`
	Status          StockTransferStatus `json:"status"`
	Notes           string              `json:"notes"`
	Items           []StockTransferItem `json:"items"`
	CreatedAt       time.Time           `json:"createdAt"`
	UpdatedAt       time.Time           `json:"updatedAt"`
}

func (s *StockTransfer) FromRow(row *sql.Rows) error {
	var notes sql.NullString

	err := row.Scan(
		&s.ID,
		&s.OwnerID,
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
	TransferID int64     `json:"transferId"`
	ProductID  int64     `json:"productId"`
	Quantity   float64   `json:"quantity"`
	CreatedAt  time.Time `json:"createdAt"`
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
	FromWarehouseID int64                    `json:"fromWarehouseId"`
	ToWarehouseID   int64                    `json:"toWarehouseId"`
	Notes           string                   `json:"notes"`
	Items           []StockTransferItemInput `json:"items"`
}

type StockTransferItemInput struct {
	ProductID int64   `json:"productId"`
	Quantity  float64 `json:"quantity"`
}

type StockTransferStatusInput struct {
	ID     int64               `json:"id"`
	Status StockTransferStatus `json:"status"`
}
