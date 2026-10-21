package infrastructure

import "github.com/voxire/onion/internal/modules/erp/stock_takes/internal/domain"

type stockTakeInputSnapshot struct {
	WarehouseID int64                        `json:"warehouse_id"`
	Notes       string                       `json:"notes"`
	Items       []stockTakeItemInputSnapshot `json:"items"`
}

type stockTakeItemInputSnapshot struct {
	ProductID   int64   `json:"product_id"`
	ExpectedQty float64 `json:"expected_qty"`
	CountedQty  float64 `json:"counted_qty"`
}

type stockTakeUpdateInputSnapshot struct {
	ID     int64                        `json:"id"`
	Status domain.StockTakeStatus       `json:"status"`
	Notes  string                       `json:"notes"`
	Items  []stockTakeItemInputSnapshot `json:"items"`
}

func stockTakeInputSnapshotOf(value domain.StockTakeInput) stockTakeInputSnapshot {
	result := stockTakeInputSnapshot{
		WarehouseID: value.WarehouseID,
		Notes:       value.Notes,
	}
	if value.Items != nil {
		result.Items = make([]stockTakeItemInputSnapshot, len(value.Items))
		for i, item := range value.Items {
			result.Items[i] = stockTakeItemInputSnapshot(item)
		}
	}
	return result
}

func stockTakeItemInputSnapshotOf(value domain.StockTakeItemInput) stockTakeItemInputSnapshot {
	return stockTakeItemInputSnapshot(value)
}

func stockTakeUpdateInputSnapshotOf(value domain.StockTakeUpdateInput) stockTakeUpdateInputSnapshot {
	result := stockTakeUpdateInputSnapshot{
		ID:     value.ID,
		Status: value.Status,
		Notes:  value.Notes,
	}
	if value.Items != nil {
		result.Items = make([]stockTakeItemInputSnapshot, len(value.Items))
		for i, item := range value.Items {
			result.Items[i] = stockTakeItemInputSnapshot(item)
		}
	}
	return result
}
