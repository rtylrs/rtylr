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

func TestValidateStockTransferInput_SameWarehouse(t *testing.T) {
	input := StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   1,
		Items:           []StockTransferItemInput{{ProductID: 3, Quantity: 1.5}},
	}
	if err := ValidateStockTransferInput(input); err == nil {
		t.Fatal("expected error when from and to warehouse are identical")
	}
}

func TestValidateStockTransferStatusTransition(t *testing.T) {
	cases := []struct {
		current StockTransferStatus
		next    StockTransferStatus
		wantErr bool
	}{
		{StockTransferStatusDraft, StockTransferStatusDraft, false},
		{StockTransferStatusDraft, StockTransferStatusInTransit, false},
		{StockTransferStatusDraft, StockTransferStatusReceived, false},
		{StockTransferStatusDraft, StockTransferStatusCancelled, false},
		{StockTransferStatusInTransit, StockTransferStatusReceived, false},
		{StockTransferStatusInTransit, StockTransferStatusCancelled, false},
		{StockTransferStatusInTransit, StockTransferStatusDraft, true},
		{StockTransferStatusReceived, StockTransferStatusCancelled, true},
		{StockTransferStatusReceived, StockTransferStatusInTransit, true},
		{StockTransferStatusCancelled, StockTransferStatusDraft, true},
		{StockTransferStatusCancelled, StockTransferStatusReceived, true},
	}

	for _, tc := range cases {
		err := ValidateStockTransferStatusTransition(tc.current, tc.next)
		if (err != nil) != tc.wantErr {
			t.Errorf("transition %s -> %s: got err=%v, wantErr=%v", tc.current, tc.next, err, tc.wantErr)
		}
	}
}
