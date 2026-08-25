package stock_core

import (
	"errors"
	"testing"
)

func TestMovementBusinessIDDefaultsToOwnerID(t *testing.T) {
	if got := movementBusinessID(Movement{OwnerID: 42}); got != 42 {
		t.Fatalf("movementBusinessID() = %d, want owner id 42", got)
	}
}

func TestMovementBusinessIDUsesExplicitBusinessID(t *testing.T) {
	if got := movementBusinessID(Movement{OwnerID: 42, BusinessID: 99}); got != 99 {
		t.Fatalf("movementBusinessID() = %d, want explicit business id 99", got)
	}
}

func TestSetTotalTxRejectsNegativeTarget(t *testing.T) {
	_, _, err := SetTotalTx(nil, nil, Movement{LocationID: 1}, -1)
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("SetTotalTx() error = %v, want ErrInsufficientStock", err)
	}
}

func TestSetTotalTxRequiresLocation(t *testing.T) {
	_, _, err := SetTotalTx(nil, nil, Movement{OwnerID: 42}, 1)
	if err == nil || err.Error() != "stock_core: location is required for total adjustment" {
		t.Fatalf("SetTotalTx() error = %v, want missing-location error", err)
	}
}

func TestFIFOLayerConsumption(t *testing.T) {
	got := fifoLayerConsumption([]float64{2, 5, 10}, 6)
	want := []float64{2, 4, 0}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("consumption[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}
