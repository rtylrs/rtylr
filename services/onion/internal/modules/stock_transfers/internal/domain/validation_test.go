package domain

import "testing"

func TestValidateStockTransferInput(t *testing.T) {
	valid := StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   2,
		Items:           []StockTransferItemInput{{ProductID: 3, Quantity: 1.5}},
	}
	if err := ValidateStockTransferInput(valid); err != nil {
		t.Fatalf("valid transfer rejected: %v", err)
	}

	tests := []StockTransferInput{
		{FromWarehouseID: 1, ToWarehouseID: 2, Items: []StockTransferItemInput{{ProductID: 3, Quantity: 0}}},
		{FromWarehouseID: 1, ToWarehouseID: 2, Items: []StockTransferItemInput{{ProductID: 3, Quantity: -1}}},
		{FromWarehouseID: 1, ToWarehouseID: 2, Items: []StockTransferItemInput{{ProductID: 0, Quantity: 1}}},
		{FromWarehouseID: 1, ToWarehouseID: 2},
	}
	for i, input := range tests {
		if err := ValidateStockTransferInput(input); err == nil {
			t.Errorf("invalid transfer %d was accepted", i)
		}
	}
}
