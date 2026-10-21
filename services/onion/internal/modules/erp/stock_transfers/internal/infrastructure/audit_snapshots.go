package infrastructure

import (
	"time"

	"github.com/voxire/onion/internal/modules/erp/stock_transfers/internal/domain"
)

type stockTransferSnapshot struct {
	ID              int64                       `json:"id"`
	OwnerID         int64                       `json:"owner_id"`
	BusinessID      int64                       `json:"business_id"`
	FromWarehouseID int64                       `json:"from_warehouse_id"`
	ToWarehouseID   int64                       `json:"to_warehouse_id"`
	Status          domain.StockTransferStatus  `json:"status"`
	Notes           string                      `json:"notes"`
	Items           []stockTransferItemSnapshot `json:"items"`
	CreatedAt       time.Time                   `json:"created_at"`
	UpdatedAt       time.Time                   `json:"updated_at"`
}

type stockTransferItemSnapshot struct {
	ID         int64     `json:"id"`
	TransferID int64     `json:"transfer_id"`
	ProductID  int64     `json:"product_id"`
	Quantity   float64   `json:"quantity"`
	CreatedAt  time.Time `json:"created_at"`
}

type stockTransferInputSnapshot struct {
	FromWarehouseID int64                            `json:"from_warehouse_id"`
	ToWarehouseID   int64                            `json:"to_warehouse_id"`
	Notes           string                           `json:"notes"`
	Items           []stockTransferItemInputSnapshot `json:"items"`
}

type stockTransferItemInputSnapshot struct {
	ProductID int64   `json:"product_id"`
	Quantity  float64 `json:"quantity"`
}

func stockTransferSnapshotOf(value domain.StockTransfer) stockTransferSnapshot {
	result := stockTransferSnapshot{
		ID:              value.ID,
		OwnerID:         value.OwnerID,
		BusinessID:      value.BusinessID,
		FromWarehouseID: value.FromWarehouseID,
		ToWarehouseID:   value.ToWarehouseID,
		Status:          value.Status,
		Notes:           value.Notes,
		CreatedAt:       value.CreatedAt,
		UpdatedAt:       value.UpdatedAt,
	}
	if value.Items != nil {
		result.Items = make([]stockTransferItemSnapshot, len(value.Items))
		for i, item := range value.Items {
			result.Items[i] = stockTransferItemSnapshot(item)
		}
	}
	return result
}

func stockTransferItemSnapshotOf(value domain.StockTransferItem) stockTransferItemSnapshot {
	return stockTransferItemSnapshot(value)
}

func stockTransferInputSnapshotOf(value domain.StockTransferInput) stockTransferInputSnapshot {
	result := stockTransferInputSnapshot{
		FromWarehouseID: value.FromWarehouseID,
		ToWarehouseID:   value.ToWarehouseID,
		Notes:           value.Notes,
	}
	if value.Items != nil {
		result.Items = make([]stockTransferItemInputSnapshot, len(value.Items))
		for i, item := range value.Items {
			result.Items[i] = stockTransferItemInputSnapshot(item)
		}
	}
	return result
}

func stockTransferItemInputSnapshotOf(value domain.StockTransferItemInput) stockTransferItemInputSnapshot {
	return stockTransferItemInputSnapshot(value)
}
